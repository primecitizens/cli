// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"os"
)

// HelperTerminal writes help messages for a terminal user.
type HelperTerminal interface {
	// HelplnCmdTerminal writes help messages for the target command
	// with a newline in the end.
	HelplnCmdTerminal(out io.Writer, route Route, linePrefix string) (int, error)

	// HelplnFlagTerminal writes help meesages for the flag with a newline
	// in the end.
	HelplnFlagTerminal(
		out io.Writer, route Route, linePrefix string,
		indent int, flag FlagInfo, groupHasShorthand bool,
	) (int, error)
}

// HandleHelpRequest calls HandleArgErrorAsHelpRequest with nil error.
func HandleHelpRequest(
	opts *CmdOptions, route Route, args []string, helpArgAt int,
) error {
	return HandleArgErrorAsHelpRequest(opts, route, args, helpArgAt, nil)
}

// HandleArgErrorAsHelpRequest prints the error and usage text of the target
// command to stderr, it tries to cast Cmd.Extra and Flag.Extra() as
// TerminalHelper to write messages.
func HandleArgErrorAsHelpRequest(
	opts *CmdOptions, route Route, args []string, badArgAt int, cmdErr error,
) error {
	c := route.Target()
	if c == nil {
		return cmdErr
	}

	out := opts.PickStderr(os.Stderr)

	if cmdErr != nil {
		// ignore errors to always write the error line as a whole
		_, _ = wstr(out, "Error: ")
		_, _ = wstr(out, cmdErr.Error())
		_, _ = wstr(out, "\n\n")
	}

	const LinePrefix = ""

	switch ct := c.Extra.(type) {
	case HelperTerminal:
		_, _ = ct.HelplnCmdTerminal(out, route, LinePrefix)
		return cmdErr
	default:
		var err error
		if hasRouteLine(route) {
			if len(LinePrefix) != 0 {
				_, err = wstr(out, LinePrefix)
				if err != nil {
					return cmdErr
				}
			}

			_, err = FormatRoute(out, route, " ")
			if err != nil {
				return cmdErr
			}
		}

		_, err = write(out, route.Target().BriefUsage, "", "\n\n", LinePrefix)
		if err != nil {
			return cmdErr
		}

		_, err = printSubcmds(out, c.Children, "\n\n", LinePrefix)
		if err != nil {
			return cmdErr
		}

		_, err = printlnTargetCmdFlags(out, route, "\n\nFlags:\n", LinePrefix+"  ")
		return cmdErr
	}
}

func hasRouteLine(r Route) bool {
	for _, c := range r {
		if len(c.Name()) != 0 {
			return true
		}
	}

	return false
}

func writeSpaces(out io.Writer, count int) (n int, err error) {
	for i, x := 0, 0; i < count; i++ {
		x, err = wstr(out, " ")
		n += x
		if err != nil {
			return
		}
	}

	return
}
