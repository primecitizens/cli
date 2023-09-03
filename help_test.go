// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"strings"
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

var (
	_ CompAction     = (*FlagHelp)(nil)
	_ HelperTerminal = (*FlagHelp)(nil)
	_ HelperTerminal = (*CmdHelp)(nil)
)

func TestHelper(t *testing.T) {
	root := &Cmd{
		Pattern:    "test [foo] {a|b}",
		BriefUsage: "This is just a test command",
		Extra:      nil,

		FlagRule: MergeFlagRules(OneOf("i32", "i64"), AllOf("i64")),
		Flags: NewMapIndexer().Add(
			&Int32{
				BriefUsage: "set a 32-bit integer",
				Ext: &FlagHelp{
					Deprecation:  "Use i64 instead",
					Experimental: "it bites",
					Changelog:    "deprecated since i64 added",
				},
			}, "i32",
		).Add(
			&Int64V{
				State_:     FlagStateHidden,
				BriefUsage: "set a 64-bit integer",
			}, "i64", "i",
		),
		Children: []*Cmd{
			{Pattern: "foo"},
			{Pattern: "bar"},
		},
	}

	err := HandleArgErrorAsHelpRequest(nil, nil, nil, 0, nil)
	assert.NoError(t, err)

	var sb strings.Builder
	err = HandleHelpRequest(&CmdOptions{Stderr: &sb}, Route{root}, nil, -1)
	expected := "" +
		"test [foo] {a|b}\n" +
		"\n" +
		"This is just a test command\n" +
		"\n" +
		"Sub-Commands:\n" +
		"- foo\n" +
		"- bar\n" +
		"\n" +
		"Flags:\n" +
		"     --i32 int  (oneof[--i32, --i64]) set a 32-bit integer\n" +
		"                DEPRECATED: Use i64 instead\n" +
		"                EXPERIMENTAL: it bites\n" +
		"  -i --i64 int  (hidden, required) set a 64-bit integer\n"
	assert.NoError(t, err)
	assert.Eq(t, expected, sb.String())

	root.Extra = &CmdHelp{
		Deprecation: "use x instead",
		Example:     "test foo a",
		Changelog:   "Added since day one",
	}

	sb.Reset()
	err = HandleArgErrorAsHelpRequest(
		&CmdOptions{Stderr: &sb}, Route{root}, nil, 0, ErrTimeout{},
	)

	expected = "" +
		"Error: timeout\n" +
		"\n" +
		"test [foo] {a|b}\n" +
		"\n" +
		"This is just a test command\n" +
		"\n" +
		"Sub-Commands:\n" +
		"- foo\n" +
		"- bar\n" +
		"\n" +
		"DEPRECATED: use x instead\n" +
		"\n" +
		"Example:\n" +
		"\n" +
		"test foo a\n" +
		"\n" +
		"Flags:\n" +
		"     --i32 int  (oneof[--i32, --i64]) set a 32-bit integer\n" +
		"                DEPRECATED: Use i64 instead\n" +
		"                EXPERIMENTAL: it bites\n" +
		"  -i --i64 int  (hidden, required) set a 64-bit integer\n" +
		"\n" +
		"Changes:\n" +
		"\n" +
		"Added since day one\n" +
		"\n"

	assert.Eq(t, expected, sb.String())
	assert.Error(t, err)
}
