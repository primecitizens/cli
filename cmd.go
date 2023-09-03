// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"strings"
)

type (
	// ArgErrorHandleFunc for args error.
	//
	// route is the Cmd route from the root to the last known good Cmd.
	//
	// args are the same args passed to Cmd.ResolveTarget
	//
	// when badArgAt >= 0, it is where argErr happened, and args[badArgAt]
	// is the bad arg.
	//
	// Return nil to ignore the error.
	ArgErrorHandleFunc = func(
		opts *CmdOptions, route Route, args []string, badArgAt int, argErr error,
	) error

	// HelpHandleFunc for handling help requests.
	//
	// when helpArgAt < 0, this function was called as the fallback of
	// ArgErrorHandleFunc, and the argErr is discarded.
	//
	// when helpArgAt >= 0, args[helpArgAt] is the arg initiated the help
	// request, in this case return nil error will be replaced with
	// ErrHelpRequestHandled{}.
	HelpHandleFunc = func(
		opts *CmdOptions, route Route, args []string, helpArgAt int,
	) error
)

type (
	// PreRunFunc
	//
	// route is the full Cmd route leading to the target, and route[prerunAt]
	// is the Cmd that owns this PreRun.
	//
	// posArgs and dashArgs are meant for the target Cmd.
	//
	// Return an error to cancel all subsequent PreRun calls and the error
	// will be returned by Cmd.Exec.
	PreRunFunc = func(opts *CmdOptions, route Route, prerunAt int, posArgs, dashArgs []string) error

	// RunFunc
	//
	// function parameters' definition is the same as PreRun.
	RunFunc = func(opts *CmdOptions, route Route, posArgs, dashArgs []string) error

	// PostRunFunc
	//
	// route is the full Cmd route leading to the target, and route[postrunAt]
	// is the Cmd that owns this PostRun.
	//
	// runErr is the error returned by the Run func and can only be non-nil
	// when passed to the first PostRun (not necessarily the one from the
	// target Cmd).
	//
	// Returning an error cancels all subsequent PostRun call and the error
	// will be returned by Cmd.Exec.
	PostRunFunc = func(opts *CmdOptions, route Route, postrunAt int, runErr error) error
)

// CmdOptions are options for Cmd execution.
type CmdOptions struct {
	ParseOptions *ParseOptions

	// RouteBuf is the buffer for building the Cmd route.
	RouteBuf Route

	// Stdin is the stdin of the cmd.
	//
	// Defaults to nil.
	Stdin io.Reader

	// Stdout is the stdout of the cmd.
	//
	// Defaults to nil.
	Stdout io.Writer

	// Stderr is the stderr of the cmd.
	//
	// Defaults to nil.
	Stderr io.Writer

	// HandleArgError is the function get called to handle errors happened
	// during target resolving.
	HandleArgError ArgErrorHandleFunc

	// HandleHelpRequest is the fallback help request handle func.
	//
	// In a Cmd.Exec call, if both target Cmd.Help func and
	// CmdOptions.HandleHelpRequest are nil, no help will be provided.
	HandleHelpRequest HelpHandleFunc

	// Extra custom data.
	Extra any

	// SkipPostRun skips the Cmd.PostRun when set to true.
	SkipPostRun bool

	// DoNotSetFlags skips setting flag values.
	DoNotSetFlags bool
}

// PickStdin returns def if c.Stdin is nil.
func (c *CmdOptions) PickStdin(def ...io.Reader) io.Reader {
	if c != nil && c.Stdin != nil {
		return c.Stdin
	}

	for _, r := range def {
		if r != nil {
			return r
		}
	}

	return nil
}

// PickStdout returns def if c.Stdout is nil.
func (c *CmdOptions) PickStdout(def ...io.Writer) io.Writer {
	if c != nil && c.Stdout != nil {
		return c.Stdout
	}

	for _, w := range def {
		if w != nil {
			return w
		}
	}

	return nil
}

// PickStderr returns def if c.Stderr is nil.
func (c *CmdOptions) PickStderr(def ...io.Writer) io.Writer {
	if c != nil && c.Stderr != nil {
		return c.Stderr
	}

	for _, w := range def {
		if w != nil {
			return w
		}
	}

	return nil
}

type CmdState uint32

const (
	// CmdStateHidden hides the cmd from completion when set.
	CmdStateHidden CmdState = 1 << iota
	// CmdStatePreRunOnce to require the PreRun only gets called once.
	CmdStatePreRunOnce
	CmdStatePreRunCalled

	// CmdStatePostRunOnce to require the PostRun only gets called once.
	CmdStatePostRunOnce
	CmdStatePostRunCalled
)

func (s CmdState) Hidden() bool        { return s&CmdStateHidden != 0 }
func (s CmdState) PreRunOnce() bool    { return s&CmdStatePreRunOnce != 0 }
func (s CmdState) PreRunCalled() bool  { return s&CmdStatePreRunCalled != 0 }
func (s CmdState) PostRunOnce() bool   { return s&CmdStatePostRunOnce != 0 }
func (s CmdState) PostRunCalled() bool { return s&CmdStatePostRunCalled != 0 }

// AnyMaybeHelperTerminal is an alias of `any` and indicates
// some component will try to cast the value as a HelperTerminal.
type AnyMaybeHelperTerminal = any

// A Cmd represents a command.
//
// TODO: define Command interface.
type Cmd struct {
	// Pattern is supposed to be a one-line usage pattern of the command.
	//
	// Text before the first space is used for matching args in order to pick
	// this Cmd as sub-command, multiple names can be provided by joining with
	// pipe sign ('|').
	//
	// As of the text after the first space, here is the recommended syntax:
	//
	//	- `[ ]` to define an optional argument.
	//	- `...` to allow multiple values for the previous arg.
	//	- `|` to provide mutually exclusive options.
	//	- `{ }` to define a group of mutually exclusive args.
	//
	// Example (where foo is the command name):
	//
	//	foo|f [-F file | -D dir]... [-f {text|audio}] profile
	//
	// In the above example, `foo` is the command name and `f` is its alias.
	Pattern string

	// BriefUsage introduces the command briefly.
	BriefUsage string

	// Flags are flags accessible from both this Cmd and all its children.
	Flags FlagFinderMaybeIter

	// LocalFlags are flags only accessible from this Cmd.
	//
	// It is preferred to Flags for flag looking up.
	LocalFlags FlagFinderMaybeIter

	// FlagRule enforces certain rule to flags.
	FlagRule Rule

	// PreRun hook, see the type alias definition for parameter details.
	//
	// In Cmd.Exec, it is called from the root Cmd down to the target Cmd
	// before calling Run (even if the target Cmd's Run may be nil).
	PreRun PreRunFunc

	// Run, see the type alias definition for parameter details.
	//
	// In Cmd.Exec, it is called only when the owner Cmd is the target Cmd.
	Run RunFunc

	// PostRun hook, see the type alias definition for parameter details.
	//
	// In Cmd.Exec, it is called from that target Cmd up to the root Cmd after
	// the Run function returned.
	PostRun PostRunFunc

	// Help provides command specific help request handling.
	//
	// If not nil, it is called on help request for this command, otherwise
	// fallback to CmdOptions.HandleHelpRequest.
	Help HelpHandleFunc

	// Completion is the shell completion helper to suggest args for the
	// command.
	Completion CompAction

	// Extra stores application specific custom data.
	Extra AnyMaybeHelperTerminal

	// Children are sub-commands beloning to this Cmd.
	Children []*Cmd

	// State is Cmd's current state.
	State CmdState
}

// Name returns the first name in Pattern of this Cmd.
func (c *Cmd) Name() (name string) {
	name, _, _ = strings.Cut(c.Pattern, " ")
	name, _, _ = strings.Cut(name, "|")
	return
}

// Is returns true if s is considered a name of this Cmd.
func (c *Cmd) Is(s string) bool {
	var name string
	names, _, _ := strings.Cut(c.Pattern, " ")
	for len(names) != 0 {
		name, names, _ = strings.Cut(names, "|")
		if s == name {
			return true
		}
	}

	return false
}

func pick(fns ...HelpHandleFunc) HelpHandleFunc {
	for _, fn := range fns {
		if fn != nil {
			return fn
		}
	}

	return nil
}

// ResolveTarget walks the Cmd tree from c to the target Cmd by parsing args.
//
// On a successful return, the `route` leads to the target Cmd with this Cmd
// being the first entry.
func (c *Cmd) ResolveTarget(opts *CmdOptions, args ...string) (
	route Route, posArgs, dashArgs []string, err error,
) {
	var (
		popts     *ParseOptions
		nParsed   int
		offset    int
		posDash   int
		helpArgAt int

		fallbackHelp HelpHandleFunc
		handleArgErr ArgErrorHandleFunc
		setFlagValue bool = true
	)

	if opts != nil {
		popts = opts.ParseOptions
		handleArgErr = opts.HandleArgError
		route = opts.RouteBuf
		setFlagValue = !opts.DoNotSetFlags
		fallbackHelp = opts.HandleHelpRequest

		if popts != nil {
			posArgs = popts.PosArgsBuf
		}
	}

	helpRequested := func() bool {
		if helpArgAt < 0 {
			return false
		}

		if handleHelp := pick(c.Help, fallbackHelp); handleHelp != nil {
			if err = handleHelp(opts, route, args, helpArgAt); err == nil {
				err = ErrHelpHandled{}
			}
		} else {
			err = &ErrHelpPending{
				HelpArg: args[helpArgAt],
				At:      helpArgAt,
			}
		}

		return true
	}

	errReturn := func() bool {
		if err == nil {
			return false
		}

		if handleArgErr == nil {
			if handleHelp := pick(c.Help, fallbackHelp); handleHelp != nil {
				if helpArgAt >= 0 {
					_ = handleHelp(opts, route, args, helpArgAt)
				} else {
					_ = handleHelp(opts, route, args, -1)
				}
			}

			return true
		}

		err = handleArgErr(opts, route, args, offset+nParsed-1, err)
		if err != nil {
			return true
		}

		return false
	}

	protue := noescape(&route)
	for route = route.Push(c); len(c.Children) != 0; route = route.Push(c) {
		var foundPosArgs bool
		nParsed, _, foundPosArgs, _, helpArgAt, err = ParseFlagsLowLevel(
			args, protue, popts,
			offset,
			false,        // appendPosArgs
			true,         // stopAtFirstPosArg
			setFlagValue, // setFlagValue
			nil,          // posArgsBuf
		)
		if errReturn() {
			return
		}

		offset += nParsed
		if !foundPosArgs || offset >= len(args) {
			// exhausted all args
			if helpRequested() {
				return
			}

			break
		}

		expectedName := args[offset]
		noSuchCmd := true
		for _, child := range c.Children {
			if child.Is(expectedName) {
				c = child
				noSuchCmd = false
				break
			}
		}

		if noSuchCmd {
			if helpRequested() {
				return
			}

			// this arg is a positional arg for current Cmd
			break
		}

		if helpRequested() {
			return
		}

		offset++
	}

	nParsed, posDash, _, posArgs, helpArgAt, err = ParseFlagsLowLevel(
		args, protue, popts,
		offset,
		true,         // appendPosArgs
		false,        // stopAtFirstPosArg
		setFlagValue, // setFlagValue
		posArgs,      // posArgsBuf
	)
	if errReturn() {
		return
	}

	if helpRequested() {
		return
	}

	if posDash >= 0 {
		dashArgs = args[posDash+1:]
	}

	return
}

// Exec tries to find and run the Cmd with longest matching Cmd.Pattern in args.
//
// When called, this Cmd assumes itself as the root command.
func (c *Cmd) Exec(opts *CmdOptions, args ...string) (err error) {
	route, posArgs, dashArgs, err := c.ResolveTarget(opts, args...)
	if err != nil {
		return
	}

	var popts *ParseOptions
	if opts != nil && opts.ParseOptions != nil {
		popts = opts.ParseOptions
	}

	proute := noescape(&route)

	var i int
	for i, c = range route {
		err = tryAssignFlagsDefaultValue(c.LocalFlags, popts)
		if err != nil {
			return
		}

		err = tryAssignFlagsDefaultValue(c.Flags, popts)
		if err != nil {
			return
		}

		if rule := c.FlagRule; rule != nil {
			for k := 0; ; k++ {
				violation, ok := rule.NthEx(proute, k)
				if !ok {
					break
				}

				return &FlagViolation{
					Key:    violation.Key,
					Reason: violation.Reason,
				}
			}
		}

		if c.PreRun == nil || (c.State.PreRunOnce() && c.State.PreRunCalled()) {
			continue
		}

		c.State |= CmdStatePreRunCalled
		err = c.PreRun(opts, route, i, posArgs, dashArgs)
		if err != nil {
			// TODO: default handling of PreRun error?
			return
		}
	}

	if c.Run == nil {
		err = &ErrCmdNotRunnable{
			Name: c.Name(),
		}

		if opts != nil {
			if opts.HandleArgError != nil {
				err = opts.HandleArgError(opts, route, args, -1, err)
			} else if help := pick(c.Help, opts.HandleHelpRequest); help != nil {
				_ = help(opts, route, args, -1)
			}
		} else {
			if c.Help != nil {
				_ = c.Help(opts, route, args, -1)
			}
		}

		return
	}

	err = c.Run(opts, route, posArgs, dashArgs)
	if opts != nil && opts.SkipPostRun {
		return
	}

	for i = len(route) - 1; i >= 0; i-- {
		c = route[i]
		if c.PostRun == nil || (c.State.PostRunOnce() && c.State.PostRunCalled()) {
			continue
		}

		c.State |= CmdStatePostRunCalled
		err = c.PostRun(opts, route, i, err)
		if err != nil {
			// TODO: default handling of PostRun error?
			return
		}
	}

	return
}

func tryAssignFlagsDefaultValue(flags FlagFinderMaybeIter, opts *ParseOptions) error {
	if flags == nil {
		return nil
	}

	indexer, ok := flags.(FlagIndexer)
	if !ok || indexer == nil {
		return nil
	}

	return AssignFlagsDefaultValue(indexer, opts)
}

// AssignFlagsDefaultValue iterates through all flags and call Flag.Decode on
// flags with default value (indicated by FlagInfo.DefaultValue) but without
// FlagStateValueChanged set (indicated by both FlagInfo.State and
// Flag.State()).
func AssignFlagsDefaultValue(flags FlagIndexer, opts *ParseOptions) (err error) {
	for i := 0; ; i++ {
		info, ok := flags.NthFlag(i)
		if !ok {
			break
		}

		if len(info.DefaultValue) == 0 || info.State.ValueChanged() {
			continue
		}

		name, flag, ok := FindFlag(flags, info.Name, info.Shorthand)
		if !ok {
			name = info.Name
			if len(name) == 0 {
				name = info.Shorthand
			}

			return &ErrFlagUndefined{
				Name: name,
				At:   -1,
			}
		}

		if flag.State().ValueChanged() { // defensive check
			continue
		}

		if def := info.DefaultValue; def[0] == '[' && def[len(def)-1] == ']' {
			var ent string
			for def = def[1 : len(def)-1]; len(def) > 0; {
				ent, def, _ = strings.Cut(def, ", ")
				err = flag.Decode(opts, name, ent, true)
				if err != nil {
					return
				}
			}
		} else {
			err = flag.Decode(opts, name, def, true)
			if err != nil {
				return
			}
		}
	}

	return nil
}
