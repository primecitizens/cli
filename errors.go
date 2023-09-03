// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"strconv"
)

// A FlagViolation represents a rule violation caused by flag.
type FlagViolation Violation

func (err *FlagViolation) Error() string {
	prefix := "--"
	if IsShorthand(err.Key) {
		prefix = "-"
	}

	reason := ""
	switch err.Reason {
	case ViolationCodeNoViolation:
		reason = "no violation"
	case ViolationCodeEmptyAllOf:
		reason = "all flags in the group are required, but none set"
	case ViolationCodePartialAllOf, ViolationCodePartialAllOrNone:
		reason = "not set along with other flags in the same group"
	case ViolationCodeExcessiveOneOf:
		reason = "conflict with other flags in the same group"
	case ViolationCodeEmptyOneOf, ViolationCodeEmptyAnyOf:
		reason = "at least one flag in the group must be set"

	default:
		reason = "unknown (internal error)"
	}

	return "flag rule violation found on `" + prefix + err.Key + "`: " + reason
}

// ErrFlagSetAtMostOnce for flags marked once but appeared more than once.
type ErrFlagSetAtMostOnce struct{}

func (ErrFlagSetAtMostOnce) Error() string {
	return "flag can only be set at most once"
}

type ErrEmptyRoute struct{}

func (ErrEmptyRoute) Error() string { return "empty route" }

// ErrAmbiguousArgs caused by an arg (.Value) with hyphen (`-`) prefix, and
// the arg before it (.Name) is a flag name having implicit value but also
// accepts it as value.
type ErrAmbiguousArgs struct {
	Name  string
	Value string
	At    int
}

func (err *ErrAmbiguousArgs) Error() string {
	prefix := "--"
	if IsShorthand(err.Name) {
		prefix = "-"
	}

	return "ambiguous arg combination `" + prefix + err.Name + " " + err.Value +
		"`: implicit flag followed by potential flag"
}

// ErrShorthandOfExplicitFlagInMiddle caused by a cluster of flag shorthands
// with explicit value assigning (e.g. -abcdefg=foo), some shorthand in middle
// rather than the last one requires a explicit value.
type ErrShorthandOfExplicitFlagInMiddle struct {
	Shorthand        string
	ShorthandCluster string
	Value            string
}

func (err *ErrShorthandOfExplicitFlagInMiddle) Error() string {
	return "non-implicit flag -" + err.Shorthand +
		" cannot use value specified with `=` in middle of shorthands (-" +
		err.ShorthandCluster + "=" + err.Value + ")"
}

// ErrDuplicateFlag is the panic value caused by MapIndexer found
// duplicate flag registration.
type ErrDuplicateFlag struct {
	// Name of the flag found duplicate.
	Name string
}

// Error implements error.
func (err *ErrDuplicateFlag) Error() string {
	prefix := "-"
	if IsShorthand(err.Name) {
		prefix = ""
	}

	return "duplicate flag -" + prefix + err.Name
}

// ErrFlagUndefined
type ErrFlagUndefined struct {
	// Name of the missing flag.
	Name string
	// At
	//
	// When >= 0, the error occurred during flag parsing and args[At] is the
	// arg containing the flag.
	//
	// When < 0, just a flag not defined.
	At int
}

// Error implements error.
func (err *ErrFlagUndefined) Error() string {
	prefix := "-"
	if IsShorthand(err.Name) {
		prefix = ""
	}

	if err.At < 0 {
		return "undefined flag -" + prefix + err.Name
	}

	return "undefined flag -" + prefix + err.Name +
		" (index: " + strconv.FormatInt(int64(err.At), 10) + ")"
}

// ErrFlagValueMissing
type ErrFlagValueMissing struct {
	// Name is a single flag name without standard hyphen prefix.
	Name string

	// At is the arg index into the full arg list.
	At int
}

// Error implements error.
func (err *ErrFlagValueMissing) Error() string {
	prefix := "-"
	if IsShorthand(err.Name) {
		prefix = ""
	}

	return "missing value for flag -" + prefix + err.Name +
		" (index: " + strconv.FormatInt(int64(err.At), 10) + ")"
}

// ErrFlagValueInvalid
type ErrFlagValueInvalid struct {
	// Name of the flag having invalid value.
	Name string

	// Value is the invalid value.
	Value string

	// NameAt is the arg index into the full arg list (`args`)
	//
	// args[NameAt] is the arg containing the flag name.
	NameAt int

	// ValueAt is the arg index into the full arg list (`args`)
	//
	// when >= 0, args[ValueAt] is the arg containing the flag value.
	// otherwise, the invalid value was implied by the flag.
	ValueAt int

	// Reason is the error caused this error.
	Reason error
}

func (err *ErrFlagValueInvalid) Error() string {
	prefix := "--"
	if IsShorthand(err.Name) {
		prefix = "-"
	}

	var reason string
	if err.Reason != nil {
		reason = ": " + err.Reason.Error()
	}

	return "invalid value for flag " + prefix + err.Name +
		" (index: " + strconv.FormatInt(int64(err.NameAt), 10) +
		", value index: " + strconv.FormatInt(int64(err.ValueAt), 10) +
		")" + reason
}

// ErrCmdNotRunnable
type ErrCmdNotRunnable struct {
	Name string
}

func (err *ErrCmdNotRunnable) Error() string {
	return "command " + err.Name + " is not runnable (not having function Run)"
}

// ErrHelpPending for help but no help handle func could be found.
type ErrHelpPending struct {
	// HelpArg is the arg value that requested the help handling.
	HelpArg string
	// At is the HelpArg index into the full arg list.
	At int
}

func (err *ErrHelpPending) Error() string {
	return "help requested by arg `" + err.HelpArg +
		"` (index: " + strconv.FormatInt(int64(err.At), 10) + ") but not handled"
}

// ErrHelpHandled is used to notify caller the help request has been handled.
type ErrHelpHandled struct{}

func (ErrHelpHandled) Error() string { return "help request handled" }

// ErrTimeout
type ErrTimeout struct{}

func (ErrTimeout) Error() string { return "timeout" }

// ErrInvalidValue
type ErrInvalidValue struct {
	// Type is the type or format the Value supposed to be.
	Type string

	// Value is the actual bad value.
	Value string

	// Partial When set to true, means the Value contains a invalid part,
	// and that part is meant to be a Type value.
	Partial bool
}

// Error implements error.
//
//   - When v.Partial is true: "$v.Value contains invalid $v.Type value"
//   - When v.Partial is false: "$v.Value is not a valid %v.Type value"
func (v *ErrInvalidValue) Error() string {
	if v.Partial {
		return v.Value + " contains invalid " + v.Type + " value"
	}

	return v.Value + " is not a valid " + v.Type + " value"
}
