// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/primecitizens/cli"
)

type Config struct {
	Subject string                `cli:"s|subject,hide,#a simple string flag"`
	Dur     *time.Duration        `cli:"d|duration,value=dur,#example duration flag"`
	Size    uint64                `cli:"size,value=size,def=32G,#size accepting common units"`
	Entries *[]string             `cli:"e|entry,#log entries"`
	Metrics map[time.Duration]int `cli:"m|metric,key=dur,value=sum,#metrics summary"`
	Pattern []*regexp.Regexp      `cli:"p|pattern,value=regexp,def=.*,def=^p,#regexp patterns"`
}

func run(opts *cli.CmdOptions, route cli.Route, posArgs, dashArgs []string) error {
	config := route.Root().Extra.(*Config)

	fmt.Println("After", config.Dur.Hours(), "hours,",
		config.Subject, "would have", config.Size, "bytes of data")
	fmt.Println("Some of the entries will be like:")
	fmt.Println(strings.Join(*config.Entries, "\n"))
	fmt.Println("Some of the metrics will be like:")
	var keys []time.Duration
	for k := range config.Metrics {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		fmt.Println("[metric]", k.String(), "=", config.Metrics[k])
	}
	return nil
}

func ExampleCmd_reflectIndexer() {
	var cfg Config
	root := cli.Cmd{
		Pattern: "example",
		Flags:   cli.NewReflectIndexer(cli.DefaultReflectVPFactory{}, &cfg),
		Run:     run,
		Extra:   &cfg, // run func will find this
	}

	err := root.Exec(nil,
		"-d", "32.5d", "--subject", "log file", "--size", "10T",
		"-e", "[app] err", "-e", "[app] fatal",
		"-m", "10s=2", "--metric", "10s=5",
	)
	if err != nil {
		panic(err)
	}

	// Output:
	// After 780 hours, log file would have 10995116277760 bytes of data
	// Some of the entries will be like:
	// [app] err
	// [app] fatal
	// Some of the metrics will be like:
	// [metric] 10s = 7
}

func ExampleCmd_customHelpArg() {
	opts := &cli.CmdOptions{
		ParseOptions: &cli.ParseOptions{
			// define custom args can initiate help request.
			HelpArgs: []string{"-?", "why not?"},
		},
		Stderr: os.Stdout, // override stderr to work with golang example test
		// You can provide your own HelpHandleFunc, but here we use the one comes with
		// the library.
		HandleHelpRequest: cli.HandleHelpRequest,
	}

	root := cli.Cmd{
		Pattern:    "example -?",
		BriefUsage: "Just like all other indexers, terminal help also works with the ReflectIndexer",
		Flags: cli.NewReflectIndexer(cli.DefaultReflectVPFactory{}, &Config{
			Subject: "Some default value deduced from flag state",
		}),
	}

	err := root.Exec(opts, "-?")
	if err != nil && !errors.Is(err, cli.ErrHelpHandled{}) {
		panic(err)
	}

	// Output:
	// example -?
	//
	// Just like all other indexers, terminal help also works with the ReflectIndexer
	//
	// Flags:
	//   -s --subject str          (hidden) a simple string flag (default: Some default value deduced from flag state)
	//   -d --duration dur         example duration flag
	//      --size size            size accepting common units (default: 32G)
	//   -e --entry []str          log entries
	//   -m --metric map[dur]isum  metrics summary
	//   -p --pattern []regexp     regexp patterns (default: [.*, ^p])
}

func ExampleCmd_mapIndexer() {
	var (
		str string

		flagB cli.BoolV
		flagI cli.IntV
		flagS = cli.String{
			Value: &str,
		}
	)

	root := &cli.Cmd{
		Pattern:    "example {print-flag-values|help|completion}",
		BriefUsage: "Using MapIndexer feels the same as a builder (but way more decentralized).",
		Flags: cli.NewMapIndexer().
			Add(&flagB, "bool", "B").
			Add(&flagI, "int", "I").
			Add(&flagS, "string", "S"),
		LocalFlags: nil,
		FlagRule:   cli.AllOf("bool", "int"),
		Completion: cli.CompActionFiles{},
		Children: []*cli.Cmd{
			{
				Pattern: "print-flag-values",
				Run: func(opts *cli.CmdOptions, route cli.Route, posArgs, dashArgs []string) error {
					fmt.Println("bool:", flagB.Value)
					fmt.Println("int:", flagI.Value)
					fmt.Println("string:", str)
					fmt.Println("posArgs:", posArgs)
					fmt.Println("dashArgs:", dashArgs)
					return nil
				},
			},
			(&cli.CompCmdShells{}).Setup("", -1, false),
		},
	}

	err := root.Exec(&cli.CmdOptions{
		HandleArgError:    cli.HandleArgErrorAsHelpRequest,
		HandleHelpRequest: cli.HandleHelpRequest,
	},
		"print-flag-values", "-BI", "10", "--string", "str", "pos1", "--", "dd1",
	)
	if err != nil {
		panic(err)
	}

	// Output:
	// bool: true
	// int: 10
	// string: str
	// posArgs: [pos1]
	// dashArgs: [dd1]
}
