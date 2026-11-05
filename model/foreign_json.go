// SPDX-FileCopyrightText: 2026 Latere AI
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

// foreignJSON reads JSON this package does not produce: a checkpoint's config.json.
// Duplicate member names (the last one wins), invalid UTF-8 and any letter case
// in member names are accepted, as the readers of these documents always have.
var foreignJSON = json.JoinOptions(
	jsontext.AllowDuplicateNames(true),
	jsontext.AllowInvalidUTF8(true),
	json.MatchCaseInsensitiveNames(true),
)
