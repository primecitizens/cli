// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"strings"
	"unicode/utf8"
)

// CompFmt defines completion result formatter.
type CompFmt interface {
	// Format writes one CompItem each line with shell specific character
	// escaping and formatting.
	Format(out io.Writer, finishedTask *CompTask) error
}

// CompFmtBash implements [CompFmt] for bash.
//
// It produces two kinds of lines:
//   - ' <value>' (space prefixed) where <value> contains arguments to bash function _filedir.
//   - others (without space prefix), as bash-completion COMPREPLY element.
type CompFmtBash struct {
	// Cols is supposed to be the $COLUMNS in bash completion.
	Cols int

	// CompType is the type of completion:
	//
	//  - '\t' (9) for normal completion
	//  - '?' (63) for listing completions after successive tabs
	//  - '!' (33) for listing alternatives on partial word completion
	//  - '@' (64) to list completions if the word is not unmodified
	//  - '%' (37) for menu completion
	//  - '*' (42) for insert completion
	//
	// Refs:
	//  - https://www.gnu.org/savannah-checkouts/gnu/bash/manual/bash.html#index-COMP_005fTYPE
	CompType int
}

func (fmt *CompFmtBash) Format(out io.Writer, tsk *CompTask) (err error) {
	var (
		indent          int
		omitDescription bool
		wantFiles       bool
		wantDirs        bool
	)

	switch fmt.CompType {
	case '\t': // 9
	case '?': // 63
	case '!': // 33
	case '@': // 64
	case '%': // 37
		omitDescription = true
	case '*': // 42
		omitDescription = true
	}

	if !omitDescription {
		// find the longest value, assume monospace font
		for i := 0; ; i++ {
			item, ok := tsk.Nth(i)
			if !ok {
				break
			}

			x := utf8.RuneCountInString(item.Value) + strings.Count(item.Value, "\x20")
			if item.Kind == CompKindFlagName {
				if IsShorthand(item.Value) {
					x += 1
				} else {
					x += 2
				}
			}

			if x > indent {
				indent = x
			}
		}

		indent += 4 /* spaces between value and description */
	}

	for i := 0; ; i++ {
		item, ok := tsk.Nth(i)
		if !ok {
			break
		}

		switch item.Kind {
		case CompKindFiles:
			wantFiles = true
			continue
		case CompKindDirs:
			wantDirs = true
			continue
		case CompKindFlagValue:
			if len(item.Value) == 0 {
				continue
			}

			_, err = fmt.EscapeSpaces(out, tsk.FlagValuePrefix, true)
			if err != nil {
				return
			}
		case CompKindFlagName:
			if len(item.Value) == 0 {
				continue
			}

			if IsShorthand(item.Value) {
				_, err = wstr(out, "-")
			} else {
				_, err = wstr(out, "--")
			}
			if err != nil {
				return
			}
		default:
			if len(item.Value) == 0 {
				continue
			}
		}

		_, err = fmt.EscapeSpaces(out, item.Value, true)
		if err != nil {
			return
		}

		if !omitDescription && len(item.Description) != 0 {
			if i == 0 {
				if _, ok = tsk.Nth(1); !ok {
					// there is only one option, omit the description
					goto newline
				}
			}

			spaces := indent - utf8.RuneCountInString(item.Value) - strings.Count(item.Value, "\x20")
			descCap := fmt.Cols - indent
			if descCap <= 0 {
				goto newline
			}

			for j := 0; j < spaces; j++ {
				_, err = wstr(out, " ")
				if err != nil {
					return
				}
			}

			if descLen := utf8.RuneCountInString(item.Description); descCap >= descLen {
				_, err = writeline(out, item.Description)
				if err != nil {
					return
				}
			} else {
				var j int
				for j = range item.Description {
					if j == descCap-3 {
						break
					}
				}

				_, err = writeline(out, item.Description[:j])
				if err != nil {
					return
				}

				_, err = wstr(out, "...")
				if err != nil {
					return
				}
			}
		}

	newline:
		_, err = wstr(out, "\n")
		if err != nil {
			return
		}
	}

	// add extra '\x20' (space) prefix as other completion output can never
	// have such line prefix.
	if wantFiles {
		tsk.Debug("add", "file", "matching")
		_, err = wstr(out, "\x20")
	} else if wantDirs {
		tsk.Debug("add", "dir", "matching")
		_, err = wstr(out, "\x20-d")
	}
	if err != nil {
		return
	}

	if wantFiles || wantDirs {
		var hasFilter bool
		for i := 0; ; i++ {
			item, ok := tsk.Nth(i)
			if !ok {
				break
			}

			switch item.Kind {
			case CompKindFiles, CompKindDirs:
				if len(item.Value) == 0 {
					continue
				}
			default:
				continue
			}

			if !hasFilter {
				hasFilter = true
				_, err = wstr(out, "\x20'")
				if err != nil {
					return
				}

				_, err = writeline(out, item.Value)
				if err != nil {
					return
				}
			} else {
				_, err = wstr(out, "|")
				if err != nil {
					return
				}

				_, err = writeline(out, item.Value)
				if err != nil {
					return
				}
			}
		}

		if hasFilter {
			_, err = wstr(out, "'\n")
		} else {
			_, err = wstr(out, "\n")
		}
		if err != nil {
			return
		}
	}

	return nil
}

func (fmt *CompFmtBash) EscapeSpaces(out io.Writer, s string, oneline bool) (int, error) {
	if oneline {
		s, _, _ = strings.Cut(s, "\n")
	}

	return replaceFuncW(
		out, s, filterBashSpaces, replaceBashSpaces,
	)
}

func filterBashSpaces(r rune) bool { return r == '\x20' }
func replaceBashSpaces(out io.Writer, matched string) (n int, err error) {
	var x int
	for _, r := range matched {
		switch r {
		case '\x20':
			x, err = wstr(out, "\\\x20")
		default:
			panic("unreachable")
		}
		n += x
		if err != nil {
			return
		}
	}
	return
}

// CompFmtZsh implements [CompFmt] for zsh.
//
// It produces two kinds of lines:
//   - `<value>:<description>` for zsh function _describe.
//   - `:<argument-spec>` (note the colon prefix) for zsh function _arguments,
//     currently only used for filename and dirname completion.
type CompFmtZsh struct{}

func (fmt CompFmtZsh) Format(out io.Writer, tsk *CompTask) (err error) {
	var (
		wantFiles bool
		wantDirs  bool
	)

	for i := 0; ; i++ {
		item, ok := tsk.Nth(i)
		if !ok {
			break
		}

		switch item.Kind {
		case CompKindFiles:
			wantFiles = true
			continue
		case CompKindDirs:
			wantDirs = true
			continue
		case CompKindFlagValue:
			if len(item.Value) == 0 {
				continue
			}

			_, err = fmt.EscapeColons(out, tsk.FlagValuePrefix, true)
			if err != nil {
				return
			}
		case CompKindFlagName:
			if len(item.Value) == 0 {
				continue
			}

			if IsShorthand(item.Value) {
				_, err = wstr(out, "-")
			} else {
				_, err = wstr(out, "--")
			}
			if err != nil {
				return
			}
		default:
			if len(item.Value) == 0 {
				continue
			}
		}

		// zsh doc for _describe:
		//
		//	The array name1 contains the possible completions with their descriptions
		//	in the form ‘completion:description’.
		//	Any literal colons in completion must be quoted with a backslash.

		_, err = fmt.EscapeColons(out, item.Value, true)
		if err != nil {
			return
		}

		if len(item.Description) > 0 {
			_, err = wstr(out, ":")
			if err != nil {
				return
			}

			_, err = fmt.EscapeColons(out, item.Description, true)
			if err != nil {
				return
			}
		}

		_, err = wstr(out, "\n")
		if err != nil {
			return
		}
	}

	// add extra ':' prefix as other completion output can never
	// have such line prefix (indicating empty value with description).
	if wantFiles { // want regular files and dirs
		_, err = wstr(out, ":*:filename:_files")
	} else if wantDirs { // only want dirs
		_, err = wstr(out, ":*:dirname:_files -/")
	}
	if err != nil {
		return
	}

	if wantFiles || wantDirs {
		var hasFilter bool
		for i := 0; ; i++ {
			item, ok := tsk.Nth(i)
			if !ok {
				break
			}

			switch item.Kind {
			case CompKindFiles, CompKindDirs:
				if len(item.Value) == 0 {
					continue
				}
			default:
				continue
			}

			if !hasFilter {
				hasFilter = true
				_, err = wstr(out, " -g (")
				if err != nil {
					return
				}

				_, err = writeline(out, item.Value)
				if err != nil {
					return
				}
			} else {
				_, err = wstr(out, "|")
				if err != nil {
					return
				}

				_, err = writeline(out, item.Value)
				if err != nil {
					return
				}
			}
		}

		if hasFilter {
			_, err = wstr(out, ")\n")
		} else {
			_, err = wstr(out, "\n")
		}
		if err != nil {
			return
		}
	}

	return
}

func (CompFmtZsh) EscapeColons(out io.Writer, s string, oneline bool) (int, error) {
	if oneline {
		s, _, _ = strings.Cut(s, "\n")
	}

	return replaceFuncW(out, s, filterZshColons, replaceZshColons)
}

func filterZshColons(r rune) bool { return r == ':' }
func replaceZshColons(out io.Writer, matched string) (n int, err error) {
	var x int
	for _, r := range matched {
		switch r {
		case ':':
			x, err = wstr(out, "\\:")
		// case '"':
		// 	x, err = writeString(out,"\\\"")
		default:
			panic("unreachable")
		}
		n += x
		if err != nil {
			return
		}
	}
	return
}

// CompFmtPwsh implements [CompFmt] for powershell.
//
// It produces two kinds of lines:
//   - `<value> ;<description>` (note the space and unescaped semi-colon) for
//     creating CompletionResult items.
//   - `;<argument-spec>` (note the unescaped semi-colon prefix) for filesystem
//     related completion.
type CompFmtPwsh struct {
	// Mods is the PowerShell completion mode, possible values are:
	//
	//  - TabCompleteNext (default windows style - on each key press the next option is displayed)
	//  - Complete (works like bash)
	//  - MenuComplete (works like zsh)
	Mode string
}

func (fmt *CompFmtPwsh) Format(out io.Writer, tsk *CompTask) (err error) {
	var (
		isBash    = fmt.Mode == "Complete"
		wantFiles bool
		wantDirs  bool

		indent int
	)

	if isBash { // needs padding between value and description
		// find the longest value, assume monospace font
		for i := 0; ; i++ {
			item, ok := tsk.Nth(i)
			if !ok {
				break
			}

			x := utf8.RuneCountInString(item.Value)
			if item.Kind == CompKindFlagName {
				if IsShorthand(item.Value) {
					x += 1
				} else {
					x += 2
				}
			}

			if x > indent {
				indent = x
			}
		}

		indent += 4 /* spaces between value and description */
	}

	for i := 0; ; i++ {
		item, ok := tsk.Nth(i)
		if !ok {
			break
		}

		switch item.Kind {
		case CompKindFiles:
			wantFiles = true
			continue
		case CompKindDirs:
			wantDirs = true
			continue
		case CompKindFlagValue:
			if len(item.Value) == 0 {
				continue
			}

			_, err = fmt.EscapeSpecialChars(out, tsk.FlagValuePrefix, true)
			if err != nil {
				return
			}
		case CompKindFlagName:
			if len(item.Value) == 0 {
				continue
			}

			if IsShorthand(item.Value) {
				_, err = wstr(out, "-")
			} else {
				_, err = wstr(out, "--")
			}
			if err != nil {
				return
			}
		default:
			if len(item.Value) == 0 {
				continue
			}
		}

		_, err = fmt.EscapeSpecialChars(out, item.Value, true)
		if err != nil {
			return
		}

		if len(item.Description) != 0 {
			_, err = wstr(out, " ;")
			if err != nil {
				return
			}

			if isBash {
				if i == 0 {
					if _, ok = tsk.Nth(1); !ok {
						// there is only one option, omit the description
						goto newline
					}
				}

				spaces := indent - utf8.RuneCountInString(item.Value)
				for j := 0; j < spaces; j++ {
					_, err = wstr(out, " ")
					if err != nil {
						return
					}
				}

				_, err = writeline(out, item.Description)
				if err != nil {
					return
				}
			} else {
				_, err = writeline(out, item.Description)
				if err != nil {
					return
				}
			}
		}

	newline:
		_, err = wstr(out, "\n")
		if err != nil {
			return
		}
	}

	// TODO: support dir only fs match
	if wantFiles || wantDirs {
		// add extra ';' (semi-colon) prefix as other completion output can never
		// have such line prefix (indicating empty value with description).
		_, err = wstr(out, ";")

		var hasFilter bool
		for i := 0; ; i++ {
			item, ok := tsk.Nth(i)
			if !ok {
				break
			}

			switch item.Kind {
			case CompKindFiles, CompKindDirs:
				if len(item.Value) == 0 {
					continue
				}
			default:
				continue
			}

			if !hasFilter {
				hasFilter = true
				_, err = wstr(out, "'(")
				if err != nil {
					return
				}

				_, err = writeline(out, item.Value)
				if err != nil {
					return
				}
			} else {
				_, err = wstr(out, "|")
				if err != nil {
					return
				}

				_, err = writeline(out, item.Value)
				if err != nil {
					return
				}
			}
		}

		if hasFilter {
			_, err = wstr(out, ")'\n")
		} else {
			_, err = wstr(out, "'")
			if err != nil {
				return
			}

			_, err = wstr(out, tsk.ToComplete)
			if err != nil {
				return
			}

			_, err = wstr(out, "*'\n")
		}
		if err != nil {
			return
		}
	}

	return
}

func (fmt *CompFmtPwsh) EscapeSpecialChars(out io.Writer, s string, oneline bool) (int, error) {
	if oneline {
		s, _, _ = strings.Cut(s, "\n")
	}

	return replaceFuncW(
		out, s, isSpecialCharInPwsh, replacePwshSpecialChars,
	)
}

func isSpecialCharInPwsh(r rune) bool {
	switch r {
	case '\t', '\x20', '\r', '\n', '\x00',
		'{', '}', '(', ')', '<', '>',
		'#', '@', '$', ';', ',', '\'', '"', '`', '\\', '&', '|':
		return true
	}
	return false
}

func replacePwshSpecialChars(out io.Writer, matched string) (n int, err error) {
	for i, x := 0, 0; i < len(matched); i++ {
		x, err = wstr(out, "`")
		n += x
		if err != nil {
			return
		}

		x, err = wstr(out, matched[i:i+1])
		n += x
		if err != nil {
			return
		}
	}

	return
}

func writeline(out io.Writer, s string) (int, error) {
	s, _, _ = strings.Cut(s, "\n")
	return wstr(out, s)
}
