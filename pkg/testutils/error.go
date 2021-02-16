// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package testutils

import (
	"regexp"

	"github.com/auxten/postgresql-parser/pkg/sql/pgwire/pgerror"
)

// IsError returns true if the error string matches the supplied regex.
// An empty regex is interpreted to mean that a nil error is expected.
func IsError(err error, re string) bool {
	if err == nil && re == "" {
		return true
	}
	if err == nil || re == "" {
		return false
	}
	errString := pgerror.FullError(err)
	matched, merr := regexp.MatchString(re, errString)
	if merr != nil {
		return false
	}
	return matched
}

