// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"strings"
	"time"
	"unicode/utf8"
)

type (
	// ParseErrorHandleFunc handles the parseErr when parsing args[posErrArg].
	//
	// Return nil to ignore the parseErr and continue parsing.
	ParseErrorHandleFunc = func(opts *ParseOptions, args []string, posErrArg int, parseErr error) (err error)
)

// ParseOptions are options for flag parsing.
type ParseOptions struct {
	// StartTime is assumed to be the time of parsing start.
	StartTime time.Time

	// HandleParseError is the function to handle flag parsing errors.
	HandleParseError ParseErrorHandleFunc

	// PosArgsBuf is the buffer used by ParseFlags{,LowLevel} for appending
	// positional args.
	PosArgsBuf []string

	// HelpArgs are arg values can initiate help request.
	//
	//	- HelpArgs = nil, is equivalent to HelpArgs = []string{"--help", "-h", "help"}.
	//	- HelpArgs = []string{} will disable the help system.
	//	- otherwise, use the supplied HelpArgs to match args.
	HelpArgs []string

	// Extra custom data.
	Extra any
}

// IsHelpArg returns true if x is supposed to be an arg requesting help.
func (c *ParseOptions) IsHelpArg(x string) bool {
	if c == nil || c.HelpArgs == nil {
		switch x {
		case "--help", "-h", "help":
			return true
		}
	} else {
		for _, cmd := range c.HelpArgs {
			if cmd == x {
				return true
			}
		}
	}

	return false
}

// ParseFlags parses args with options, where args are usually the os.Args[1:].
//
// Return value posArgs are positional args, dashArgs are args after the first
// dash (`--`).
//
// Known limitations:
//
//   - Flags allow implicit value cannot accept valid standalone value prefixed
//     with hyphen (`-`) due to ambiguity.
//     e.g. `--foo -1` where flag `foo` is of type IntSum.
//     To workaround, use `--foo=-1`.
//
//   - Standalone dash (`--`) can never become flag value or positional
//     arg.
//     To workaround for flag value, use `--flag=--`.
//     There is no workaround to make `--` as a posArg.
func ParseFlags(args []string, flags FlagFinder, opts *ParseOptions) (posArgs, dashArgs []string, err error) {
	if opts != nil {
		posArgs = opts.PosArgsBuf
	}

	_, posDash, _, posArgs, _, err := ParseFlagsLowLevel(
		args, flags, opts,
		0,       // offset
		true,    // appendPosArgs
		false,   // stopAtFirstPosArg
		true,    // setFlags
		posArgs, // posArgsBuf
	)

	if posDash >= 0 {
		dashArgs = args[posDash+1:]
	}

	return
}

// ParseFlagsLowLevel is the low-level version of ParseFlags with more
// options for flag parsing control.
//
// If appendPosArgs is false, this function doesn't append positional args to
// posArgs, so the return value posArgs will be exactly the same a posArgsBuf.
//
// If stopAtFirstPosArg is true, this function returns on reaching the first
// positional arg (before the dash), that arg is not added to nParsed.
//
// If len(opts.HelpArgs) > 0, this function returns on reaching the first flag
// that matches any of help flags, in which case, the return value helpArgAt
// is expected to be greater or equal to zero, and args[helpArgAt] is the arg
// triggered this return.
//
// If setFlagValue is false, this function calls Flag.Decode() with set = false.
//
// The return value nParsed is the count of args parsed, which includes the
// bad flag on error return (in which case there are nParsed-1 known good
// args).
//
// The return value posDash is the index into args, when posDash >= 0,
// args[posDash] = "--" and args[posDash+1:] is dashArgs.
func ParseFlagsLowLevel(
	args []string,
	flags FlagFinder,
	opts *ParseOptions,
	offset int,
	appendPosArgs bool,
	stopAtFirstPosArg bool,
	setFlagValue bool,
	posArgsBuf []string,
) (
	nParsed int,
	posDash int,
	foundPosArg bool,
	posArgs []string,
	helpArgAt int,
	err error,
) {
	posArgs = posArgsBuf
	posDash = -1
	helpArgAt = -1
	i := offset

	for ; i < len(args); i++ {
		arg := args[i]

		szArg := len(arg)
		if szArg == 0 || arg[0] != '-' || szArg == 1 /* '-' */ {
			foundPosArg = true

			if appendPosArgs {
				posArgs = append(posArgs, arg)
			}

			if isHelpArg := opts.IsHelpArg(arg); isHelpArg || stopAtFirstPosArg {
				if isHelpArg {
					helpArgAt = i
				}

				// DO NOT include the pos arg (this one).
				nParsed = i - offset
				return
			}

			continue
		}

		var shiftNext bool
		if arg[1] == '-' {
			if szArg == 2 {
				// dash
				nParsed = len(args)
				posDash = i
				return
			}

			if opts.IsHelpArg(arg) {
				helpArgAt = i
				if _, ok := flags.FindFlag(arg[2:]); ok {
					// there is real help flag, parse it as application may expect
					// its value getting set.
					shiftNext, err = parseLongFlag(flags, opts, args, i, setFlagValue)
				} else {
					// TODO: the help flag is a pseudo flag.
				}

				nParsed = i + 1 - offset
				return
			}

			shiftNext, err = parseLongFlag(flags, opts, args, i, setFlagValue)
		} else {
			if opts.IsHelpArg(arg) {
				helpArgAt = i
				if _, ok := flags.FindFlag(arg[1:]); ok {
					// there is real help flag, parse it as application may expect
					// its value getting set.
					//
					// TODO: should we only parse the help flag instead of the whole shorthand cluster?
					shiftNext, err = parseShortFlags(flags, opts, args, i, setFlagValue)
				} else {
					// TODO: the help flag is a pseudo flag.
				}

				nParsed = i + 1 - offset
				return
			}

			// match shorthand, here we don't check the flag length
			// as it may contain multiple shorthand flags (e.g. `-vvv`)
			shiftNext, err = parseShortFlags(flags, opts, args, i, setFlagValue)
		}

		if err != nil {
			if opts == nil || opts.HandleParseError == nil { // no error handler
				nParsed = i + 1 - offset
				return
			}

			err = opts.HandleParseError(opts, args, i, err)
			if err != nil {
				nParsed = i + 1 - offset
				return
			}

			// error ignored
		}

		if shiftNext {
			i++
		}
	}

	nParsed = i - offset

	return
}

func parseLongFlag(
	flags FlagFinder,
	opts *ParseOptions,
	args []string, // args[i] is the long flag with dash prefix
	i int,
	set bool,
) (shiftNext bool, err error) {
	name, value, hasValue := strings.Cut(args[i][2:], "=")

	// name MUST not be empty
	if len(name) == 0 {
		// defensive check, should never happen as the caller always checks the size.
		panic("unexpected empty arg name")
	}

	f, ok := flags.FindFlag(name)
	if !ok {
		return false, &ErrFlagUndefined{
			Name: name,
			At:   i,
		}
	}

	if hasValue {
		// --foo=bar case
		err = f.Decode(opts, name, value, set)
		if err != nil {
			return false, &ErrFlagValueInvalid{
				Name:    name,
				Value:   value,
				NameAt:  i,
				ValueAt: i,
			}
		}

		return false, nil
	}

	// --foo case
	if i == len(args)-1 || args[i+1] == "--" /* never use standalone dash as value */ {
		// when not having arg value, it MUST be an implicit arg
		goto TryImplied
	}

	value = args[i+1]
	if err = f.Decode(opts, name, value, false /* set = false for checking */); err == nil {
		// can consume next arg

		if strings.HasPrefix(value, "-") {
			if _, ok = f.ImplyValue(); ok {
				return true, &ErrAmbiguousArgs{
					Name:  name,
					Value: value,
					At:    i,
				}
			}
		}

		if set {
			return true, f.Decode(opts, name, value, true)
		}

		return true, nil
	}

	// cannot consume next arg, try implied value.
TryImplied:
	if value, ok = f.ImplyValue(); ok {
		err = f.Decode(opts, name, value, set)
		if err != nil {
			return false, &ErrFlagValueInvalid{
				Name:    name,
				Value:   value,
				NameAt:  i,
				ValueAt: -1,
				Reason:  err,
			}
		}

		return false, nil
	}

	// no implied value
	if i == len(args)-1 || args[i+1] == "--" {
		return false, &ErrFlagValueMissing{Name: name, At: i}
	}

	return false, &ErrFlagValueInvalid{
		Name:    name,
		Value:   args[i+1],
		NameAt:  i,
		ValueAt: i + 1,
		Reason:  err,
	}
}

func parseShortFlags(
	flags FlagFinder,
	opts *ParseOptions,
	args []string, // args[i] is the shorthand (cluster) with hyphen prefix
	i int,
	set bool,
) (shiftNext bool, err error) {
	var (
		offset   int
		width    int
		sz       int
		f        Flag
		name     string
		value    string
		hasValue bool
		ok       bool
	)

	s, value, hasValue := strings.Cut(args[i][1:], "=")

	for sz = len(s); offset < sz; {
		_, width = utf8.DecodeRuneInString(s[offset:])
		name = s[offset : offset+width]
		f, ok = flags.FindFlag(name)
		if !ok {
			return false, &ErrFlagUndefined{
				Name: name,
				At:   i,
			}
		}

		offset += width

		if offset == sz { // reaching the last shorthand
			if hasValue {
				err = f.Decode(opts, name, value, set)
				if err != nil {
					return false, &ErrFlagValueInvalid{
						Name:    name,
						Value:   value,
						NameAt:  i,
						ValueAt: i,
						Reason:  err,
					}
				}

				return false, nil
			}

			if i == len(args)-1 || args[i+1] == "--" /* never use standalone dash as value */ {
				goto TryImplied
			}

			value = args[i+1]
			if err = f.Decode(opts, name, value, false /* set = false for checking */); err == nil {
				// can consume the next arg.

				if strings.HasPrefix(value, "-") {
					if _, ok = f.ImplyValue(); ok {
						return true, &ErrAmbiguousArgs{
							Name:  name,
							Value: value,
							At:    i,
						}
					}
				}

				if set {
					err = f.Decode(opts, name, value, true)
					if err != nil {
						return true, &ErrFlagValueInvalid{
							Name:    name,
							Value:   value,
							NameAt:  i,
							ValueAt: i + 1,
						}
					}

					return true, nil
				}

				return true, nil
			}

			// cannot consume the next arg, try implied value
		TryImplied:
			if value, ok = f.ImplyValue(); ok {
				err = f.Decode(opts, name, value, set)
				if err != nil {
					return false, &ErrFlagValueInvalid{
						Name:    name,
						Value:   value,
						NameAt:  i,
						ValueAt: -1,
						Reason:  err,
					}
				}

				return false, nil
			}

			// bad

			if i == len(args)-1 || args[i+1] == "--" {
				return false, &ErrFlagValueMissing{Name: name, At: i}
			}

			return false, &ErrFlagValueInvalid{
				Name:    name,
				Value:   args[i+1],
				NameAt:  i,
				ValueAt: i + 1,
			}
		}

		// not the last shorthand
		//
		// flags in between prefer implicit value.
		if impliedValue, ok := f.ImplyValue(); ok {
			if err := f.Decode(opts, name, impliedValue, set); err != nil {
				return false, err
			}

			continue
		}

		// flag not having implicit value

		// treate remaining value in s as arg value
		// e.g. given `-t` == `--type`, then `-tfile` == `--type file`
		//
		// but reject `-abcd=file` because we are not `d`, and assigning
		// explicit value to `c` is not intuitive.
		if hasValue {
			return false, &ErrShorthandOfExplicitFlagInMiddle{
				Shorthand:        name,
				ShorthandCluster: s,
				Value:            value,
			}
		}

		err = f.Decode(opts, name, s[offset:], set)
		if err != nil {
			return false, &ErrFlagValueInvalid{
				Name:    "",
				Value:   s[offset:],
				NameAt:  i,
				ValueAt: i,
				Reason:  err,
			}
		}
		return false, nil
	}

	return false, nil
}
