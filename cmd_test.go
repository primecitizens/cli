// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

var (
	_ FlagIndexer = (*Route)(nil)
)

func TestCmdOptions_PickStdio(t *testing.T) {
	stdio := &bytes.Buffer{}
	t.Run("Nil CmdOptions", func(t *testing.T) {
		var opts *CmdOptions
		assert.Type(t, stdio, opts.PickStdin(stdio))
		assert.Type(t, stdio, opts.PickStdout(stdio))
		assert.Type(t, stdio, opts.PickStderr(stdio))
	})

	t.Run("All Set", func(t *testing.T) {
		opts := &CmdOptions{
			Stdin:  stdio,
			Stdout: stdio,
			Stderr: stdio,
		}

		assert.Type(t, stdio, opts.PickStdin(nil))
		assert.Type(t, stdio, opts.PickStdout(nil))
		assert.Type(t, stdio, opts.PickStderr(nil))
	})
}

func TestCmdHelp(t *testing.T) {
	const HELP = "nothing to be seen here."

	helpFunc := func(opts *CmdOptions, route Route, args []string, helpArgAt int) error {
		switch args[helpArgAt] {
		case "-h", "--help", "help", "-?", "?", "--Why", "how":
			opts.Stdout.Write([]byte(HELP))
		default:
		}

		return nil
	}
	root := &Cmd{
		Pattern: "foo",
		Flags:   NewMapIndexer().Add(&BoolV{}, "any"),
		Children: []*Cmd{
			{
				Pattern:    "bar",
				LocalFlags: NewMapIndexer().Add(&BoolV{}, "help", "h", "?"),
			},
		},
	}

	t.Run("DefaultHelpArgs", func(t *testing.T) {
		for _, args := range [][]string{
			{"-h"},
			{"--any", "--help"},
			{"help", "--any", "bar"},
			{"bar", "help"},
			{"help", "bar"},
			{"bar", "-h", "--any"},
			{"bar", "--any", "--help"},
		} {
			t.Run("pending", func(t *testing.T) {
				var buf bytes.Buffer
				opts := &CmdOptions{
					Stdout: &buf,
					Stderr: &buf,
				}

				err := root.Exec(opts, args...)
				assert.Type(t, &ErrHelpPending{}, err)
			})
			t.Run("handled", func(t *testing.T) {
				var buf bytes.Buffer
				opts := &CmdOptions{
					Stdout:            &buf,
					Stderr:            &buf,
					HandleHelpRequest: helpFunc,
				}

				err := root.Exec(opts, args...)
				assert.Eq(t, HELP, buf.String())
				assert.ErrorIs(t, ErrHelpHandled{}, err)
			})
		}
	})

	t.Run("CustomHelpArgs", func(t *testing.T) {
		for _, args := range [][]string{
			{"-?"},
			{"--Why"},
			{"how"},
			{"?"},
		} {
			t.Run("", func(t *testing.T) {
				var buf bytes.Buffer
				opts := &CmdOptions{
					Stdout: &buf,
					Stderr: &buf,

					ParseOptions: &ParseOptions{
						HelpArgs: []string{"-?", "--Why", "how", "?"},
					},

					HandleHelpRequest: helpFunc,
				}

				err := root.Exec(opts, args...)
				assert.Eq(t, HELP, buf.String())
				assert.ErrorIs(t, ErrHelpHandled{}, err)
			})
		}
	})
}

func TestCmd(t *testing.T) {
	type TestValues struct {
		AlicePublic  int
		AlicePrivate int

		BobPublic  int
		BobPrivate int

		CharliePublic  int
		CharliePrivate int

		FooPublic  int
		FooPrivate int

		BarPublic  int
		BarPrivate int
	}

	var (
		actualValues    TestValues
		actualPosArgs   []string
		actualDDPosArgs []string
	)

	preRunFn := func(opts *CmdOptions, route Route, i int, posArgs, dashArgs []string) error {
		return nil
	}

	runFn := func(opts *CmdOptions, route Route, posArgs, dashArgs []string) error {
		actualPosArgs, actualDDPosArgs = posArgs, dashArgs
		return nil
	}

	postRunFn := func(opts *CmdOptions, route Route, i int, runErr error) error {
		return nil
	}

	createIndexer := func(name string, short rune, ret Flag) FlagIndexer {
		return FuncIndexer(func(flag string, index int) (f Flag, info FlagInfo, ok bool) {
			switch flag {
			case name, string(short):
				return ret, FlagInfo{Name: name, Shorthand: string(short)}, true
			default:
				return
			}
		})
	}

	root := Cmd{
		Pattern: "Alice",
		PreRun:  preRunFn,
		Run:     runFn,
		PostRun: postRunFn,
		Flags: createIndexer("alice-public", 'A', &IntSum{
			Value: &actualValues.AlicePublic,
		}),
		LocalFlags: createIndexer("alice-private", 'a', &Int{
			Value: &actualValues.AlicePrivate,
		}),
		Children: []*Cmd{
			{
				Pattern: "NotRunnable",
			},
			{
				Pattern: "Bob",
				PreRun:  preRunFn,
				Run:     runFn,
				PostRun: postRunFn,
				Flags: createIndexer("bob-public", 'B', &IntSum{
					Value: &actualValues.BobPublic,
				}),
				LocalFlags: createIndexer("bob-private", 'b', &Int{
					Value: &actualValues.BobPrivate,
				}),
				Children: []*Cmd{
					{
						Pattern: "Foo",
						PreRun:  preRunFn,
						Run:     runFn,
						PostRun: postRunFn,
						Flags: createIndexer("foo-public", 'F', &IntSum{
							Value: &actualValues.FooPublic,
						}),
						LocalFlags: createIndexer("foo-private", 'f', &Int{
							Value: &actualValues.FooPrivate,
						}),
					},
				},
			},
			{
				Pattern: "Charlie",
				PreRun:  preRunFn,
				Run:     runFn,
				PostRun: postRunFn,
				Flags: createIndexer("charlie-public", 'C', &IntSum{
					Value: &actualValues.CharliePublic,
				}),
				LocalFlags: createIndexer("charlie-private", 'c', &Int{
					Value: &actualValues.CharliePrivate,
				}),
			},
		},
	}

	t.Run("FormatRoute", func(t *testing.T) {
		var sb strings.Builder
		n, err := FormatRoute(&sb, Route{
			&root,
			root.Children[1],
			root.Children[1].Children[0],
		}, " ")
		assert.NoError(t, err)
		assert.Eq(t, sb.Len(), n)
		assert.Eq(t, "Alice Bob Foo", sb.String())
	})

	t.Run("NotRunnable", func(t *testing.T) {
		err := root.Exec(nil, "NotRunnable")
		assert.ErrorIs(t, &ErrCmdNotRunnable{
			Name: "NotRunnable",
		}, err)
	})

	t.Run("bad-args", func(t *testing.T) {
		called := 0
		err := root.Exec(
			&CmdOptions{
				HandleArgError: func(opts *CmdOptions, route Route, args []string, i int, argErr error) error {
					called++
					assert.Eq(t, 0, i)
					assert.ErrorIs(t, &ErrFlagUndefined{Name: args[i][2:], At: i}, argErr)
					return argErr
				},
			},
			"--invalid-flag", "foo",
		)
		assert.Error(t, err)
		assert.Eq(t, 1, called)

		called = 0
		poptCalled := 0
		err = root.Exec(
			&CmdOptions{
				ParseOptions: &ParseOptions{
					HandleParseError: func(opts *ParseOptions, args []string, i int, argErr error) (err error) {
						poptCalled++
						assert.Eq(t, 1, i)
						assert.ErrorIs(t, &ErrFlagUndefined{Name: args[i][2:], At: i}, argErr)
						return argErr
					},
				},
				HandleArgError: func(opts *CmdOptions, route Route, args []string, i int, argErr error) error {
					called++
					assert.Eq(t, 1, i)
					assert.ErrorIs(t, &ErrFlagUndefined{Name: args[i][2:], At: i}, argErr)
					return argErr
				},
			},
			"Alice", "--invalid-flag", "foo",
		)
		assert.Error(t, err)
		assert.Eq(t, 1, called)
		assert.Eq(t, 1, poptCalled)
	})

	for _, test := range []struct {
		name string
		args []string

		values   TestValues
		posArgs  []string
		dashArgs []string
	}{
		{
			name: "Run Alice",
			args: []string{
				"--alice-public", "--alice-private", "100",
			},
			values: TestValues{
				AlicePublic:  1,
				AlicePrivate: 100,
			},
		},
		{
			name: "Run Bob",
			args: []string{
				"-Aa", "100",
				"Bob", "-ABb=200",
			},
			values: TestValues{
				AlicePublic:  2,
				AlicePrivate: 100,
				BobPublic:    1,
				BobPrivate:   200,
			},
		},
		{
			name: "Run Foo",
			args: []string{
				"--alice-public", "--alice-private", "100",
				"Bob", "-AB", "--bob-private=200",
				"Foo", "--alice-public", "--bob-public", "--foo-public", "--foo-private", "300",
			},
			values: TestValues{
				AlicePublic:  3,
				AlicePrivate: 100,
				BobPublic:    2,
				BobPrivate:   200,
				FooPublic:    1,
				FooPrivate:   300,
			},
		},
		{
			name: "Run Bob with positional args",
			args: []string{
				"Bob", "a", "b", "c",
			},
			values:  TestValues{},
			posArgs: []string{"a", "b", "c"},
		},
		{
			name: "Run Foo with positional and dash args",
			args: []string{
				"Bob", "Foo", "a", "b", "c", "--", "d", "e", "f",
			},
			values:   TestValues{},
			posArgs:  []string{"a", "b", "c"},
			dashArgs: []string{"d", "e", "f"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			actualValues = TestValues{}
			actualPosArgs = nil
			actualDDPosArgs = nil

			err := root.Exec(nil, test.args...)
			assert.NoError(t, err)
			assert.Eq(t, test.values, actualValues)
			assert.EqS(t, test.posArgs, actualPosArgs)
			assert.EqS(t, test.dashArgs, actualDDPosArgs)
		})
	}
}

func TestCmdFlagDefaultValue(t *testing.T) {
	type Config struct {
		Str string `cli:"foo,def=str"`
	}

	var (
		preRunCalled  bool
		runCalled     bool
		postRunCalled bool
		actual        Config
	)
	root := Cmd{
		Flags: NewReflectIndexer(DefaultReflectVPFactory{}, &actual),
		PreRun: func(opts *CmdOptions, route Route, prerunAt int, posArgs, dashArgs []string) error {
			preRunCalled = true
			assert.Eq(t, "str", actual.Str)
			return nil
		},
		Run: func(opts *CmdOptions, route Route, posArgs, dashArgs []string) error {
			runCalled = true
			assert.Eq(t, "str", actual.Str)
			return nil
		},
		PostRun: func(opts *CmdOptions, route Route, postrunAt int, runErr error) error {
			postRunCalled = true
			assert.Eq(t, "str", actual.Str)
			return nil
		},
	}

	err := root.Exec(nil)
	assert.NoError(t, err)
	assert.Eq(t, "str", actual.Str)
	assert.True(t, preRunCalled)
	assert.True(t, runCalled)
	assert.True(t, postRunCalled)
}

func BenchmarkCmd(b *testing.B) {
	var (
		flag Int64SumV
	)

	root := Cmd{
		Flags: FuncIndexer(func(flagName string, index int) (f Flag, into FlagInfo, ok bool) {
			switch flagName {
			case "sum":
				return &flag, FlagInfo{Name: "sum"}, true
			default:
				return
			}
		}),
	}

	n := 0
	runFn := func(opts *CmdOptions, route Route, posArgs, posArgsAfterDoubleDash []string) error {
		n++
		return nil
	}

	// 15 sub commands each with 1 flag referring to the flag define in
	// the root command
	args := make([]string, 30)
	for i, cur := 0, &root; i < len(args); i += 2 {
		cur.Children = []*Cmd{{
			Pattern: "sub",
			Run:     runFn,
		}}
		cur = cur.Children[0]

		args[i] = "sub"
		args[i+1] = "--sum=1"
	}

	opts := &CmdOptions{
		RouteBuf: make(Route, 0, 16),
	}
	if !assert.NoError(b, root.Exec(opts, args...)) {
		return
	}

	n = 0
	flag.Value = 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if root.Exec(opts, args...) != nil {
			b.FailNow()
		}
	}
	b.StopTimer()

	assert.Eq(b, b.N, n)
	assert.Eq(b, int64(b.N)*int64(len(args)/2), flag.Value)
}
