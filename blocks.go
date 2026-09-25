// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package forma

import (
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"

	"golang.design/x/accel"

	"latere.ai/x/forma/internal/prefix"
	"latere.ai/x/forma/model"
)

// CacheBlock is how many positions one physical block holds.
//
// specs/016-prefix-cache.md §3: sharing is block-aligned, so a genuine common
// prefix loses at most CacheBlock-1 tokens to rounding. Against a prefix of
// hundreds that is a rounding error, and the alternative — two sequences
// writing one block at different offsets — is a correctness problem rather than
// an optimization.
//
// It is a constant rather than an option because it is a plan parameter: accel
// folds the block size into the plan's attributes, so two block sizes are two
// compiled graphs of every shape, and a knob nobody can evaluate would buy a
// second plan cache.
const CacheBlock = 32

// blockPool is the device memory a process-scoped prefix cache shares, and the
// bookkeeping over it.
//
// One key state and one value state for the whole model rather than one pair
// per session, which is the change that makes sharing possible and is also the
// change that bounds a server's memory: a process with this pool costs the pool
// whatever its concurrency is, where a process with per-session caches costs
// sessions times context.
type blockPool struct {
	pool         *prefix.Pool
	keys, values *accel.Buffer

	// dtype is what the states hold: **f16**.
	//
	// [C5](specs/010-conformance.md) is the argument. The key and value states
	// are the largest allocation a serving process has after the weights and
	// the only one that scales with *both* concurrency and context, so halving
	// them is twice the blocks, twice the prefixes worth keeping, and twice the
	// batch size worth reaching -- [008 §1](specs/008-scheduler.md) makes the
	// throughput ceiling proportional to 1/A.
	//
	// It is defensible where a narrow *accumulator* would not be. K and V are
	// operands: the score accumulates in f32 whatever they are stored as, which
	// is the trade MatMul already makes.
	//
	// It is a field rather than a constant because it was f32 for six hours.
	// [C24](specs/010-conformance.md) was accel selecting the f16 prefill
	// kernel and then overwriting the selection whenever a page table was
	// supplied -- and a pool is paged by construction, so the narrow cache was
	// reachable for a contiguous single sequence and for nobody who shares
	// blocks. Filed as accel#25 and fixed upstream the same day.
	dtype accel.DType

	// positions is the pool's capacity, blocks*CacheBlock, and is the row
	// count of both states.
	positions int

	// domain and minted make the salt of a session opened without
	// [WithCacheSalt]. See [blockPool.mintSalt].
	domain string
	minted atomic.Uint64
}

// mintSalt is the salt of a session opened without [WithCacheSalt]: one no
// other session holds, so the session shares no block with any other and still
// finds its own blocks on its next turn.
//
// The empty salt passed through would be one domain. Under [CacheProcess] the
// chain seed's scope domain is empty (internal/prefix/prefix.go), so every
// unsalted session would hash into it and hit the blocks every other unsalted
// session published. A hit is faster than a miss, so that is a membership test
// over another conversation's prompt (016 §7), and 019-D3 says an unkeyed
// request shares with nobody.
//
// The salt is [Pool.salt]'s and [Runner.salt]'s, for their reasons: random
// bytes, so no caller can name the domain and share into it, and a counter, so
// no two sessions hold one. It is minted per session rather than per request
// because a session is one conversation, and reusing its own earlier turns is
// what the cache is for. Sessions that should share a prefix say so with one
// [WithCacheSalt].
func (bp *blockPool) mintSalt() string {
	return bp.domain + "-" + strconv.FormatUint(bp.minted.Add(1), 36)
}

// newBlockPool allocates the shared states and the bookkeeping over them.
//
// positions is rounded *down* to whole blocks and refused below one. Rounding
// up would allocate memory the caller did not ask for, and 005-D3's rule is
// that the number is reported before the allocation rather than after the
// failure.
func newBlockPool(dev *accel.Device, c *model.Config, scope prefix.Scope,
	positions int) (*blockPool, error) {

	blocks := positions / CacheBlock
	if blocks < 1 {
		return nil, fmt.Errorf("forma: a shared prefix cache of %d positions holds no "+
			"whole block of %d; it needs at least one", positions, CacheBlock)
	}
	p, err := prefix.New(prefix.Config{
		Block: CacheBlock, Blocks: blocks, Scope: scope,
	})
	if err != nil {
		return nil, fmt.Errorf("forma: %w", err)
	}
	domain, err := mintDomain()
	if err != nil {
		return nil, err
	}
	bp := &blockPool{pool: p, positions: blocks * CacheBlock, dtype: accel.F16,
		domain: domain}

	// The layers that have a key/value cache, and a gated-delta layer has
	// none. Allocating over the whole stack would reserve four times the pool
	// for a hybrid, three quarters of which nothing ever writes
	// ([023-D3](specs/023-cache-kinds.md)).
	n := cachedLayers(c) * bp.positions * c.NumKVHeads * c.HeadDim
	for _, a := range []struct {
		dst   **accel.Buffer
		label string
	}{{&bp.keys, model.PortKeys}, {&bp.values, model.PortValues}} {
		b, err := dev.NewBuffer(accel.BufferDescriptor{
			DType: bp.dtype, Count: n, Label: a.label,
			Usage: accel.BufferStorage | accel.BufferCopyDst | accel.BufferCopySrc,
		})
		if err != nil {
			_ = bp.close()
			// Named in the operator's own terms, because the number that
			// produced it is theirs: the pool is sessions x context positions,
			// and a device that cannot hold it is answered by lowering one of
			// the two rather than by reading an allocator's error.
			return nil, fmt.Errorf("forma: the shared prefix cache needs %s for its %s "+
				"state at %d positions, and the device refused it: %w; lower "+
				"--sessions or --context, which are the two numbers that produced "+
				"the pool", bytesText(int64(n)*int64(bp.dtype.Size())), a.label,
				bp.positions, err)
		}
		*a.dst = b
	}
	// Zeroed and not merely allocated, for the reason a session's own cache is:
	// a length of zero means no row is read, and a NaN left by a previous
	// tenant would still reach attention through a prefill's padded rows.
	//
	// Written as the dtype the buffer holds. accel checks the host slice's
	// element type against the buffer's, so the two move together and a plane
	// of the wrong width is a refusal rather than a silent halving.
	var zero any = make([]float32, n)
	if bp.dtype == accel.F16 {
		zero = make([]uint16, n)
	}
	for _, b := range []*accel.Buffer{bp.keys, bp.values} {
		if err := dev.Queue().WriteBuffer(b, 0, zero); err != nil {
			_ = bp.close()
			return nil, fmt.Errorf("forma: clearing the shared cache: %w", err)
		}
	}
	if err := dev.Queue().Flush().Wait(); err != nil {
		_ = bp.close()
		return nil, fmt.Errorf("forma: clearing the shared cache: %w", err)
	}
	return bp, nil
}

// maxPages is the width of the page-table port: a sequence may in principle
// hold every block in the pool.
func (b *blockPool) maxPages() int { return b.positions / CacheBlock }

func (b *blockPool) close() error {
	var errs []error
	for _, buf := range []*accel.Buffer{b.keys, b.values} {
		if buf != nil {
			errs = append(errs, buf.Close())
		}
	}
	return errors.Join(errs...)
}
