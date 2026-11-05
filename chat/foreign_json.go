// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package chat

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// foreignJSON checks tool arguments and schemas, which models and callers
// write. Duplicate member names and invalid UTF-8 are accepted, as they always
// have been, because the bytes are copied into the prompt unchanged.
var foreignJSON = json.JoinOptions(
	jsontext.AllowDuplicateNames(true),
	jsontext.AllowInvalidUTF8(true),
)
