// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

var (
	_ CompAction = (*CompActionStatic)(nil)
	_ CompAction = CompActionFunc(nil)
	_ CompAction = CompActionDirs{}
	_ CompAction = CompActionFiles{}
	_ CompAction = CompActionDisable{}
)

func TestCompTask_AddDefault(t *testing.T) {
	var cc CompCmdShells
	flag := &StringV{
		Ext: &FlagHelp{
			Completion: &CompActionStatic{
				Suggestions: []CompItem{
					{Value: "foo", Kind: CompKindFlagValue},
					{Value: "bar", Kind: CompKindFlagValue},
					{Value: "far", Kind: CompKindFlagValue},
				},
				Want:  CompStateHasFlagValues,
				State: testCompStateFlagCompletionAdded,
			},
		},
	}

	root := &Cmd{
		Flags: NewMapIndexer().Add(flag, "string", "s"),
		Completion: &CompActionStatic{
			Suggestions: []CompItem{
				{Value: "aha"},
				{Value: "blah"},
				{Value: "brb"},
				{Value: "-hmm"},
				{Value: "--wat"},
			},
			Want:  CompStateHasSubcmds,
			State: testCompStateCmdCompletionAdded,
		},
		Children: []*Cmd{
			cc.Setup("completion", 0, false),
		},
	}

	for _, test := range []struct {
		name     string
		at       int
		args     []string
		expected *CompTask
	}{
		{"empty arg", 1, []string{"./test", ""}, &CompTask{
			debug: nil,
			result: []CompItem{
				// default call Cmd.Completion
				{Value: "aha", Kind: CompKindText},
				{Value: "blah", Kind: CompKindText},
				{Value: "brb", Kind: CompKindText},
				{Value: "-hmm", Kind: CompKindText},
				{Value: "--wat", Kind: CompKindText},
				// default add flag names
				{Value: "string", Kind: CompKindFlagName},
				{Value: "s", Kind: CompKindFlagName},
				// default add subcmds
				{Value: "completion", Description: "shell completion", Kind: CompKindText},
			},
			ExecutablePath:   "./test",
			Args:             []string{""},
			At:               0,
			ToComplete:       "",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasFlagNames | CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasFlagNames | CompStateHasSubcmds,
		}},
		{"invalid 'at' as empty arg", -1, []string{"./test"}, &CompTask{
			debug: nil,
			result: []CompItem{
				// default call Cmd.Completion
				{Value: "aha", Kind: CompKindText},
				{Value: "blah", Kind: CompKindText},
				{Value: "brb", Kind: CompKindText},
				{Value: "-hmm", Kind: CompKindText},
				{Value: "--wat", Kind: CompKindText},
				// default add flag names
				{Value: "string", Kind: CompKindFlagName},
				{Value: "s", Kind: CompKindFlagName},
				// default add subcmds
				{Value: "completion", Description: "shell completion", Kind: CompKindText},
			},
			ExecutablePath:   "./test",
			Args:             []string{},
			At:               -1,
			ToComplete:       "",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasFlagNames | CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasFlagNames | CompStateHasSubcmds,
		}},
		{"empty 'args' ok", 0, []string{}, &CompTask{
			debug: nil,
			result: []CompItem{
				// default call Cmd.Completion
				{Value: "aha", Kind: CompKindText},
				{Value: "blah", Kind: CompKindText},
				{Value: "brb", Kind: CompKindText},
				{Value: "-hmm", Kind: CompKindText},
				{Value: "--wat", Kind: CompKindText},
				// default add flag names
				{Value: "string", Kind: CompKindFlagName},
				{Value: "s", Kind: CompKindFlagName},
				// default add subcmds
				{Value: "completion", Description: "shell completion", Kind: CompKindText},
			},
			ExecutablePath:   "",
			Args:             []string{},
			At:               0,
			ToComplete:       "",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasFlagNames | CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasFlagNames | CompStateHasSubcmds,
		}},
		{"prefix c", 1, []string{"./test", "c"}, &CompTask{
			debug: nil,
			result: []CompItem{
				// default add subcmds
				{Value: "completion", Description: "shell completion", Kind: CompKindText},
			},
			ExecutablePath:   "./test",
			Args:             []string{"c"},
			At:               0,
			ToComplete:       "c",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasSubcmds,
		}},
		{"single hyphen", 1, []string{"./test", "-"}, &CompTask{
			debug: nil,
			result: []CompItem{
				// default add subcmds
				{Value: "-hmm", Kind: CompKindText},
				{Value: "--wat", Kind: CompKindText},
				// default add flag names
				{Value: "string", Kind: CompKindFlagName},
				{Value: "s", Kind: CompKindFlagName},
			},
			ExecutablePath:   "./test",
			Args:             []string{"-"},
			At:               0,
			ToComplete:       "-",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasFlagNames | CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasFlagNames | CompStateHasSubcmds,
		}},
		{"single dash", 1, []string{"./test", "--"}, &CompTask{
			debug: nil,
			result: []CompItem{
				// default add subcmds
				{Value: "--wat", Kind: CompKindText},
				// default add flag names
				{Value: "string", Kind: CompKindFlagName},
			},
			ExecutablePath:   "./test",
			Args:             []string{"--"},
			At:               0,
			ToComplete:       "--",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasFlagNames | CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasFlagNames | CompStateHasSubcmds,
		}},
		{"single dash and s", 1, []string{"./test", "--s"}, &CompTask{
			debug: nil,
			result: []CompItem{
				// default add flag names
				{Value: "string", Kind: CompKindFlagName},
			},
			ExecutablePath:   "./test",
			Args:             []string{"--s"},
			At:               0,
			ToComplete:       "--s",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasFlagNames | CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasFlagNames | CompStateHasSubcmds,
		}},
		{"separate empty flag value", 2, []string{"./test", "--string", ""}, &CompTask{
			debug: nil,
			result: []CompItem{
				{Value: "foo", Kind: CompKindFlagValue},
				{Value: "bar", Kind: CompKindFlagValue},
				{Value: "far", Kind: CompKindFlagValue},
			},
			ExecutablePath:   "./test",
			Args:             []string{"--string", ""},
			At:               1,
			ToComplete:       "",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: flag,
			state:            CompStateHasFlagValues | testCompStateFlagCompletionAdded,
			want:             CompStateHasFlagValues,
		}},
		{"long name value pair", 1, []string{"./test", "--string=f"}, &CompTask{
			debug: nil,
			result: []CompItem{
				{Value: "foo", Kind: CompKindFlagValue},
				{Value: "far", Kind: CompKindFlagValue},
			},
			ExecutablePath:   "./test",
			Args:             []string{"--string=f"},
			At:               0,
			ToComplete:       "f",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: flag,
			state:            CompStateHasFlagValues | testCompStateFlagCompletionAdded,
			want:             CompStateHasFlagValues,
		}},
		{"long name value pair not flag", 1, []string{"./test", "--foo=f"}, &CompTask{
			debug:            nil,
			result:           []CompItem{},
			ExecutablePath:   "./test",
			Args:             []string{"--foo=f"},
			At:               0,
			ToComplete:       "--foo=f",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasSubcmds,
		}},
		{"flag shorthand cluster value pair", 1, []string{"./test", "-fafojvioernoaifjos=b"}, &CompTask{
			debug:            nil,
			result:           []CompItem{{Value: "bar", Kind: CompKindFlagValue}},
			ExecutablePath:   "./test",
			Args:             []string{"-fafojvioernoaifjos=b"},
			At:               0,
			ToComplete:       "b",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: flag,
			state:            CompStateHasFlagValues | testCompStateFlagCompletionAdded,
			want:             CompStateHasFlagValues,
		}},
		{"flag shorthand value pair not flag", 1, []string{"./test", "-h=b"}, &CompTask{
			debug:            nil,
			result:           []CompItem{},
			ExecutablePath:   "./test",
			Args:             []string{"-h=b"},
			At:               0,
			ToComplete:       "-h=b",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasSubcmds,
		}},
		{"like name value pair but invalid flag form", 1, []string{"./test", "--f="}, &CompTask{
			debug:            nil,
			result:           []CompItem{},
			ExecutablePath:   "./test",
			Args:             []string{"--f="},
			At:               0,
			ToComplete:       "--f=",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasSubcmds,
		}},
		{"hyphen prefix but not a flag", 1, []string{"./test", "-h"}, &CompTask{
			debug: nil,
			result: []CompItem{
				{Value: "-hmm", Kind: CompKindText},
			},
			ExecutablePath:   "./test",
			Args:             []string{"-h"},
			At:               0,
			ToComplete:       "-h",
			Route:            []*Cmd{root},
			PosArgs:          nil,
			DashArgs:         nil,
			FlagMissingValue: nil,
			state:            CompStateHasFlagNames | CompStateHasSubcmds | testCompStateCmdCompletionAdded,
			want:             CompStateHasFlagNames | CompStateHasSubcmds,
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var (
				tsk  CompTask
				opts CmdOptions
			)
			tsk.Init(root, &opts, test.at, test.args...)

			assert.Eq(t, len(test.expected.result), tsk.AddDefault())
			assert.Eq(t, test.expected.want, tsk.want)
			assert.Eq(t, test.expected.state, tsk.state)
			assert.Eq(t, test.expected.FlagMissingValue, tsk.FlagMissingValue)
			assert.EqS(t, test.expected.DashArgs, tsk.DashArgs)
			assert.EqS(t, test.expected.PosArgs, tsk.PosArgs)
			assert.EqS(t, test.expected.Route, tsk.Route)
			assert.Eq(t, test.expected.ToComplete, tsk.ToComplete)
			assert.Eq(t, test.expected.At, tsk.At)
			assert.EqS(t, test.expected.Args, tsk.Args)
			assert.Eq(t, test.expected.ExecutablePath, tsk.ExecutablePath)
			assert.EqS(t, test.expected.result, tsk.result)
		})
	}
}

func TestCompTask_AddSubcmds(t *testing.T) {
	const descr = "some description"
	root := &Cmd{
		Pattern: "root",
		Children: []*Cmd{
			{Pattern: "foo"},
			{Pattern: "bar", BriefUsage: descr},
			{Pattern: "far"},
			{Pattern: "fff", State: CmdStateHidden},
		},
	}

	for _, test := range []struct {
		toComplete string
		expected   []CompItem
	}{
		{"", []CompItem{
			{Value: "foo"},
			{Value: "bar", Description: descr},
			{Value: "far"},
		}},
		{"x", nil},
		{"f", []CompItem{
			{Value: "foo"},
			{Value: "far"},
		}},
		{"fo", []CompItem{
			{Value: "foo"},
		}},
		{"foo", []CompItem{
			{Value: "foo"},
		}},
	} {
		t.Run(test.toComplete, func(t *testing.T) {
			tsk := CompTask{
				ToComplete: test.toComplete,
			}

			assert.Eq(t, len(test.expected), tsk.AddSubcmds(false, root, true))
			assert.EqS(t, test.expected, tsk.result)

			assert.Eq(t, 0, tsk.AddSubcmds(false, root, true))
			assert.EqS(t, test.expected, tsk.result)
		})
	}
}

func TestCompTask_AddFlagNames(t *testing.T) {
	const descr = "some description"
	flags := NewMapIndexer().
		Add(&FlagEmptyV{}, "foo", "f").
		Add(&FlagEmptyV{BriefUsage: descr}, "bar", "b").
		Add(&FlagEmptyV{}, "far").
		Add(&FlagEmptyV{State_: FlagStateHidden}, "fff")

	for _, test := range []struct {
		toComplete string
		expected   []CompItem
	}{
		{"", []CompItem{
			{Value: "foo", Kind: CompKindFlagName},
			{Value: "f", Kind: CompKindFlagName},
			{Value: "bar", Description: descr, Kind: CompKindFlagName},
			{Value: "b", Description: descr, Kind: CompKindFlagName},
			{Value: "far", Kind: CompKindFlagName},
		}},
		{"-", []CompItem{
			{Value: "foo", Kind: CompKindFlagName},
			{Value: "f", Kind: CompKindFlagName},
			{Value: "bar", Description: descr, Kind: CompKindFlagName},
			{Value: "b", Description: descr, Kind: CompKindFlagName},
			{Value: "far", Kind: CompKindFlagName},
		}},
		{"-f", []CompItem{
			{Value: "f", Kind: CompKindFlagName},
		}},
		{"-x", nil},
		{"--", []CompItem{
			{Value: "foo", Kind: CompKindFlagName},
			{Value: "bar", Description: descr, Kind: CompKindFlagName},
			{Value: "far", Kind: CompKindFlagName},
		}},
		{"--f", []CompItem{
			{Value: "foo", Kind: CompKindFlagName},
			{Value: "far", Kind: CompKindFlagName},
		}},
		{"--fo", []CompItem{
			{Value: "foo", Kind: CompKindFlagName},
		}},
		{"--foo", []CompItem{
			{Value: "foo", Kind: CompKindFlagName},
		}},
		{"--undefined", nil},
	} {
		t.Run(test.toComplete, func(t *testing.T) {
			tsk := CompTask{
				ToComplete: test.toComplete,
			}

			assert.Eq(t, len(test.expected), tsk.AddFlagNames(false, flags, true))
			assert.EqS(t, test.expected, tsk.result)

			assert.Eq(t, 0, tsk.AddFlagNames(false, flags, true))
			assert.EqS(t, test.expected, tsk.result)
		})
	}
}

func TestCompTask_AddFlagValues(t *testing.T) {
	flag := &FlagEmptyV{
		Ext: &FlagHelp{
			Experimental: "",
			Deprecation:  "",
			Changelog:    "",
			Completion: &CompActionStatic{
				Suggestions: []CompItem{
					{Value: "foo", Kind: CompKindFlagValue},
				},
				State: 0,
			},
		},
	}

	for _, test := range []struct {
		toComplete string
		expected   []CompItem
	}{
		{"", []CompItem{
			{Value: "foo", Kind: CompKindFlagValue},
			{Value: "bar", Description: "default value", Kind: CompKindFlagValue},
			{Value: "far", Description: "default value", Kind: CompKindFlagValue},
		}},
		{"f", []CompItem{
			{Value: "foo", Kind: CompKindFlagValue},
			{Value: "far", Description: "default value", Kind: CompKindFlagValue},
		}},
		{"fo", []CompItem{
			{Value: "foo", Kind: CompKindFlagValue},
		}},
		{"x", nil},
	} {
		t.Run(test.toComplete, func(t *testing.T) {
			tsk := CompTask{
				ToComplete: test.toComplete,
			}

			assert.Eq(t, len(test.expected), tsk.AddFlagValues(false, flag, "[bar, far]", true))
			assert.EqS(t, test.expected, tsk.result)

			assert.Eq(t, 0, tsk.AddFlagValues(false, flag, "[bar, far]", false))
			assert.EqS(t, test.expected, tsk.result)
		})
	}
}

func TestCompTask_AddFiles(t *testing.T) {
	var tsk CompTask

	assert.Eq(t, 1, tsk.AddFiles(false, "1"))
	assert.Eq(t, 0, tsk.AddFiles(false, "2", "3", "ignored"))
	assert.Eq(t, 1, tsk.AddFiles(true))

	assert.EqS(t, []CompItem{
		{Value: "1", Kind: CompKindFiles},
		{Kind: CompKindFiles},
	}, tsk.result)
}

func TestCompTask_AddDirs(t *testing.T) {
	var tsk CompTask

	assert.Eq(t, 1, tsk.AddDirs(false, "1"))
	assert.Eq(t, 0, tsk.AddDirs(false, "2", "3", "ignored"))
	assert.Eq(t, 1, tsk.AddDirs(true))

	assert.EqS(t, []CompItem{
		{Value: "1", Kind: CompKindDirs},
		{Kind: CompKindDirs},
	}, tsk.result)
}
