// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package assert

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func Type(t testing.TB, expected, actual any) bool {
	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		t.Errorf("want %T, got %T", expected, actual)
		return false
	}

	return true
}

func ErrorIs[T comparable](t testing.TB, expected T, err error) bool {
	if x, ok := err.(T); ok {
		if reflect.DeepEqual(expected, x) {
			return true
		}

		t.Errorf("want (%[1]T) %[1]v, got (%[2]T) %[2]v", expected, x)
		return false
	}

	t.Errorf("want (%[1]T) %[1]v, got (%[2]T) %[2]v", expected, err)
	return false
}

func NoError(t testing.TB, err error) bool {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return false
	}

	return true
}

func Error(t testing.TB, err error) bool {
	if err == nil {
		t.Error("unexpected no error")
		return false
	}

	return true
}

func EqS[T comparable](t testing.TB, expected, actual []T) (ok bool) {
	ok = true
	if len(expected) >= len(actual) {
		for i, v := range expected {
			if i < len(actual) {
				if v != actual[i] {
					ok = false
					t.Errorf("%d-th: want %v, got %v", i, v, actual[i])
				}
			} else {
				ok = false
				t.Errorf("missing %d-th elem: %v", i, v)
			}
		}
	} else {
		for i, v := range actual {
			if i < len(expected) {
				if v != expected[i] {
					ok = false
					t.Errorf("%d-th: want %v, got %v", i, expected[i], v)
				}
			} else {
				ok = false
				t.Errorf("extra %d-th elem: %v", i, v)
			}
		}
	}

	return
}

func True(t testing.TB, actual bool) bool {
	return Eq(t, true, actual, 1)
}

func False(t testing.TB, actual bool) bool {
	return Eq(t, false, actual, 1)
}

func Eq[T comparable](t testing.TB, expected, actual T, skips ...int) bool {
	if expected != actual {
		skip := 0
		if len(skips) != 0 {
			skip = skips[0]
		}
		_, file, line, ok := runtime.Caller(1 + skip)
		if ok {
			t.Errorf("%s:%d\n%s", file, line, diff(expected, actual))
		} else {
			t.Errorf(diff(expected, actual))
		}
		return false
	}
	return true
}

func diff[E any, A any](expected E, actual A) string {
	wantLines := strings.SplitAfter(fmt.Sprint(expected), "\n")
	actualLines := strings.SplitAfter(fmt.Sprint(actual), "\n")
	var sb strings.Builder
	sb.WriteString("\n")
	for i, ln := range wantLines {
		if len(actualLines) <= i {
			sb.WriteString("ln:")
			sb.WriteString(fmt.Sprintf("%-3d", i+1))
			sb.WriteString("- ")
			sb.WriteString(strconv.Quote(ln))
			sb.WriteString("\n")
			continue
		}

		if ln == actualLines[i] {
			continue
		}

		sb.WriteString("ln:")
		sb.WriteString(fmt.Sprintf("%-3d", i+1))
		sb.WriteString(" - ")
		sb.WriteString(strconv.Quote(ln))
		sb.WriteString("\n       + ")
		sb.WriteString(strconv.Quote(actualLines[i]))
		sb.WriteString("\n")
	}
	return sb.String()
}
