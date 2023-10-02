// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"unicode/utf8"
)

// FlagHelp defines commonly used flag metadata
//
// Hint: use .Extra of a flag to hold FlagHelp for documentation
// purpose.
type FlagHelp struct {
	// Experimental explains why the flag is Experimental.
	//
	// Only use this field to note the reason of being experimental, and
	// DO NOT include prefix like `EXPERIMENTAL: `
	Experimental string

	// Deprecation is the deprecation message of the flag.
	//
	// Only use this field to note the reason of deprecation, and
	// DO NOT include prefix like `DEPRECATED: `
	Deprecation string

	// Changelog may contain upgrade notice of the flag.
	//
	// DO NOT include prefix like `CHANGELOG: `
	Changelog string

	// Completion is the completion action to be called to add suggestions.
	Completion CompAction

	// Extra custom data.
	Extra any
}

// Suggest implements [CompAction].
func (h *FlagHelp) Suggest(tsk *CompTask) (added int, state CompState) {
	if h.Completion == nil {
		return 0, 0
	}

	return h.Completion.Suggest(tsk)
}

func (h *FlagHelp) HelplnCmdTerminal(out io.Writer, route Route, prefix string) (int, error) {
	return 0, nil
}

func (h *FlagHelp) HelplnFlagTerminal(
	out io.Writer, route Route, prefix string, indent int, info FlagInfo, groupHasShorthand bool,
) (n int, err error) {
	cursor, n, err := printFlagBasicInfo(out, route, prefix, indent, info, groupHasShorthand)
	if err != nil {
		return
	}

	var x int
	if cursor > indent {
		x, err = wstr(out, "\n")
		n += x
		if err != nil {
			return
		}
		cursor = 0
	}

	if len(h.Deprecation) != 0 {
		x, err = wstr(out, prefix)
		n += x
		if err != nil {
			return
		}

		x, err = writeSpaces(out, indent-cursor)
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = write(out, h.Deprecation, "\n", "DEPRECATED: ")
		n += x
		cursor += x
		if err != nil {
			return
		}
		cursor = 0
	}

	if len(h.Experimental) != 0 {
		x, err = wstr(out, prefix)
		n += x
		if err != nil {
			return
		}

		x, err = writeSpaces(out, indent-cursor)
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = write(out, h.Experimental, "\n", "EXPERIMENTAL: ")
		n += x
		cursor += x
		if err != nil {
			return
		}
		cursor = 0
	}

	return
}

// printlnValidFlag
func printlnValidFlag(
	out io.Writer, route Route, linePrefix string,
	indent int, info FlagInfo, groupHasShorthand bool,
) (cursor, n int, err error) {
	if (len(info.Name) == 0 || IsShorthand(info.Name)) &&
		(len(info.Shorthand) == 0 || !IsShorthand(info.Shorthand)) {
		return
	}

	proute := noescape(&route)
	_, flag, ok := FindFlag(proute, info.Name, info.Shorthand)
	if !ok {
		return
	}

	if ft, ok := flag.(HelperTerminal); ok {
		n, err = ft.HelplnFlagTerminal(out, route, linePrefix, indent, info, groupHasShorthand)
		return
	} else if ft, ok = flag.Extra().(HelperTerminal); ok {
		n, err = ft.HelplnFlagTerminal(out, route, linePrefix, indent, info, groupHasShorthand)
		return
	}

	cursor, n, err = printFlagBasicInfo(out, route, linePrefix, indent, info, groupHasShorthand)
	if err != nil {
		return
	}

	x, err := wstr(out, "\n")
	n += x
	cursor = 0
	return
}

func printFlagBasicInfo(
	out io.Writer, route Route, prefix string, indent int, info FlagInfo, groupHasShorthand bool,
) (cursor, n int, err error) {
	var (
		x        int
		hasShort = IsShorthand(info.Shorthand)
	)

	if hasShort {
		if len(prefix) != 0 {
			x, err = wstr(out, prefix)
			n += x
			if err != nil {
				return
			}
		}

		x, err = wstr(out, "-")
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = wstr(out, info.Shorthand)
		n += x
		if err != nil {
			return
		}
		cursor += 1
	}

	if len(info.Name) != 0 && !IsShorthand(info.Name) {
		if hasShort {
			x, err = wstr(out, " --")
		} else {
			if len(prefix) != 0 {
				// doesn't count as cursor
				x, err = wstr(out, prefix)
				n += x
				if err != nil {
					return
				}
			}

			if groupHasShorthand {
				x, err = wstr(out, "   --")
			} else {
				x, err = wstr(out, "--")
			}
		}
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = wstr(out, info.Name)
		n += x
		if err != nil {
			return
		}

		cursor += utf8.RuneCountInString(info.Name)
	} else if !hasShort {
		// defensive check, should have been filtered out
		return
	}

	proute := noescape(&route)
	_, flag, ok := FindFlag(proute, info.Name, info.Shorthand)
	if !ok {
		return
	}

	typ, ok := flag.Type()
	if ok && cursor > 0 && len(typ) != 0 {
		x, err = write(out, typ, "", " ")
		n += x
		if err != nil {
			return
		}
		cursor += 1 + utf8.RuneCountInString(typ)
	}

	cursor, x, err = printFlagRules(out, route, info, indent, cursor)
	n += x
	if err != nil {
		return
	}

	if usage := flag.Usage(); len(usage) > 0 {
		x, err = writeSpaces(out, indent-cursor)
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = wstr(out, usage)
		n += x
		cursor += utf8.RuneCountInString(usage)
		if err != nil {
			return
		}
	}

	if len(info.DefaultValue) != 0 {
		x, err = write(out, info.DefaultValue, ")", " (default: ")
		n += x
		cursor += x // approx
		if err != nil {
			return
		}
	} else if !flag.State().ValueChanged() && flag.HasValue() {
		// default value implied by flag state
		x, err = wstr(out, " (default: ")
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = flag.PrintValue(out)
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = wstr(out, ")")
		n += x
		cursor += x
		if err != nil {
			return
		}
	}

	return
}

func printFlagRules(out io.Writer, route Route, info FlagInfo, indent, cur int) (cursor, n int, err error) {
	var x int
	cursor = cur

	proute := noescape(&route)
	_, flag, ok := FindFlag(proute, info.Name, info.Shorthand)
	if !ok {
		return
	}

	if flag.State().Hidden() {
		x, err = writeSpaces(out, indent-cursor)
		n += x
		cursor += x
		if err != nil {
			return
		}

		x, err = wstr(out, "(hidden")
		n += x
		cursor += x
		if err != nil {
			return
		}
	}

	hasTag := false
	// 1st round: check if required
	for i := len(route) - 1; i >= 0; i-- {
		rule := route[i].FlagRule
		if rule == nil {
			continue
		}

		if RuleRequiresAny(rule, info.Name, info.Shorthand) {
			if cursor > indent {
				x, err = wstr(out, ", ")
			} else {
				x, err = writeSpaces(out, indent-cursor)
				n += x
				cursor += x
				if err != nil {
					return
				}

				x, err = wstr(out, "(")
			}
			n += x
			cursor += x
			if err != nil {
				return
			}

			x, err = wstr(out, "required) ")
			n += x
			cursor += x
			return
		}

		if !hasTag && RuleContainsAny(rule, info.Name, info.Shorthand) {
			hasTag = true
		}
	}

	if !hasTag {
		goto Fixup
	}

	// 2nd round, only when the flag is not required: add non required tags
	for i := len(route) - 1; i >= 0; i-- {
		rule := route[i].FlagRule
		if rule == nil {
			continue
		}

		if RuleContainsAny(rule, info.Name, info.Shorthand) {
			if cursor > indent {
				x, err = wstr(out, ", ")
			} else {
				x, err = writeSpaces(out, indent-cursor)
				n += x
				cursor += x
				if err != nil {
					return
				}

				x, err = wstr(out, "(")
			}
			n += x
			cursor += x
			if err != nil {
				return
			}

			x, err = callWriteFlagRule(rule, out, info.Name, info.Shorthand)
			n += x
			cursor += x
			if err != nil {
				return
			}
		}
	}

Fixup:
	if cursor > indent {
		x, err = wstr(out, ") ")
		n += x
		cursor += x
		return
	}

	return
}
