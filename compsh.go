// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	_ "embed" // for go:embed
)

// CompCmdShells is wraps all supported shell completion commands:
//
//   - bash {,complete}
//   - zsh {,complete}
//   - pwsh {,complete}
//
// To use it, the return value of (&CompCmdShells{}).Setup(...) should
// be a direct child to your application's root command.
type CompCmdShells struct {
	self Cmd

	bash CompCmdBash
	zsh  CompCmdZsh
	pwsh CompCmdPwsh

	shells [3]*Cmd

	// opComp shared by all shell sub-commands
	opComp CompCmdOpComplete
}

// Setup the shell completion command.
//
// name will be set to "completion" if it is empty.
//
// defaultTimeout will be 5s, if it is negative, set it to 0 to disable
// timeout.
func (cc *CompCmdShells) Setup(name string, defaultTimeout time.Duration, hide bool) *Cmd {
	if defaultTimeout < 0 {
		defaultTimeout = 5 * time.Second
	}

	if len(name) == 0 {
		name = "completion"
	}

	var state CmdState
	if hide {
		state |= CmdStateHidden
	}

	*cc = CompCmdShells{
		self: Cmd{
			Pattern:    name,
			State:      state,
			BriefUsage: "shell completion",
			Children:   cc.shells[:],
		},
	}
	opCompCmd := cc.opComp.Setup(defaultTimeout)

	cc.shells = [3]*Cmd{
		cc.bash.Setup(opCompCmd),
		cc.zsh.Setup(opCompCmd),
		cc.pwsh.Setup(opCompCmd),
	}

	return &cc.self
}

type opCompContext struct {
	tsk   CompTask
	copts CmdOptions
	popts ParseOptions
}

// CompCmdOpComplete wraps the `complete` operation sub-command for type
// CompCmd{Bash, Zsh, Pwsh}.
//
// It prepares the CompTask and invokes its parent Cmd to process the task.
type CompCmdOpComplete struct {
	// Timeout is the time limit to one completion request.
	Timeout DurationV

	// DebugFile is the file to write debug messages.
	DebugFile StringV

	// At is the position the cursor currently At.
	//
	// for zsh, it should be ($CURRENT - 1).
	// for bash, it should be $cword
	// for pwsh, it should be arg index base on $cursorPosition
	At UintV

	self     Cmd
	flagRule RuleAllOf

	ctx       opCompContext
	strBuf    [16]string
	resultBuf [16]CompItem
	routeBuf  [8]*Cmd
}

// Task returns the prepared CompTask.
func (cc *CompCmdOpComplete) Task() *CompTask {
	return &cc.ctx.tsk
}

// Setup returns the initialized *Cmd.
func (cc *CompCmdOpComplete) Setup(defaultTimeout time.Duration) *Cmd {
	*cc = CompCmdOpComplete{
		DebugFile: StringV{
			BriefUsage: "write internal debug messages to this file",
		},
		Timeout: DurationV{
			BriefUsage: "set the duration to wait for a completion task",
			Value:      defaultTimeout,
		},
		At: UintV{
			BriefUsage: "set arg index the cursor currently at",
		},
		strBuf:   [16]string{0: "at"},
		flagRule: RuleAllOf{Keys: cc.strBuf[:1]},
		ctx: opCompContext{
			tsk: CompTask{
				result: cc.resultBuf[:0:len(cc.resultBuf)],
			},
			copts: CmdOptions{
				ParseOptions: &cc.ctx.popts,
				RouteBuf:     cc.routeBuf[:0:len(cc.routeBuf)],
				Extra:        &cc.ctx.tsk,
			},
			popts: ParseOptions{
				PosArgsBuf: cc.strBuf[1:1:len(cc.strBuf)],
				HelpArgs:   cc.strBuf[:0], // disable help request
				Extra:      &cc.ctx.tsk,   // used in HandleParseError
			},
		},
		self: Cmd{
			Pattern:    "complete",
			BriefUsage: "Handle shell completion request",
			FlagRule:   &cc.flagRule,
			LocalFlags: cc,
			Run:        runOpComplete,
		},
	}

	return &cc.self
}

func runOpComplete(opts *CmdOptions, route Route, posArgs, dashArgs []string) error {
	var (
		file *os.File
		err  error
	)

	// allow panic on nil/incompatible target, as this cmd is misused.
	// panic will provide useful stacktrace.
	self := route.Target().LocalFlags.(*CompCmdOpComplete)
	if opts != nil {
		self.ctx.copts.Stdin = opts.Stdin
		self.ctx.copts.Stdout = opts.Stdout
		self.ctx.copts.Stderr = opts.Stderr
		self.ctx.copts.HandleArgError = opts.HandleArgError
		self.ctx.copts.HandleHelpRequest = opts.HandleHelpRequest
		self.ctx.copts.SkipPostRun = opts.SkipPostRun
		self.ctx.copts.DoNotSetFlags = opts.DoNotSetFlags

		if opts.ParseOptions != nil {
			// all other fields are specific to this command.
			self.ctx.copts.ParseOptions.StartTime = opts.ParseOptions.StartTime
		}
	}

	if debugFile := self.DebugFile.Value; len(debugFile) != 0 {
		file, err = os.OpenFile(debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			self.ctx.tsk.SetDebugOutput(opts.PickStderr(os.Stderr))
		} else {
			self.ctx.tsk.SetDebugOutput(file)
			defer file.Close()
		}
	}

	return route.Up().Target().Run(&self.ctx.copts, route, posArgs, dashArgs)
}

// NthFlag implements [FlagIter].
func (cc *CompCmdOpComplete) NthFlag(i int) (info FlagInfo, ok bool) {
	switch i {
	case 0:
		return FlagInfo{Name: "at"}, true
	case 1:
		return FlagInfo{Name: "timeout"}, true
	case 2:
		return FlagInfo{Name: "debug-file"}, true
	default:
		return
	}
}

// FindFlag implements [FlagFinder].
func (cc *CompCmdOpComplete) FindFlag(name string) (Flag, bool) {
	switch name {
	case "at":
		return &cc.At, true
	case "timeout":
		return &cc.Timeout, true
	case "debug-file":
		return &cc.DebugFile, true
	default:
		return nil, false
	}
}

// CompCmdBash is a bash completion command.
type CompCmdBash struct {
	ops  [1]*Cmd
	self Cmd
}

// Setup returns the prepared command.
//
// opComplete is expected to be the return value of CompCmdOpComplete.Setup.
func (cc *CompCmdBash) Setup(opComplete *Cmd) *Cmd {
	*cc = CompCmdBash{
		self: Cmd{
			Pattern:  "bash",
			Children: cc.ops[:],
			Run:      generateBashCompletion,
			Help:     helpBash,
		},
		ops: [1]*Cmd{opComplete},
	}

	return &cc.self
}

func helpBash(opts *CmdOptions, route Route, args []string, helpArgAt int) error {
	return writeScript(opts, route, WriteShellCompUsageBash)
}

func generateBashCompletion(opts *CmdOptions, route Route, posArgs, dashArgs []string) error {
	target := route.Target()
	if target == nil || target.Name() != "complete" {
		return writeScript(opts, route, WriteShellCompScriptBash)
	}

	var (
		op = target.LocalFlags.(*CompCmdOpComplete)

		strCols, strCompType string
		cols, compType       uint64

		tsk = op.Task()
	)

	if len(posArgs) > 0 {
		var found bool
		strCols, strCompType, found = strings.Cut(posArgs[0], ",")
		if !found {
			tsk.Debug("bad cols,compType arg", posArgs[0])
			return nil
		}

		var err error
		cols, err = strconv.ParseUint(strCols, 10, 64)
		if err != nil {
			tsk.Debug("bad cols value", strCols, err.Error())
			return nil
		}

		if len(strCompType) > 0 {
			compType, err = strconv.ParseUint(strCompType, 10, 64)
			if err != nil {
				tsk.Debug("bad compType arg", strCompType, err.Error())
				return nil
			}

			tsk.Debug("cols =", strCols, "compType =", strCompType)
		} else {
			compType = 30
			tsk.Debug("cols =", strCols, "compType = 30")
		}
	} else {
		cols = 80
		compType = 30
		tsk.Debug("cols = 80 compType = 30")
	}

	fmt := CompFmtBash{Cols: int(cols), CompType: int(compType)}
	return generateCompletion(
		tsk, route[0], opts, dashArgs, op.At.Value, op.Timeout.Value, noescape(&fmt),
	)
}

// CompCmdZsh is a zsh completion command.
type CompCmdZsh CompCmdBash

// Setup returns the prepared command.
//
// opComplete is expected to be the return value of CompCmdOpComplete.Setup().
func (cc *CompCmdZsh) Setup(opComplete *Cmd) *Cmd {
	*cc = CompCmdZsh{
		self: Cmd{
			Pattern:  "zsh",
			Children: cc.ops[:],
			Run:      generateZshCompletion,
			Help:     helpZsh,
		},
		ops: [1]*Cmd{opComplete},
	}

	return &cc.self
}

func helpZsh(opts *CmdOptions, route Route, args []string, helpArgAt int) error {
	return writeScript(opts, route, WriteShellCompUsageZsh)
}

func generateZshCompletion(opts *CmdOptions, route Route, posArgs, dashArgs []string) error {
	target := route.Target()
	if target == nil || target.Name() != "complete" {
		return writeScript(opts, route, WriteShellCompScriptZsh)
	}

	var (
		op = target.LocalFlags.(*CompCmdOpComplete)
	)

	return generateCompletion(
		op.Task(), route[0], opts, dashArgs, op.At.Value, op.Timeout.Value, CompFmtZsh{},
	)
}

// CompCmdPwsh is a powershell completion command.
type CompCmdPwsh CompCmdBash

// Setup returns the prepared command.
//
// opComplete is expected to be the return value of CompCmdOpComplete.Setup.
func (cc *CompCmdPwsh) Setup(opComplete *Cmd) *Cmd {
	*cc = CompCmdPwsh{
		self: Cmd{
			Pattern:  "pwsh",
			Children: cc.ops[:],
			Run:      generatePwshCompletion,
			Help:     helpPwsh,
		},
		ops: [1]*Cmd{opComplete},
	}
	return &cc.self
}

func helpPwsh(opts *CmdOptions, route Route, args []string, helpArgAt int) error {
	return writeScript(opts, route, WriteShellCompUsagePwsh)
}

func generatePwshCompletion(opts *CmdOptions, route Route, posArgs, dashArgs []string) error {
	target := route.Target()
	if target == nil || target.Name() != "complete" {
		return writeScript(opts, route, WriteShellCompScriptPwsh)
	}

	var (
		op  = target.LocalFlags.(*CompCmdOpComplete)
		tsk = op.Task()
		fmt CompFmtPwsh
	)

	mode := ""
	if len(posArgs) > 0 {
		mode = posArgs[0]
	}

	tsk.Debug("mode =", mode)
	fmt = CompFmtPwsh{Mode: mode}

	return generateCompletion(
		tsk, route[0], opts, dashArgs, op.At.Value, op.Timeout.Value, noescape(&fmt),
	)
}

func writeScript(
	opts *CmdOptions,
	route Route,
	write func(out io.Writer, rootCmdName, completionCmdName string) (int, error),
) error {
	_, err := write(opts.PickStdout(os.Stdout), route[0].Name(), route[1].Name())
	return err
}

func generateCompletion(
	tsk *CompTask,
	root *Cmd,
	opts *CmdOptions,
	dashArgs []string,
	at uint,
	timeout time.Duration,
	fmt CompFmt,
) (err error) {
	tsk.Init(root, opts, int(at), dashArgs...)

	if timeout > 0 {
		done := make(chan struct{})
		go func() {
			defer close(done)

			tsk.AddDefault()
		}()

		timer := time.NewTimer(timeout)
		defer func() {
			if !timer.Stop() {
				<-timer.C
			}
		}()

		select {
		case <-timer.C:
			tsk.Debug("timeout after", timeout.String())
			return ErrTimeout{}
		case <-done:
		}
	} else {
		tsk.AddDefault()
	}

	err = writeCompletions(opts.PickStdout(os.Stdout), tsk, fmt)
	if err != nil {
		tsk.Debug("error writing completion result:", err.Error())
		return
	}

	tsk.Debug("done.")
	return
}

func writeCompletions(out io.Writer, tsk *CompTask, fmt CompFmt) (err error) {
	s := tsk.State()
	if s&CompStateFailed != 0 {
		return nil
	}

	addComma := false
	if s&CompStateOptionNospace != 0 {
		_, err = wstr(out, "nospace")
		if err != nil {
			return
		}

		tsk.Debug("add option: nospace")
		addComma = true
	}

	if s&CompStateOptionNosort != 0 {
		if addComma {
			_, err = wstr(out, ",")
			if err != nil {
				return
			}
		}

		_, err = wstr(out, "nosort")
		if err != nil {
			return
		}

		tsk.Debug("add option: nosort")
		addComma = true
	}

	_, err = wstr(out, "\n")
	if err != nil {
		return
	}

	tsk.Debug("done adding options, now adding completions")
	return fmt.Format(out, noescape(tsk))
}

var (
	//go:embed scripts/bash-usage.txt
	bashCompUsage string
	//go:embed scripts/bash-comp.sh
	bashCompScript string

	//go:embed scripts/zsh-usage.txt
	zshCompUsage string
	//go:embed scripts/zsh-comp.sh
	zshCompScript string

	//go:embed scripts/pwsh-usage.txt
	pwshCompUsage string
	//go:embed scripts/pwsh-comp.ps1
	pwshCompScript string
)

// WriteShellCompScriptBash writes the bash completion script to out.
func WriteShellCompScriptBash(out io.Writer, rootCmdName, completionCmdName string) (int, error) {
	return replaceFuncWEx(
		out, bashCompScript, placeholders{
			rootCmdName:       rootCmdName,
			completionCmdName: completionCmdName,
		}, placeholderFilterFunc, placeholderReplaceFunc,
	)
}

// WriteShellCompScriptBash writes the usage of bash completion script to out.
func WriteShellCompUsageBash(out io.Writer, rootCmdName, completionCmdName string) (int, error) {
	return replaceFuncWEx(
		out, bashCompUsage, placeholders{
			rootCmdName:       rootCmdName,
			completionCmdName: completionCmdName,
		}, placeholderFilterFunc, placeholderReplaceFunc,
	)
}

// WriteShellCompScriptBash writes the zsh completion script to out.
func WriteShellCompScriptZsh(out io.Writer, rootCmdName, completionCmdName string) (int, error) {
	return replaceFuncWEx(
		out, zshCompScript, placeholders{
			rootCmdName:       rootCmdName,
			completionCmdName: completionCmdName,
		}, placeholderFilterFunc, placeholderReplaceFunc,
	)
}

// WriteShellCompScriptBash writes the usage of zsh completion script to out.
func WriteShellCompUsageZsh(out io.Writer, rootCmdName, completionCmdName string) (int, error) {
	return replaceFuncWEx(
		out, zshCompUsage, placeholders{
			rootCmdName:       rootCmdName,
			completionCmdName: completionCmdName,
		}, placeholderFilterFunc, placeholderReplaceFunc,
	)
}

// WriteShellCompScriptBash writes the powershell completion script to out.
func WriteShellCompScriptPwsh(out io.Writer, rootCmdName, completionCmdName string) (int, error) {
	return replaceFuncWEx(
		out, pwshCompScript, placeholders{
			rootCmdName:       rootCmdName,
			completionCmdName: completionCmdName,
		}, placeholderFilterFunc, placeholderReplaceFunc,
	)
}

// WriteShellCompScriptBash writes the usage of powershell completion script to out.
func WriteShellCompUsagePwsh(out io.Writer, rootCmdName, completionCmdName string) (int, error) {
	return replaceFuncWEx(
		out, pwshCompUsage, placeholders{
			rootCmdName:       rootCmdName,
			completionCmdName: completionCmdName,
		}, placeholderFilterFunc, placeholderReplaceFunc,
	)
}

const (
	placeholderName              = "🖖"
	placeholderNameForIdent      = "999"
	placeholderCompletionCmdName = "👀"
)

func placeholderFilterFunc(r rune) bool { return r == '9' || r == '👀' || r == '🖖' }

type placeholders struct {
	rootCmdName       string
	completionCmdName string
}

func placeholderReplaceFunc(w io.Writer, matched string, ctx placeholders) (int, error) {
	switch matched {
	case placeholderName:
		return wstr(w, ctx.rootCmdName)
	case placeholderNameForIdent:
		return replaceFuncW(
			w,
			ctx.rootCmdName,
			filterNonAlphanumeric,
			replaceEveryRuneWithUnderscore,
		)
	case placeholderCompletionCmdName:
		return wstr(w, ctx.completionCmdName)
	default:
		return wstr(w, matched)
	}
}

// filterNonAlphanumeric retruns true if r is not in range [a-zA-Z0-9_]
func filterNonAlphanumeric(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
		return false
	}
	return true
}

func replaceEveryRuneWithUnderscore(out io.Writer, matched string) (n int, err error) {
	var x int
	for range matched {
		x, err = wstr(out, "_")
		n += x
		if err != nil {
			return
		}
	}

	return
}
