// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package main

import (
	"os"
	"regexp"
	"time"

	"github.com/primecitizens/cli"
)

type Config struct {
	FileMatch map[string][]*regexp.Regexp `cli:"f|filematch,value=regexp,comp=x=.*,comp=x=foo,comp=x=a"`
}

var (
	scc cli.CompCmdShells

	root = cli.Cmd{
		Pattern: "e2e",
		Flags:   cli.NewReflectIndexer(cli.DefaultReflectVPFactory{}, &Config{}),
		Children: []*cli.Cmd{
			scc.Setup("", time.Second, false),
			{
				Pattern: "generic",
				Children: []*cli.Cmd{
					{
						Pattern:    "dir",
						BriefUsage: "test dir only completion",
						Completion: cli.CompActionDirs{},
					},
					{
						Pattern:    "file",
						BriefUsage: "test file & dir completion",
						Completion: cli.CompActionFiles{},
					},
				},
			},
			{
				Pattern: "bash",
				Children: []*cli.Cmd{
					{
						Pattern:    "long-description-cmd",
						BriefUsage: "a somewhat long description with other long command.",
					},
					{
						Pattern:    "command-with-a-long-name-to-test-completion-description",
						BriefUsage: "short",
					},
				},
			},
			{
				Pattern: "zsh",
				Children: []*cli.Cmd{
					{
						Pattern:    "colon:in:value",
						BriefUsage: "colons should be escaped",
					},
					{
						Pattern:    "colon-in-description",
						BriefUsage: "colon:in:description doesn't need to be escaped",
					},
				},
			},
		},
	}
)

func main() {
	err := root.Exec(
		&cli.CmdOptions{
			HandleHelpRequest: cli.HandleHelpRequest,
			// HandleArgError:    cli.HandleArgErrorAsHelpRequest,
		},
		os.Args[1:]...,
	)
	if err != nil {
		panic(err)
	}
}
