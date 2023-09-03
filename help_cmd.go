// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"unicode/utf8"
)

// CmdHelp contains commonly used metadata for a cli command.
//
// To produce consistent content, all values SHOULD NOT contain leading or
// trailling whitespaces.
//
// Hint: use Cmd.Extra to hold *CmdHelp for documentation purpose.
type CmdHelp struct {
	// Example lists some typical use cases of the command.
	//
	// DO NOT include prefix like `Example: `
	Example string

	// LongDescription is the detailed description of the command.
	LongDescription string

	// Experimental explains why the command is Experimental.
	//
	// Only use this field to note the reason of being experimental, and
	// DO NOT include prefix like `Experimental: `
	Experimental string

	// Deprecation is the deprecation message of the command.
	//
	// Only use this message to note the reason of deprecation, and
	// DO NOT include prefix like `Deprecated: `
	Deprecation string

	// Changelog may contain upgrade notice of the command.
	//
	// DO NOT include prefix like `Changelog: `
	Changelog string

	// Extra custom data
	Extra any
}

// HelplnCmdTerminal implements [TerminalHelper].
func (m *CmdHelp) HelplnCmdTerminal(out io.Writer, route Route, linePrefix string) (n int, err error) {
	var x int
	if hasRouteLine(route) {
		if len(linePrefix) != 0 {
			n, err = wstr(out, linePrefix)
			if err != nil {
				return
			}
		}

		x, err = FormatRoute(out, route, " ")
		n += x
		if err != nil {
			return
		}
	}

	x, err = write(out, route.Target().BriefUsage, "", "\n\n", linePrefix)
	n += x
	if err != nil {
		return
	}

	x, err = printSubcmds(out, route.Target().Children, "\n\n", linePrefix)
	n += x
	if err != nil {
		return
	}

	x, err = write(out, m.Deprecation, "", "\n\n", linePrefix, "DEPRECATED: ")
	n += x
	if err != nil {
		return
	}

	x, err = write(out, m.Experimental, "", "\n\n", linePrefix, "EXPERIMENTAL: ")
	n += x
	if err != nil {
		return
	}

	x, err = write(out, m.Example, "", "\n\n", linePrefix, "Example:\n\n", linePrefix)
	n += x
	if err != nil {
		return
	}

	x, err = printlnTargetCmdFlags(out, route, "\n\nFlags:\n", linePrefix+"  ")
	n += x
	if err != nil {
		return
	}

	x, err = write(out, m.Changelog, "\n\n", "\nChanges:\n\n")
	n += x
	return
}

// HelplnFlagTerminal implements [TerminalHelper].
func (h *CmdHelp) HelplnFlagTerminal(
	out io.Writer, route Route, linePrefix string,
	indent int, flag FlagInfo, groupHasShorthand bool,
) (int, error) {
	return 0, nil
}

func printSubcmds(out io.Writer, children []*Cmd, before, linePrefix string) (n int, err error) {
	var (
		x     int
		wrote bool
	)
	for _, c := range children {
		if len(c.Pattern) == 0 {
			continue
		}

		if !wrote {
			wrote = true
			x, err = wstr(out, before)
			n += x
			if err != nil {
				return
			}

			x, err = wstr(out, "Sub-Commands:")
			n += x
			if err != nil {
				return
			}
		}

		x, err = write(out, c.Pattern, "", "\n", linePrefix, "- ")
		n += x
		if err != nil {
			return
		}
	}

	return
}

// printlnTargetCmdFlags writes cli usage text of all flags accessible to this Cmd.
func printlnTargetCmdFlags(
	out io.Writer, route Route, before, linePrefix string,
) (n int, err error) {
	var (
		// cursorFlagEnd is the cursor position where flag description aligns.
		cursorFlagEnd int
		proute        = noescape(&route)

		hasShorthand bool
	)

	for i := 0; ; i++ {
		info, ok := proute.NthFlag(i)
		if !ok {
			break
		}

		x := utf8.RuneCountInString(info.Name)
		if x > 1 {
			x += 2 // `--`
		} else { // either empty or invalid long name
			x = 0
		}

		if IsShorthand(info.Shorthand) {
			hasShorthand = true

			if x != 0 {
				x += 1 // space sep between name and shorthand
			}

			x += 2 // `-f`
		}

		if x == 0 { // no name
			continue
		}

		_, flag, ok := FindFlag(proute, info.Name, info.Shorthand)
		if !ok {
			continue
		}

		typ, ok := flag.Type()
		if ok && len(typ) != 0 {
			x += utf8.RuneCountInString(typ) + 1 // space between flag name and type
		}

		if x > cursorFlagEnd {
			cursorFlagEnd = x
		}
	}

	if cursorFlagEnd > 0 {
		cursorFlagEnd += 2 // 2 spaces between flag name and description

		var x int
		x, err = wstr(out, before)
		n += x
		if err != nil {
			return
		}

		for i := 0; ; i++ {
			info, ok := proute.NthFlag(i)
			if !ok {
				break
			}

			_, x, err = printlnValidFlag(out, route, linePrefix, cursorFlagEnd, info, hasShorthand)
			n += x
			if err != nil {
				return
			}
		}
	} else {
		_, err = wstr(out, "\n")
	}

	return
}

// write writes when len(content) != 0
func write(out io.Writer, content, suffix string, prefixes ...string) (n int, err error) {
	if len(content) == 0 {
		return 0, nil
	}

	var x int
	for _, p := range prefixes {
		if len(p) == 0 {
			continue
		}

		x, err = wstr(out, p)
		n += x
		if err != nil {
			return
		}
	}

	x, err = wstr(out, content)
	n += x
	if err != nil {
		return
	}

	if len(suffix) != 0 {
		x, err = wstr(out, suffix)
		n += x
	}

	return
}
