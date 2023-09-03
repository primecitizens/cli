// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

func TestErrors(t *testing.T) {
	for _, test := range []struct {
		err error
		msg string
	}{
		{&ErrAmbiguousArgs{Name: "foo", Value: "-1"},
			"ambiguous arg combination `--foo -1`: implicit flag followed by potential flag"},
		{&ErrAmbiguousArgs{Name: "f", Value: "-1"},
			"ambiguous arg combination `-f -1`: implicit flag followed by potential flag"},
		{&ErrShorthandOfExplicitFlagInMiddle{Shorthand: "f", ShorthandCluster: "cfd", Value: "pri"},
			"non-implicit flag -f cannot use value specified with `=` in middle of shorthands (-cfd=pri)"},
		{&ErrDuplicateFlag{Name: "foo"},
			"duplicate flag --foo"},
		{&ErrDuplicateFlag{Name: "f"},
			"duplicate flag -f"},
		{&ErrFlagUndefined{Name: "foo"},
			"undefined flag --foo (index: 0)"},
		{&ErrFlagUndefined{Name: "f", At: -1},
			"undefined flag -f"},
		{&ErrFlagValueMissing{Name: "foo", At: 1},
			"missing value for flag --foo (index: 1)"},
		{&ErrFlagValueMissing{Name: "f", At: 1},
			"missing value for flag -f (index: 1)"},
		{&ErrCmdNotRunnable{Name: "foo"},
			"command foo is not runnable (not having function Run)"},
		{&ErrHelpPending{HelpArg: "foo", At: 1},
			"help requested by arg `foo` (index: 1) but not handled"},
		{&ErrHelpHandled{},
			"help request handled"},
	} {
		assert.Eq(t, test.msg, test.err.Error())
	}
}
