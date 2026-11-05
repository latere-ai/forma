// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// foreignJSON reads JSON this package does not produce: request bodies sent by third-party clients.
// Duplicate member names (the last one wins), invalid UTF-8 and any letter case
// in member names are accepted, as the readers of these documents always have.
var foreignJSON = json.JoinOptions(
	jsontext.AllowDuplicateNames(true),
	jsontext.AllowInvalidUTF8(true),
	json.MatchCaseInsensitiveNames(true),
)

// modelText is set on every response body that can carry model output or
// request text, which are byte strings and need not be valid UTF-8. The
// invalid bytes are written as U+FFFD.
var modelText = jsontext.AllowInvalidUTF8(true)
