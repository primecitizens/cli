// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import "io"

type ViolationCode uint32

const (
	ViolationCodeNoViolation ViolationCode = iota

	ViolationCodeEmptyAllOf       // for RuleAllOf: none present.
	ViolationCodePartialAllOf     // for RuleAllOf: at least one present but not all present.
	ViolationCodePartialAllOrNone // for RuleAllOrNone: at least one present but not all present.
	ViolationCodeExcessiveOneOf   // for RuleOneOf: more than one present.
	ViolationCodeEmptyOneOf       // for RuleOneOf: none present.
	ViolationCodeEmptyAnyOf       // for RuleAnyOf: none present.
)

type Violation struct {
	Key    string
	Reason ViolationCode
}

// An Inspector is an external verifier used to check rule enforcement.
type Inspector interface {
	// CheckFlagValueChanged returns ture if the flag's value has been changed.
	CheckFlagValueChanged(key string) bool
}

// Rule
type Rule interface {
	// Requires returns true if this rule requires Inspector.CheckFlagValueChanged() to
	// return true for this key to avoid violation.
	Requires(key string) bool

	// Contains returns true if the rule has something to do with the key.
	Contains(key string) bool

	// NthEx returns (i-th violation, true), if any, otherwise (_, false).
	NthEx(f Inspector, i int) (Violation, bool)

	// WriteFlagRule writes text representation of the rule.
	//
	// If len(keys) > 0, only add rule applies to at least one of the keys,
	// otherwise add all.
	WriteFlagRule(out io.Writer, keys ...string) (int, error)
}

func RuleContainsAny[R Rule](rule R, keys ...string) bool {
	for _, arg := range keys {
		if rule.Contains(arg) {
			return true
		}
	}

	return false
}

func RuleRequiresAny[R Rule](rule R, keys ...string) bool {
	for _, arg := range keys {
		if rule.Requires(arg) {
			return true
		}
	}

	return false
}

// MergeFlagRules merges multiple Rules into one.
func MergeFlagRules(rules ...Rule) Rule {
	switch len(rules) {
	case 0:
		return RuleAny{}
	case 1:
		return rules[0]
	default:
		return &MultiRule{Rules: rules}
	}
}

// RuleAny
type RuleAny struct{}

func (RuleAny) Requires(key string) bool                        { return false }
func (RuleAny) Contains(key string) bool                        { return false }
func (RuleAny) NthEx(Inspector, int) (Violation, bool)          { return Violation{}, false }
func (RuleAny) WriteFlagRule(io.Writer, ...string) (int, error) { return 0, nil }

// MultiRule
type MultiRule struct {
	Rules []Rule
}

func (r *MultiRule) Requires(key string) bool {
	for _, rule := range r.Rules {
		if rule.Requires(key) {
			return true
		}
	}

	return false
}

func (r *MultiRule) Contains(key string) bool {
	for _, rule := range r.Rules {
		if rule.Contains(key) {
			return true
		}
	}

	return false
}

// NthEx implements [Rule].
func (r *MultiRule) NthEx(f Inspector, i int) (Violation, bool) {
	for _, rule := range r.Rules {
		for j := 0; ; j++ {
			v, ok := rule.NthEx(f, j)
			if !ok {
				break
			}

			if i == 0 {
				return v, true
			}

			i--
		}
	}

	return Violation{}, false
}

func (r *MultiRule) WriteFlagRule(out io.Writer, keys ...string) (n int, err error) {
	var x int
	for _, rule := range r.Rules {
		if len(keys) != 0 && !RuleContainsAny(rule, keys...) {
			x = 0
			continue
		}

		if x > 0 {
			// last rule wrote at least one tag
			x, err = wstr(out, " & ")
			n += x
			if err != nil {
				return
			}
		}

		x, err = rule.WriteFlagRule(out, keys...)
		n += x
		if err != nil {
			return
		}
	}

	return
}

// AllOf creates a *RuleAllOf from provided keys.
func AllOf(keys ...string) *RuleAllOf {
	return &RuleAllOf{Keys: keys}
}

type RuleAllOf struct {
	Keys []string
}

func (r *RuleAllOf) Requires(key string) bool { return sliceContains(r.Keys, key) }
func (r *RuleAllOf) Contains(key string) bool { return sliceContains(r.Keys, key) }

// NthEx implements [Rule].
func (r *RuleAllOf) NthEx(f Inspector, i int) (Violation, bool) {
	if i < 0 || r == nil {
		return Violation{}, false
	}

	var (
		someNotSet, someSet bool
		idxFirstNotSet      int
	)

	for j, name := range r.Keys {
		if f.CheckFlagValueChanged(name) {
			someSet = true
			if someNotSet {
				break
			}

			continue
		}

		if !someNotSet {
			idxFirstNotSet = j
			someNotSet = true
		}
		if someSet {
			break
		}
	}

	if someNotSet {
		for cur, j := 0, idxFirstNotSet; j < len(r.Keys); j++ {
			if f.CheckFlagValueChanged(r.Keys[j]) {
				continue
			}

			if cur == i {
				if someSet {
					return Violation{Key: r.Keys[j], Reason: ViolationCodePartialAllOf}, true
				}

				return Violation{Key: r.Keys[j], Reason: ViolationCodeEmptyAllOf}, true
			}

			cur++
		}
	}

	return Violation{}, false
}

func (r *RuleAllOf) WriteFlagRule(out io.Writer, keys ...string) (n int, err error) {
	if len(keys) == 0 || RuleContainsAny(r, keys...) {
		return formatFlagRuleTags(out, "allof[", "]", r.Keys)
	}

	return 0, nil
}

// AllOrNone creates a *RuleAllOrNone from provided keys.
func AllOrNone(keys ...string) *RuleAllOrNone {
	return &RuleAllOrNone{Keys: keys}
}

// RuleAllOrNone represents a group of keys MUST either all set or all not set.
type RuleAllOrNone struct {
	Keys []string
}

func (r *RuleAllOrNone) Requires(key string) bool { return false }
func (r *RuleAllOrNone) Contains(key string) bool { return sliceContains(r.Keys, key) }

// NthEx implements [Rule].
func (r *RuleAllOrNone) NthEx(f Inspector, i int) (Violation, bool) {
	if i < 0 || r == nil {
		return Violation{}, false
	}

	var (
		someSet, someNotSet bool
		idxFirstNotSet      int
	)

	for j, name := range r.Keys {
		if f.CheckFlagValueChanged(name) {
			someSet = true
			if someNotSet {
				break
			}

			continue
		}

		if !someNotSet {
			idxFirstNotSet = j
			someNotSet = true
			if someSet {
				break
			}
		}
	}

	if someSet && someNotSet { // violation presents
		for cur, j := 0, idxFirstNotSet; j < len(r.Keys); j++ {
			if f.CheckFlagValueChanged(r.Keys[j]) {
				continue
			}

			if cur == i {
				return Violation{Key: r.Keys[j], Reason: ViolationCodePartialAllOrNone}, true
			}

			cur++
		}
	}

	return Violation{}, false
}

func (r *RuleAllOrNone) WriteFlagRule(out io.Writer, keys ...string) (int, error) {
	if len(keys) == 0 || RuleContainsAny(r, keys...) {
		return formatFlagRuleTags(out, "allOrNone[", "]", r.Keys)
	}

	return 0, nil
}

// OneOf creates a *RuleOneOf from provided keys.
func OneOf(keys ...string) *RuleOneOf {
	return &RuleOneOf{Keys: keys}
}

// RuleOneOf defines a group of mutually exclusive flags.
type RuleOneOf struct {
	Keys []string
}

func (r *RuleOneOf) Requires(key string) bool { return singleRequire(r.Keys, key) }
func (r *RuleOneOf) Contains(key string) bool { return sliceContains(r.Keys, key) }

// NthEx implements [Rule].
func (r *RuleOneOf) NthEx(f Inspector, i int) (Violation, bool) {
	if i < 0 || r == nil {
		return Violation{}, false
	}

	var (
		someSet      bool
		idxSecondSet int
	)

	for j, name := range r.Keys {
		if !f.CheckFlagValueChanged(name) {
			continue
		}

		if !someSet {
			someSet = true
		} else {
			idxSecondSet = j
			break
		}
	}

	if !someSet {
		if i < len(r.Keys) && i >= 0 {
			return Violation{Key: r.Keys[i], Reason: ViolationCodeEmptyOneOf}, true
		}

		return Violation{}, false
	}

	if idxSecondSet != 0 {
		for cur, j := 0, idxSecondSet; j < len(r.Keys); j++ {
			if !f.CheckFlagValueChanged(r.Keys[j]) {
				continue
			}

			if cur == i {
				return Violation{Key: r.Keys[j], Reason: ViolationCodeExcessiveOneOf}, true
			}

			cur++
		}
	}

	return Violation{}, false
}

func (r *RuleOneOf) WriteFlagRule(out io.Writer, keys ...string) (int, error) {
	if len(keys) == 0 || RuleContainsAny(r, keys...) {
		return formatFlagRuleTags(out, "oneof[", "]", r.Keys)
	}

	return 0, nil
}

// AnyOf creates a *RuleAnyOf from provided keys.
func AnyOf(keys ...string) *RuleAnyOf {
	return &RuleAnyOf{Keys: keys}
}

// RuleAnyOf represents a group of flags that at least one MUST present.
type RuleAnyOf struct {
	Keys []string
}

func (r *RuleAnyOf) Requires(key string) bool { return singleRequire(r.Keys, key) }
func (r *RuleAnyOf) Contains(key string) bool { return sliceContains(r.Keys, key) }

// NthEx implements [Rule].
func (r *RuleAnyOf) NthEx(f Inspector, i int) (Violation, bool) {
	if i < 0 || r == nil {
		return Violation{}, false
	}

	var someSet bool
	for _, name := range r.Keys {
		if f.CheckFlagValueChanged(name) {
			someSet = true
			break
		}
	}

	if !someSet && i < len(r.Keys) && i >= 0 {
		return Violation{Key: r.Keys[i], Reason: ViolationCodeEmptyAnyOf}, true
	}

	return Violation{}, false
}

func (r *RuleAnyOf) WriteFlagRule(out io.Writer, keys ...string) (int, error) {
	if len(keys) == 0 || RuleContainsAny(r, keys...) {
		return formatFlagRuleTags(out, "anyof[", "]", r.Keys)
	}

	return 0, nil
}

// DependOn creates a *RuleDepends from the given condition and branches.
func DependOn[X, Y, Z Rule](ifX X, thenY Y, elseZ Z) *RuleDepends[X, Y, Z] {
	return &RuleDepends[X, Y, Z]{
		If:   ifX,
		Then: thenY,
		Else: elseZ,
	}
}

// RuleDepends represents a group of flags required only when dependent rules
// are met.
type RuleDepends[X, Y, Z Rule] struct {
	// If rules in `If` produces no violation, `Then` is used to decide
	// this rule has been satisfied if `Then` produces no violation.
	//
	// `Else` is used as `Then` if `If` produced violation.
	//
	// It is valid that `If` contains no rule.
	If   X
	Then Y
	Else Z
}

func (r *RuleDepends[X, Y, Z]) Requires(key string) bool {
	return r.Then.Requires(key) && r.Else.Requires(key)
}

func (r *RuleDepends[X, Y, Z]) Contains(key string) bool {
	return r.If.Contains(key) || r.Then.Contains(key) || r.Else.Contains(key)
}

// NthEx implements [Rule].
func (r *RuleDepends[X, Y, Z]) NthEx(f Inspector, i int) (p Violation, ok bool) {
	if i < 0 || r == nil {
		return p, false
	}

	useThen := true
	// only need to check whether there is violation, so passing `0` is fine.
	_, ok = r.If.NthEx(f, 0)
	if ok {
		useThen = false
	}

	if useThen {
		for k := 0; ; k++ {
			p, ok = r.Then.NthEx(f, k)
			if !ok {
				break
			}

			if i == 0 {
				return p, true
			}

			i--
		}
	} else {
		for k := 0; ; k++ {
			p, ok = r.Else.NthEx(f, k)
			if !ok {
				break
			}

			if i == 0 {
				return p, true
			}

			i--
		}
	}

	return Violation{}, false
}

func (r *RuleDepends[X, Y, Z]) WriteFlagRule(out io.Writer, keys ...string) (n int, err error) {
	n, err = wstr(out, "(if ")
	if err != nil {
		return
	}

	x, err := r.If.WriteFlagRule(out, keys...)
	n += x
	if err != nil {
		return
	}

	if x == 0 {
		x, err = wstr(out, "nop")
		n += x
		if err != nil {
			return
		}
	}

	x, err = wstr(out, "; then ")
	n += x
	if err != nil {
		return
	}

	x, err = r.Then.WriteFlagRule(out, keys...)
	n += x
	if err != nil {
		return
	}

	if x == 0 {
		x, err = wstr(out, "nop")
		n += x
		if err != nil {
			return
		}
	}

	x, err = wstr(out, "; else ")
	n += x
	if err != nil {
		return
	}

	x, err = r.Else.WriteFlagRule(out)
	n += x
	if err != nil {
		return
	}

	if x == 0 {
		x, err = wstr(out, "nop)")
		n += x
		if err != nil {
			return
		}
	}

	return
}

func sliceContains(ss []string, keys ...string) bool {
	for _, x := range ss {
		for _, y := range keys {
			if x == y {
				return true
			}
		}
	}

	return false
}

func singleRequire(slice []string, keys ...string) bool {
	hasKey := false
	for _, k := range slice {
		if len(k) == 0 {
			continue
		}

		if !sliceContains(keys, k) {
			return false
		}

		hasKey = true
	}

	return hasKey
}

func formatFlagRuleTags(
	out io.Writer,
	prefix, suffix string,
	ruleKeys []string,
) (n int, err error) {
	var (
		x     int
		wrote bool
	)
	n, err = wstr(out, prefix)
	if err != nil {
		return
	}

	for _, key := range ruleKeys {
		if len(key) == 0 {
			continue
		}

		if wrote {
			x, err = wstr(out, ", ")
			n += x
			if err != nil {
				return
			}
		} else {
			wrote = true
		}

		if IsShorthand(key) {
			x, err = wstr(out, "-")
		} else {
			x, err = wstr(out, "--")
		}
		n += x
		if err != nil {
			return
		}

		x, err = wstr(out, key)
		n += x
		if err != nil {
			return
		}
	}

	x, err = wstr(out, suffix)
	n += x
	return
}
