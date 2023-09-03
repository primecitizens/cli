// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"strings"
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

func TestCompCmdShells(t *testing.T) {
	var cc CompCmdShells
	root := &Cmd{
		Pattern: "foo",
		Children: []*Cmd{
			cc.Setup("", -1, true),
			{
				Pattern:    "dirs",
				BriefUsage: "complete dirs",
				Completion: CompActionDirs{},
			},
			{
				Pattern:    "files-and-dirs",
				Completion: CompActionFiles{},
			},
			{
				Pattern:    "none",
				BriefUsage: "complete nothing",
				Completion: CompActionDisable{},
			},
		},
	}

	for _, shell := range []string{
		"zsh",
		"bash",
		"pwsh",
	} {
		cc.opComp.At.State_ = 0 // reset

		t.Run("PrintScript", func(t *testing.T) {
			var sb strings.Builder
			err := root.Exec(&CmdOptions{Stdout: &sb},
				"completion", shell,
			)
			assert.NoError(t, err)

			assert.False(t, strings.Contains(sb.String(), placeholderName))
			assert.False(t, strings.Contains(sb.String(), placeholderNameForIdent))
			assert.False(t, strings.Contains(sb.String(), placeholderCompletionCmdName))
			assert.True(t, sb.Len() > 100)
		})

		t.Run("PrintUsage", func(t *testing.T) {
			var sb strings.Builder
			err := root.Exec(&CmdOptions{Stdout: &sb},
				"completion", shell, "--help",
			)
			assert.ErrorIs(t, ErrHelpHandled{}, err)

			assert.False(t, strings.Contains(sb.String(), placeholderName))
			assert.False(t, strings.Contains(sb.String(), placeholderNameForIdent))
			assert.False(t, strings.Contains(sb.String(), placeholderCompletionCmdName))
			assert.True(t, sb.Len() > 100)
		})

		t.Run("BadRequest", func(t *testing.T) {
			var sb strings.Builder
			err := root.Exec(&CmdOptions{Stderr: &sb},
				"completion", shell, "complete",
			)
			assert.ErrorIs(t, &FlagViolation{
				Key:    "at",
				Reason: ViolationCodeEmptyAllOf,
			}, err)
		})

		t.Run("GoodRequest", func(t *testing.T) {
			var sb strings.Builder
			err := root.Exec(
				&CmdOptions{
					Stdout: &sb,
				},
				"completion", shell, "complete", "--at", "5",
				"--",
				"arg0", "completion", shell, "complete", "--debug-file", "",
			)
			assert.NoError(t, err)
		})
	}
}

func TestCompCmdOpComplete(t *testing.T) {
	var cmd CompCmdOpComplete
	cmd.Setup(0)

	flags := []string{"debug-file", "at", "timeout"}
	i := 0
	for ; i < len(flags); i++ {
		_, ok := cmd.NthFlag(i)
		assert.True(t, ok)
	}
	_, ok := cmd.NthFlag(i)
	assert.False(t, ok)

	for _, flag := range flags {
		f, ok := cmd.FindFlag(flag)
		assertFlagTrue(t, f, ok)
	}
	f, ok := cmd.FindFlag("non-existing")
	assertNoflagFalse(t, f, ok)
}
