// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"strings"
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

func TestRule_WriteFlagRule(t *testing.T) {
	for _, test := range []struct {
		rule     Rule
		expected string
	}{
		{AllOf(), "allof[]"},
		{AllOf(""), "allof[]"},
		{AllOf("foo", ""), "allof[--foo]"},
		{AllOf("foo", "bar"), "allof[--foo, --bar]"},

		{AllOrNone(), "allOrNone[]"},
		{AllOrNone(""), "allOrNone[]"},
		{AllOrNone("foo", ""), "allOrNone[--foo]"},
		{AllOrNone("foo", "bar"), "allOrNone[--foo, --bar]"},

		{OneOf(), "oneof[]"},
		{OneOf(""), "oneof[]"},
		{OneOf("foo", ""), "oneof[--foo]"},
		{OneOf("foo", "bar"), "oneof[--foo, --bar]"},

		{AnyOf(), "anyof[]"},
		{AnyOf(""), "anyof[]"},
		{AnyOf("foo", ""), "anyof[--foo]"},
		{AnyOf("foo", "bar"), "anyof[--foo, --bar]"},

		{MergeFlagRules(AllOf("foo", "bar"), AnyOf("bar", "woo")), "allof[--foo, --bar] & anyof[--bar, --woo]"},

		{DependOn(RuleAny{}, RuleAny{}, RuleAny{}), "(if nop; then nop; else nop)"},
	} {
		t.Run("", func(t *testing.T) {
			var sb strings.Builder
			_, err := test.rule.WriteFlagRule(&sb)
			assert.NoError(t, err)

			assert.Eq(t, test.expected, sb.String())
		})
	}
}

func TestRule_Requires(t *testing.T) {
	for _, test := range []struct {
		contains bool
		requires bool
		rule     Rule
		input    string
	}{
		{true, true, AllOf("foo"), "foo"},
		{true, true, AllOf("foo", "bar"), "foo"},
		{true, true, AllOf("foo", "bar"), "bar"},
		{false, false, AllOf("foo", "bar"), "woo"},

		{true, false, AllOrNone("foo"), "foo"},
		{true, false, AllOrNone("foo", "bar"), "foo"},
		{true, false, AllOrNone("foo", "bar"), "bar"},
		{false, false, AllOrNone("foo", "bar"), "woo"},

		{true, true, OneOf("foo"), "foo"},
		{true, false, OneOf("foo", "bar"), "foo"},
		{true, false, OneOf("foo", "bar"), "bar"},
		{false, false, OneOf("foo", "bar"), "woo"},

		{true, true, AnyOf("foo"), "foo"},
		{true, false, AnyOf("foo", "bar"), "foo"},
		{true, false, AnyOf("foo", "bar"), "bar"},
		{false, false, AnyOf("foo", "bar"), "woo"},

		{true, true, MergeFlagRules(AllOf("foo", "bar"), AllOf("bar", "woo")), "woo"},
		{true, false, MergeFlagRules(AnyOf("foo", "bar"), AnyOf("bar", "woo")), "bar"},
		{false, false, MergeFlagRules(AnyOf("foo", "bar"), AnyOf("bar", "woo")), "non-existing"},

		{false, false, DependOn(RuleAny{}, AllOf("foo"), OneOf("foo")), "bar"},
		{true, true, DependOn(RuleAny{}, AllOf("foo"), OneOf("foo")), "foo"},
		{true, false, DependOn(RuleAny{}, AllOf("foo"), OneOf("foo", "bar")), "bar"},
	} {
		t.Run("", func(t *testing.T) {
			assert.Eq(t, test.contains, test.rule.Contains(test.input))
			assert.Eq(t, test.requires, test.rule.Requires(test.input))
		})
	}
}

type testInspector struct{}

func (testInspector) CheckFlagValueChanged(key string) bool {
	if strings.HasPrefix(key, "ok") {
		return true
	} else {
		return false
	}
}

func TestRule_NthEx(t *testing.T) {
	for _, test := range []struct {
		run    func(*testing.T, Rule, ViolationCode)
		rules  Rule
		reason ViolationCode
	}{
		{noViolation, AllOf("ok-1"), ViolationCodeNoViolation},
		{noViolation, AllOf("ok-1", "ok-2"), ViolationCodeNoViolation},
		{noViolation, AllOf("ok-1", "ok-2", "ok-3"), ViolationCodeNoViolation},
		{oneViolation, AllOf("bad-1"), ViolationCodeEmptyAllOf},
		{oneViolation, AllOf("bad-1", "ok-1", "ok-2"), ViolationCodePartialAllOf},
		{oneViolation, AllOf("ok-1", "bad-1", "ok-2"), ViolationCodePartialAllOf},
		{oneViolation, AllOf("ok-1", "ok-2", "bad-1"), ViolationCodePartialAllOf},
		{twoViolation, AllOf("bad-1", "bad-2"), ViolationCodeEmptyAllOf},
		{twoViolation, AllOf("bad-1", "bad-2", "ok-1", "ok-2"), ViolationCodePartialAllOf},
		{twoViolation, AllOf("bad-1", "ok-1", "bad-2", "ok-2"), ViolationCodePartialAllOf},
		{twoViolation, AllOf("ok-1", "bad-1", "bad-2", "ok-2"), ViolationCodePartialAllOf},
		{twoViolation, AllOf("ok-1", "bad-1", "ok-2", "bad-2"), ViolationCodePartialAllOf},
		{twoViolation, AllOf("ok-1", "ok-2", "bad-1", "bad-2"), ViolationCodePartialAllOf},

		{noViolation, AllOrNone("ok-1"), ViolationCodeNoViolation},
		{noViolation, AllOrNone("ok-1", "ok-2"), ViolationCodeNoViolation},
		{noViolation, AllOrNone("ok-1", "ok-2", "ok-3"), ViolationCodeNoViolation},
		{noViolation, AllOrNone("bad-1"), ViolationCodeNoViolation},
		{noViolation, AllOrNone("bad-1", "bad-2"), ViolationCodeNoViolation},
		{noViolation, AllOrNone("bad-1", "bad-2", "bad-3"), ViolationCodeNoViolation},
		{oneViolation, AllOrNone("bad-1", "ok-1", "ok-2"), ViolationCodePartialAllOrNone},
		{oneViolation, AllOrNone("ok-1", "bad-1", "ok-2"), ViolationCodePartialAllOrNone},
		{oneViolation, AllOrNone("ok-1", "ok-2", "bad-1"), ViolationCodePartialAllOrNone},
		{twoViolation, AllOrNone("bad-1", "bad-2", "ok-1", "ok-2"), ViolationCodePartialAllOrNone},
		{twoViolation, AllOrNone("bad-1", "ok-1", "bad-2", "ok-2"), ViolationCodePartialAllOrNone},
		{twoViolation, AllOrNone("ok-1", "bad-1", "bad-2", "ok-2"), ViolationCodePartialAllOrNone},
		{twoViolation, AllOrNone("ok-1", "bad-1", "ok-2", "bad-2"), ViolationCodePartialAllOrNone},
		{twoViolation, AllOrNone("ok-1", "ok-2", "bad-1", "bad-2"), ViolationCodePartialAllOrNone},

		{noViolation, OneOf("ok-1"), ViolationCodeNoViolation},
		{noViolation, OneOf("ok-1", "bad-1"), ViolationCodeNoViolation},
		{noViolation, OneOf("bad-1", "ok-1"), ViolationCodeNoViolation},
		{noViolation, OneOf("ok-1", "bad-1", "bad-2"), ViolationCodeNoViolation},
		{noViolation, OneOf("bad-1", "ok-1", "bad-2"), ViolationCodeNoViolation},
		{noViolation, OneOf("bad-1", "bad-2", "ok-1"), ViolationCodeNoViolation},
		{oneViolation, OneOf("bad-1"), ViolationCodeEmptyOneOf},
		{oneViolation, OneOf("bad-1", "ok-1", "ok-2"), ViolationCodeExcessiveOneOf},
		{oneViolation, OneOf("ok-1", "bad-1", "ok-2"), ViolationCodeExcessiveOneOf},
		{oneViolation, OneOf("ok-1", "ok-2", "bad-1"), ViolationCodeExcessiveOneOf},
		{twoViolation, OneOf("bad-1", "ok-1", "ok-2", "ok-3"), ViolationCodeExcessiveOneOf},
		{twoViolation, OneOf("ok-1", "bad-1", "ok-2", "ok-3"), ViolationCodeExcessiveOneOf},
		{twoViolation, OneOf("ok-1", "ok-2", "bad-1", "ok-3"), ViolationCodeExcessiveOneOf},
		{twoViolation, OneOf("ok-1", "ok-2", "ok-3", "bad-1"), ViolationCodeExcessiveOneOf},

		{noViolation, AnyOf("ok-1"), ViolationCodeNoViolation},
		{noViolation, AnyOf("ok-1", "ok-2"), ViolationCodeNoViolation},
		{noViolation, AnyOf("ok-1", "ok-2", "ok-3"), ViolationCodeNoViolation},
		{noViolation, AnyOf("ok-1", "bad-1"), ViolationCodeNoViolation},
		{noViolation, AnyOf("bad-1", "ok-1"), ViolationCodeNoViolation},
		{noViolation, AnyOf("ok-1", "bad-1", "bad-2"), ViolationCodeNoViolation},
		{noViolation, AnyOf("bad-1", "ok-1", "bad-2"), ViolationCodeNoViolation},
		{noViolation, AnyOf("bad-1", "bad-2", "ok-1"), ViolationCodeNoViolation},
		{oneViolation, AnyOf("bad-1"), ViolationCodeEmptyAnyOf},
		{twoViolation, AnyOf("bad-1", "bad-2"), ViolationCodeEmptyAnyOf},

		// if:ok, then:ok, else:bad
		{noViolation, DependOn(AllOf("ok-1"), AllOf("ok-2"), AllOf("bad-1")), ViolationCodeNoViolation},
		// if:bad, then:bad, else:ok
		{noViolation, DependOn(AllOf("bad-1"), AllOf("bad-2"), AllOf("ok-1")), ViolationCodeNoViolation},
		// if:ok, then:bad, else:ok
		{oneViolation, DependOn(AllOf("ok-1"), AllOf("bad-2"), AllOf("ok-1")), ViolationCodeEmptyAllOf},
		// if:bad, then:ok, else:bad
		{oneViolation, DependOn(AllOf("bad-1"), AllOf("ok-1"), AllOf("bad-2")), ViolationCodeEmptyAllOf},
		// if:ok, then:bad, else:ok
		{twoViolation, DependOn(AllOf("ok-1"), AllOf("bad-2", "bad-3"), AllOf("ok-1")), ViolationCodeEmptyAllOf},
		// if:bad, then:ok, else:bad
		{twoViolation, DependOn(AllOf("bad-1"), AllOf("ok-1"), AllOf("bad-2", "bad-3")), ViolationCodeEmptyAllOf},

		{noViolation, MergeFlagRules(), ViolationCodeNoViolation},
		{noViolation, MergeFlagRules(OneOf("ok-1")), ViolationCodeNoViolation},
		{noViolation, MergeFlagRules(OneOf("ok-1"), AllOf("ok-2")), ViolationCodeNoViolation},
		{noViolation, MergeFlagRules(OneOf("ok-1"), AllOf("ok-2"), AnyOf("ok-3")), ViolationCodeNoViolation},
		{oneViolation, MergeFlagRules(OneOf("bad-1"), AllOf("ok-2"), AnyOf("ok-3")), ViolationCodeEmptyOneOf},
		{oneViolation, MergeFlagRules(OneOf("ok-1"), AllOf("bad-2"), AnyOf("ok-3")), ViolationCodeEmptyAllOf},
		{oneViolation, MergeFlagRules(OneOf("ok-1"), AllOf("ok-2"), AnyOf("bad-3")), ViolationCodeEmptyAnyOf},
	} {
		_, ok := test.rules.NthEx(testInspector{}, -1)
		if ok {
			t.Errorf("unexpected negative index support")
		}

		test.run(t, test.rules, test.reason)
	}
}

// noViolation expects no violation
func noViolation(t *testing.T, rule Rule, _ ViolationCode) {
	nV(t, rule, ViolationCodeNoViolation, 0)
}

// oneViolation expects exactly one violation.
func oneViolation(t *testing.T, rule Rule, code ViolationCode) {
	nV(t, rule, code, 1)
}

// twoViolation expects exactly 2 violations
func twoViolation(t *testing.T, rule Rule, code ViolationCode) {
	nV(t, rule, code, 2)
}

func nV(t *testing.T, rule Rule, reason ViolationCode, n int) {
	i := 0
	for ; ; i++ {
		p, ok := rule.NthEx(testInspector{}, i)
		if i < n {
			if !ok {
				t.Errorf(
					"unexpected no violation (i = %d): name = %s",
					i, p.Key,
				)
			}

			if reason != p.Reason {
				t.Errorf(
					"unexpected violation reason: name = %s, want = %d, got = %d",
					p.Key, reason, p.Reason,
				)
			}
		} else {
			if ok {
				t.Errorf(
					"doesn't expect violation (i = %d): name = %s, reason = %d",
					i, p.Key, p.Reason,
				)
			}
		}
		if !ok {
			break
		}
	}
	if i != n {
		t.Errorf("expecting %d violations, got %d", n, i)
	}
}
