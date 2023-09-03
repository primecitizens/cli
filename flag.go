// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"unicode/utf8"
)

// FlagState contains bits representing a Flag's state.
type FlagState uint32

const (
	// FlagStateValueChanged marks the value of the flag is changed.
	//
	// It SHOULD only be set inside method Flag.Decode when set is true.
	FlagStateValueChanged FlagState = 1 << iota

	// FlagStateHidden marks the flag hidden.
	FlagStateHidden

	// FlagStateSetAtMostOnce marks the flag should only enjoy successful
	// decoding with set=true at most once.
	FlagStateSetAtMostOnce
)

func (m FlagState) ValueChanged() bool  { return m&FlagStateValueChanged != 0 }
func (m FlagState) Hidden() bool        { return m&FlagStateHidden != 0 }
func (m FlagState) SetAtMostOnce() bool { return m&FlagStateSetAtMostOnce != 0 }

// IsShorthand returns true if s is a single rune string and is not a hyphen.
func IsShorthand(s string) bool {
	if len(s) == 0 {
		return false
	}

	if s[0] < utf8.RuneSelf {
		return len(s) == 1 && s[0] != '-'
	}

	_, sz := utf8.DecodeRuneInString(s)
	return len(s) == sz
}

// AnyMaybeCompActionAndHelperTerminal is an alias of `any` and indicates
// some component may try to cast the value as a CompAction, HelperTerminal.
type AnyMaybeCompActionAndHelperTerminal = any

// Flag defines the interface of an entry used by FlagIndexer and ParseFlag.
//
// NOTE: In this package, a Flag's default value is provided by the FlagInfo
// and will only be assign to Flags inside Cmd.Exec.
type Flag interface {
	// Type returns (typename, true) if there is type information for this
	// flag.
	Type() (string, bool)

	// ImplyValue returns the text implied by the existence of the flag name.
	//
	// return ("", false) to indicate the flag doesn't have implied value.
	ImplyValue() (string, bool)

	// Decode decodes a text argument to the flag value.
	Decode(opts *ParseOptions, name, arg string, set bool) error

	// Extra returns the user defined extra data.
	Extra() AnyMaybeCompActionAndHelperTerminal

	// FlagState returns the state of the flag.
	State() FlagState

	// HasValue returns true if calling PrintValue will write some
	// value.
	HasValue() bool

	// PrintValue writes the text representation of current value of
	// the flag.
	//
	// It MAY panic if HasValue returned false.
	PrintValue(out io.Writer) (int, error)

	// Usage returns the brief usage of the flag.
	Usage() string
}

// FlagBase holds a pointer to the actual value.
type FlagBase[T any, P VP[*T]] struct {
	// BriefUsage is the help text for terminal user.
	BriefUsage string

	// Ext is the extra custom data for this flag.
	Ext AnyMaybeCompActionAndHelperTerminal

	// Value points to the actual variable to set.
	Value *T

	VP P

	// State_ of the flag.
	State_ FlagState
}

func (f *FlagBase[T, P]) State() FlagState { return f.State_ }
func (f *FlagBase[T, P]) Usage() string    { return f.BriefUsage }
func (f *FlagBase[T, P]) Extra() any       { return f.Ext }

func (f *FlagBase[T, P]) Type() (string, bool) {
	t := f.VP.Type().String()
	return t, len(t) != 0
}

func (f *FlagBase[T, P]) ImplyValue() (string, bool) {
	return implyFromVPType(f.VP.Type())
}

func (f *FlagBase[T, P]) HasValue() bool {
	return f != nil && f.VP.HasValue(f.Value)
}

func (f *FlagBase[T, P]) PrintValue(out io.Writer) (int, error) {
	return f.VP.PrintValue(out, f.Value)
}

func (f *FlagBase[T, P]) Decode(opts *ParseOptions, name, arg string, set bool) error {
	if f.State_.SetAtMostOnce() && f.State_.ValueChanged() && set {
		return ErrFlagSetAtMostOnce{}
	}

	err := f.VP.ParseValue(opts, arg, f.Value, set)
	if err != nil {
		return err
	}

	if set {
		f.State_ |= FlagStateValueChanged
	}

	return nil
}

// FlagBaseV is FlagBase but with value embedded.
type FlagBaseV[T any, P VP[*T]] struct {
	// BriefUsage is the help text for terminal user.
	BriefUsage string

	// Ext is the extra custom data for this flag.
	Ext AnyMaybeCompActionAndHelperTerminal

	// State_ of the flag.
	State_ FlagState

	// Value of the flag.
	Value T

	VP P
}

func (f *FlagBaseV[T, P]) State() FlagState { return f.State_ }
func (f *FlagBaseV[T, P]) Usage() string    { return f.BriefUsage }
func (f *FlagBaseV[T, P]) Extra() any       { return f.Ext }

func (f *FlagBaseV[T, P]) Type() (string, bool) {
	t := f.VP.Type().String()
	return t, len(t) != 0
}

func (f *FlagBaseV[T, P]) ImplyValue() (string, bool) {
	return implyFromVPType(f.VP.Type())
}

func (f *FlagBaseV[T, P]) HasValue() bool {
	return f != nil && f.VP.HasValue(&f.Value)
}

func (f *FlagBaseV[T, P]) PrintValue(out io.Writer) (int, error) {
	return f.VP.PrintValue(out, &f.Value)
}

func (f *FlagBaseV[T, P]) Decode(opts *ParseOptions, name, arg string, set bool) error {
	if f.State_.SetAtMostOnce() && f.State_.ValueChanged() && set {
		return ErrFlagSetAtMostOnce{}
	}

	err := f.VP.ParseValue(opts, arg, &f.Value, set)
	if err != nil {
		return err
	}

	if set {
		f.State_ |= FlagStateValueChanged
	}

	return nil
}

// FlagEmptyV is a flag without value.
type FlagEmptyV FlagBaseV[struct{}, VPNop[*struct{}]]

func (f *FlagEmptyV) State() FlagState                  { return f.State_ }
func (f *FlagEmptyV) Usage() string                     { return f.BriefUsage }
func (f *FlagEmptyV) Extra() any                        { return f.Ext }
func (f *FlagEmptyV) ImplyValue() (string, bool)        { return "", true }
func (f *FlagEmptyV) Type() (string, bool)              { return "", false }
func (f *FlagEmptyV) HasValue() bool                    { return false }
func (f *FlagEmptyV) PrintValue(io.Writer) (int, error) { return 0, nil }
func (f *FlagEmptyV) Decode(opts *ParseOptions, name, arg string, set bool) error {
	return ((*FlagBaseV[struct{}, VPNop[*struct{}]])(f)).Decode(opts, name, arg, set)
}

func implyFromVPType(t VPType) (string, bool) {
	switch t & VPTypeVariantMASK {
	case VPTypeVariantSum:
		switch t & (^VPTypeVariantMASK) {
		case VPTypeInt, VPTypeUint, VPTypeFloat, VPTypeSize:
			return "1", true
		case VPTypeDuration:
			return "1s", true
		}
	case VPTypeVariantSlice:
		switch t & (^VPTypeVariantMASK) {
		case VPTypeBool:
			return "true", true
		}
	case 0:
		switch t {
		case VPTypeBool:
			return "true", true
		}
	}

	return "", false
}
