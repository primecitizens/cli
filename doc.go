// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

// Package cli supports POSIX & GNU style cli flag parsing and command running.
//
// # Terminology
//
//   - `args`: all strings provided to a cmd.
//     For a root command in real world, it usually is `os.Args[1:]`
//   - `flags`: before the first dash, strings interpreted as flag names
//     and flag values by POSIX and GNU style guide.
//   - `subcmds` (sub-commands): before the first dash, consecutive args
//     matching a serial of `Cmd.Pattern`.
//     In the below illustration, if there is a `Cmd` in root command's
//     `Children []*Cmd` field whose `Cmd.Pattern` matches `xxx`, then the
//     posArg `xxx` becomes subcmd `xxx`.
//   - `posArgs` (positional args): before the first dash, strings that
//     are not flags and subcmds.
//   - `dashArgs`: all strings after the dash.
//
// Illustration without subcmds:
//
//	                      dash
//	                        |
//	    posArg  flag name   |
//	       |        |       |
//	./foo xxx -i --join bar -- other args
//	          |          |    [ all args after the dash are dashArgs]
//	          |          |
//	          |      flag value
//	          |
//	  flag shorthand, with implicit value
package cli
