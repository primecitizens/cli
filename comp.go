// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"strings"
	"unicode/utf8"
)

// CompState used by CompTask to manage completion values.
type CompState uint32

const (
	// CompStateFailed to NOT use any completion.
	CompStateFailed CompState = 1 << iota

	CompStateDone // to NOT add further CompItems.

	// CompItem content flags

	CompStateHasFlagNames
	CompStateHasFlagValues
	CompStateHasSubcmds
	CompStateHasFiles
	CompStateHasDirs

	// shell options

	CompStateOptionNospace
	CompStateOptionNosort

	testCompStateCmdCompletionAdded
	testCompStateFlagCompletionAdded
)

func (c CompState) Failed() bool        { return c&CompStateFailed != 0 }
func (c CompState) Done() bool          { return c&CompStateDone != 0 }
func (c CompState) HasFlagNames() bool  { return c&CompStateHasFlagNames != 0 }
func (c CompState) HasFlagValues() bool { return c&CompStateHasFlagValues != 0 }
func (c CompState) HasSubcmds() bool    { return c&CompStateHasSubcmds != 0 }
func (c CompState) HasFiles() bool      { return c&CompStateHasFiles != 0 }
func (c CompState) HasDirs() bool       { return c&CompStateHasDirs != 0 }
func (c CompState) OptionNospace() bool { return c&CompStateOptionNospace != 0 }
func (c CompState) OptionNosort() bool  { return c&CompStateOptionNosort != 0 }

// CompKind gives CompItem type information.
type CompKind uint8

const (
	CompKindText CompKind = iota
	CompKindFlagName
	CompKindFlagValue
	CompKindFiles
	CompKindDirs
)

// A CompItem is a completion suggestion.
type CompItem struct {
	// Value is the suggested text arg (e.g. foo, --bar).
	//
	// When Kind is one of [Files, Dirs], it is the glob pattern
	Value string

	// Description is the help text of the value.
	Description string

	// Kind marks the completion kind.
	Kind CompKind
}

// CompAction defines the interface for a completion action.
type CompAction interface {
	// Suggest adds possible CompItems according to CompTask.
	Suggest(tsk *CompTask) (added int, state CompState)
}

// CompActionDisable marks the CompTask failed, so it can be used to disable
// the default completion behavior.
type CompActionDisable struct{}

// Suggest implements [CompAction].
func (CompActionDisable) Suggest(tsk *CompTask) (int, CompState) {
	tsk.Fail()
	return 0, CompStateFailed
}

// CompActionStatic adds its predefined suggestions for completion request.
type CompActionStatic struct {
	Suggestions []CompItem
	Want, State CompState
}

// Suggest implements [CompAction].
func (s *CompActionStatic) Suggest(tsk *CompTask) (int, CompState) {
	if s.Want != 0 && tsk.Want()&s.Want == 0 {
		return 0, 0
	}

	return tsk.AddMatched(false, s.Suggestions...), s.State
}

// CompActionFunc wraps a function as Completion implementation.
type CompActionFunc func(tsk *CompTask) (int, CompState)

// Suggest implements [CompAction].
func (fn CompActionFunc) Suggest(tsk *CompTask) (int, CompState) {
	return fn(tsk)
}

// CompActionDirs adds a CompItem to request dir completion.
type CompActionDirs struct{}

// Suggest implements [CompAction].
func (CompActionDirs) Suggest(tsk *CompTask) (int, CompState) {
	return tsk.AddDirs(false), 0
}

// CompActionFiles adds a CompItem to request file completion.
type CompActionFiles struct{}

// Suggest implements [CompAction].
func (CompActionFiles) Suggest(tsk *CompTask) (int, CompState) {
	return tsk.AddFiles(false), 0
}

// CompTask represents a completion tsk.
type CompTask struct {
	debug  io.Writer
	result []CompItem

	// fields below are set by .Init()

	// ExecutablePath is the path to the executable which requested this
	// completion task.
	ExecutablePath string

	// Args are all args present on command-line, including args after the
	// one requested the completion.
	Args []string

	// At is the position of the arg to complete.
	//
	// when At == len(Args), the task is to suggest the next arg.
	// when 0 <= At <= len(Args)-1, the task is to complete the last partial arg.
	// when < len(Args)-1, the task is to complete previous partial arg.
	At int

	// ToComplete is the arg value to be compeleted.
	//
	// When completing a flag value, it is the value part of the arg, which
	// means if command-line input triggering the completion was
	// `--foo=s<tab>` and `foo` can be found in that context, then this field
	// will get `s`.
	ToComplete string

	// Route is the Cmd route up to the arg to complete.
	Route Route

	// PosArgs are all positional args found before the arg to complete.
	//
	// Depending on the position of the arg to complete, the Args field may
	// contain positional args, but this field can get nothing.
	PosArgs []string

	// DashArgs are all args found after the first dash but before the arg
	// to complete.
	//
	// If the arg to complete is before the dash, the Args may contain dash
	// but this field will get nothing.
	DashArgs []string

	// FlagMissingValue is the Flag right before the arg to complete and is
	// missing the value in arg to complete.
	FlagMissingValue Flag
	FlagValuePrefix  string

	state CompState
	want  CompState
}

// RawToComplete returns the unprocessed arg value to complete.
func (tsk *CompTask) RawToComplete() string {
	if tsk.At < 0 || tsk.At >= len(tsk.Args) {
		return ""
	}

	return tsk.Args[tsk.At]
}

// SetDebugOutput sets the debug output used by the Debug method.
func (tsk *CompTask) SetDebugOutput(out io.Writer) {
	tsk.debug = out
}

// Want returns the CompState containing CompStateHasXxx bits indicating
// the expected completion result set.
//
// These bits are set by Init() and are used by AddDefault().
func (tsk *CompTask) Want() CompState { return tsk.want }

// Fail marks this task as failed, and should not produce any completion
// at all.
func (tsk *CompTask) Fail() { tsk.state |= CompStateFailed }

// Done marks this task as finished, no CompItem can be added without
// setting force=false.
func (tsk *CompTask) Done() { tsk.state |= CompStateDone }

// State returns the current CompState of the task.
func (tsk *CompTask) State() CompState { return tsk.state }

// Debug writes messages to the debug output.
func (tsk *CompTask) Debug(msgs ...string) {
	if tsk.debug == nil {
		return
	}

	_, _ = wstr(tsk.debug, "[app]")
	for _, msg := range msgs {
		_, _ = wstr(tsk.debug, " ")
		_, _ = wstr(tsk.debug, msg)
	}
	_, _ = wstr(tsk.debug, "\n")
}

// Nth returns the i-th CompItem has been added.
func (tsk *CompTask) Nth(i int) (CompItem, bool) {
	if i < len(tsk.result) && i >= 0 {
		return tsk.result[i], true
	}

	return CompItem{}, false
}

// Init initializes the CompTask with command-line options.
//
// If pos is in range [0, len(args)), args[pos] is the arg to complete.
//
// If the args slice is not empty, args[0] is expected to be the executable
// path.
func (tsk *CompTask) Init(root *Cmd, opts *CmdOptions, at int, args ...string) {
	if len(args) > 0 {
		// shift 1 (cannot be completing the executable path)
		tsk.ExecutablePath, args = args[0], args[1:]
		if at >= 0 {
			at--
		}
		tsk.Args, tsk.At = args, at
	} else {
		// TODO(TBD): error on no arg?
		tsk.ExecutablePath, tsk.Args, tsk.At = "", nil, 0
	}

	if at < len(args) && at >= 0 {
		tsk.ToComplete = args[at]
	} else {
		tsk.ToComplete = ""
	}

	end := at
	if end > len(args) || end < 0 {
		end = len(args)
	}

	var err error
	tsk.Route, tsk.PosArgs, tsk.DashArgs, err = root.ResolveTarget(opts, args[:end]...)
	if err != nil {
		switch e := err.(type) {
		case *ErrFlagValueMissing:
			if e.At == end-1 {
				// arg before toComplete is missing flag value => completing a flag value
				f, ok := tsk.Route.FindFlag(e.Name)
				if ok {
					tsk.want = CompStateHasFlagValues
					tsk.FlagMissingValue = f
					return
				}
			}
		}
	}

	switch toComplete := tsk.ToComplete; {
	case len(toComplete) == 0:
		tsk.want = CompStateHasFlagNames | CompStateHasSubcmds
	case !strings.HasPrefix(toComplete, "-"): // no hyphen prefix, cannot be a flag name
		tsk.want = CompStateHasSubcmds
	case /* has hyphen prefix && */ len(toComplete) > 1:
		pos := strings.IndexByte(toComplete, '=')
		if pos < 0 {
			// just in case there is sub-command name with hyphen prefix
			tsk.want = CompStateHasSubcmds | CompStateHasFlagNames
			break
		}

		// has value
		flag, value := toComplete[:pos], toComplete[pos+1:]
		switch {
		case flag[1] != '-': // can assume shorthand (maybe cluster)
			_, sz := utf8.DecodeLastRuneInString(flag)
			f, ok := tsk.Route.FindFlag(flag[len(flag)-sz:])
			if ok {
				tsk.FlagMissingValue = f
				tsk.FlagValuePrefix = toComplete[:pos+1]
				tsk.ToComplete = value
				tsk.want = CompStateHasFlagValues
			} else {
				tsk.want = CompStateHasSubcmds
			}
		case flag[1] == '-' && len(flag) > 2 && !IsShorthand(flag[2:]): // can assume long name
			flag = flag[2:]
			f, ok := tsk.Route.FindFlag(flag)
			if ok {
				tsk.FlagMissingValue = f
				tsk.FlagValuePrefix = toComplete[:pos+1]
				tsk.ToComplete = value
				tsk.want = CompStateHasFlagValues
			} else {
				tsk.want = CompStateHasSubcmds
			}
		default:
			tsk.want = CompStateHasSubcmds
		}
	default: // has hyphen only
		// just in case there is sub-command name with hyphen prefix
		tsk.want = CompStateHasSubcmds | CompStateHasFlagNames
	}
}

// Add force adds CompItems without checking prefix match
func (tsk *CompTask) Add(force bool, items ...CompItem) (added int) {
	if !force && (tsk.state&(CompStateFailed|CompStateDone) != 0) {
		return
	}

	tsk.result = append(tsk.result, items...)
	return len(items)
}

// AddMatched filters CompItems and only adds those with tsk.ToComplete prefix.
func (tsk *CompTask) AddMatched(force bool, items ...CompItem) (added int) {
	if !force && (tsk.state&(CompStateFailed|CompStateDone) != 0) {
		return
	}

	for i := range items {
		if strings.HasPrefix(items[i].Value, tsk.ToComplete) {
			added++
			tsk.result = append(tsk.result, items[i])
		}
	}

	return
}

// AddDefault adds CompItems indicated by argument parsing (Init).
func (tsk *CompTask) AddDefault() (added int) {
	if x := tsk.Route.Target().Completion; x != nil {
		var s CompState
		added, s = x.Suggest(tsk)
		tsk.state |= s
	}

	if tsk.Want().HasFlagValues() {
		added += tsk.AddFlagValues(false, tsk.FlagMissingValue, "", false)
	}

	if tsk.Want().HasFlagNames() {
		added += tsk.AddFlagNames(false, nil, true)
	}

	if tsk.Want().HasSubcmds() {
		added += tsk.AddSubcmds(false, nil, true)
	}

	return
}

// AddSubcmds adds sub-command names of cmd.
//
// If the argument `cmd` is nil, use the last Cmd in tsk.Route.
//
// Set descr to true to include description.
func (tsk *CompTask) AddSubcmds(force bool, cmd *Cmd, descr bool) (added int) {
	if !force && (tsk.state&(CompStateHasSubcmds|CompStateFailed|CompStateDone) != 0) {
		return
	}

	tsk.state |= CompStateHasSubcmds

	if cmd == nil {
		cmd = tsk.Route.Target()
	}

	for _, child := range cmd.Children {
		if child == nil || child.State.Hidden() {
			continue
		}

		var name string
		names, _, _ := strings.Cut(child.Pattern, " ")
		for len(names) != 0 {
			name, names, _ = strings.Cut(names, "|")
			if !strings.HasPrefix(name, tsk.ToComplete) &&
				!isSimilar(name, tsk.ToComplete, true) {
				continue
			}

			item := CompItem{
				Value: name,
			}

			if descr {
				item.Description = child.BriefUsage
			}

			added += tsk.Add(force, item)
		}
	}

	return
}

// AddFlagNames adds flag names, it expects tsk.ToComplete either being an
// empty string or containing a flag name prefix (`-`, `--`).
//
// If the argument `flags` is nil, use the target command's flags (tsk.Route).
func (tsk *CompTask) AddFlagNames(force bool, flags FlagIndexer, descr bool) (added int) {
	if !force && (tsk.state&(CompStateHasFlagNames|CompStateFailed|CompStateDone) != 0) {
		return
	}

	tsk.state |= CompStateHasFlagNames

	if flags == nil {
		flags = noescape(&tsk.Route)
	}

	switch toComplete := tsk.ToComplete; {
	case len(toComplete) == 0 || toComplete == "-": // all flags not hidden can be added
		for i := 0; ; i++ {
			info, ok := flags.NthFlag(i)
			if !ok {
				break
			}

			_, f, ok := FindFlag(flags, info.Name, info.Shorthand)
			if !ok || f.State().Hidden() {
				continue
			}

			if len(info.Name) != 0 && !IsShorthand(info.Name) {
				item := CompItem{
					Value: info.Name,
					Kind:  CompKindFlagName,
				}

				if descr {
					item.Description = f.Usage()
				}

				added += tsk.Add(force, item)
			}

			if IsShorthand(info.Shorthand) {
				item := CompItem{
					Value: info.Shorthand,
					Kind:  CompKindFlagName,
				}

				if descr {
					item.Description = f.Usage()
				}

				added += tsk.Add(force, item)
			}
		}
	case strings.HasPrefix(toComplete, "--"): // long flags not hidden may be added
		for i := 0; ; i++ {
			info, ok := flags.NthFlag(i)
			if !ok {
				break
			}

			if len(info.Name) == 0 ||
				IsShorthand(info.Name) ||
				!strings.HasPrefix(info.Name, toComplete[2:]) {
				if !isSimilar(info.Name, toComplete[2:], true) {
					continue
				}
			}

			_, f, ok := FindFlag(flags, info.Name, info.Shorthand)
			if !ok || f.State().Hidden() {
				continue
			}

			item := CompItem{
				Value: info.Name,
				Kind:  CompKindFlagName,
			}

			if descr {
				item.Description = f.Usage()
			}

			added += tsk.Add(force, item)
		}
	case strings.HasPrefix(toComplete, "-"):
		// has hyphen prefix but not dash prefix, and also not just a single
		// hyphen.
		// Thus only flag shorthands may be added, but since flag shorthands
		// are always one rune in length, so it is to confirm the existence
		// of the flag.
		for i := 0; ; i++ {
			info, ok := flags.NthFlag(i)
			if !ok {
				break
			}

			shorthand := info.Shorthand
			if !IsShorthand(shorthand) {
				if IsShorthand(info.Name) {
					shorthand = info.Name
				} else {
					continue
				}
			}

			_, f, ok := FindFlag(flags, info.Name, info.Shorthand)
			if !ok || f.State().Hidden() || !strings.Contains(tsk.ToComplete, shorthand) {
				continue
			}

			item := CompItem{
				Value: shorthand,
				Kind:  CompKindFlagName,
			}

			if descr {
				item.Description = f.Usage()
			}

			added += tsk.Add(force, item)
		}
	}

	return
}

// AddFlagValues adds matched values from the specified flag.
//
// It retrieves completion suggestions by trying following methods in order:
//   - cast flag as CompAction.
//   - cast flag.Extra() as CompAction.
//
// If addDefaults is true:
//   - add value returned by Flag.Default() if matched.
//   - deduce default value by flag State() unchange and HasValue() true.
func (tsk *CompTask) AddFlagValues(force bool, flag Flag, def string, addDefaults bool) (added int) {
	if flag == nil || (!force && (tsk.state&(CompStateHasFlagValues|CompStateFailed|CompStateDone) != 0)) {
		return
	}

	tsk.state |= CompStateHasFlagValues

	comp, ok := flag.(CompAction)
	if !ok {
		comp, _ = flag.Extra().(CompAction)
	}
	if comp != nil {
		var s CompState
		added, s = comp.Suggest(tsk)
		tsk.state |= s
	}

	if !addDefaults {
		return
	}

	if len(def) == 0 {
		goto DeduceDefault
	}

	if def[0] == '[' && def[len(def)-1] == ']' {
		var (
			ent   string
			state = flag.State()
		)

		def = def[1 : len(def)-1]
		for i := 0; len(def) > 0; i++ {
			ent, def, _ = strings.Cut(def, ", ")

			// TODO: check all tsk.Args to find out start from which i-th element.
			if state.ValueChanged() && i == 0 {
				continue
			}

			added += tsk.AddMatched(force, CompItem{
				Value:       ent,
				Description: "default value",
				Kind:        CompKindFlagValue,
			})
		}
	} else {
		if state := flag.State(); !state.ValueChanged() { // no previous value
			added += tsk.AddMatched(force, CompItem{
				Value:       def,
				Description: "default value",
				Kind:        CompKindFlagValue,
			})
		}
	}

DeduceDefault:
	if !flag.State().ValueChanged() && flag.HasValue() {
		var sb strings.Builder
		psb := noescape(&sb)
		if _, err := flag.PrintValue(psb); err == nil {
			added += tsk.AddMatched(force, CompItem{
				Value:       sb.String(),
				Description: "default value",
				Kind:        CompKindFlagValue,
			})
		}
	}

	return
}

// AddFiles adds requests to match filesystem files.
func (tsk *CompTask) AddFiles(force bool, globs ...string) (added int) {
	if !force && (tsk.state&(CompStateHasFiles|CompStateFailed|CompStateDone) != 0) {
		return
	}

	tsk.state |= CompStateHasFiles

	for _, glob := range globs {
		if len(glob) == 0 {
			continue
		}

		added += tsk.Add(force, CompItem{
			Value: glob,
			Kind:  CompKindFiles,
		})
	}

	if added == 0 {
		added += tsk.Add(force, CompItem{
			Kind: CompKindFiles,
		})
	}

	return
}

// AddDirs adds requests to match filesystem dirs.
func (tsk *CompTask) AddDirs(force bool, globs ...string) (added int) {
	if !force && (tsk.state&(CompStateHasDirs|CompStateFailed|CompStateDone) != 0) {
		return
	}

	tsk.state |= CompStateHasDirs

	for _, glob := range globs {
		if len(glob) == 0 {
			continue
		}

		added += tsk.Add(force, CompItem{
			Value: glob,
			Kind:  CompKindDirs,
		})
	}

	if added == 0 {
		added += tsk.Add(force, CompItem{
			Kind: CompKindDirs,
		})
	}

	return
}
