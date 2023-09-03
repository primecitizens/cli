// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/primecitizens/cli/internal/assert"
)

func TestReplaceFuncW(t *testing.T) {
	for _, test := range []struct {
		text     string
		match    func(rune) bool
		replace  func(io.Writer, string) (int, error)
		expected string
	}{
		{"foo1234bar", unicode.IsDigit,
			func(w io.Writer, s string) (int, error) { return wstr(w, "[numbers]") },
			"foo[numbers]bar"},
		{"foobar", unicode.IsLetter,
			func(w io.Writer, s string) (int, error) { return wstr(w, "[letters]") },
			"[letters]"},
	} {
		var sb strings.Builder
		_, err := replaceFuncW(&sb,
			test.text,
			test.match,
			test.replace,
		)
		if err != nil {
			t.Error(err)
		}

		if sb.String() != test.expected {
			t.Errorf("want %q, got %q", test.expected, sb.String())
		}
	}
}

func TestParseDuration(t *testing.T) {
	const (
		NS     = uint64(time.Nanosecond)
		US     = uint64(time.Microsecond)
		MS     = uint64(time.Millisecond)
		SECOND = uint64(time.Second)
		MINUTE = uint64(time.Minute)
		HOUR   = uint64(time.Hour)
		DAY    = 24 * HOUR
		WEEK   = 7 * DAY
	)

	base := time.Date(2022, time.November, 4, 18, 0, 0, 0, time.UTC)

	for _, test := range []struct {
		good     bool
		dur      string
		expected uint64
	}{
		{true, "", 0},

		{true, "1", 1 * SECOND},
		{true, "10s", 10 * SECOND},
		{true, "2.5m", 150 * SECOND},
		{true, "1.1h", 66 * MINUTE},
		{true, "1mt", 30 * DAY},
		{true, "2M", (30 + 31) * DAY},
		{true, "3mt", (30 + 31 + 31) * DAY},
		{true, "1y", 365 * DAY},
		{true, "3yr", (365 + 365 + 366) * DAY},
		{true, "2d", 2 * DAY},
		{true, "3w", 3 * WEEK},
		{true, "1yr4mt1w1d1hr1m1s1ms1us1ns", 365*DAY + (30+31+31+29)*DAY + WEEK + DAY + HOUR + MINUTE + SECOND + MS + US + NS},

		// bad values
		{false, "xxx", 0},
		{false, "1nx", 0},
		{false, "1ux", 0},
		{false, "2hh", 0},
		{false, "1.1y", 0},
		{false, "1.1M", 0},
		{false, "1.1.1", 0},
	} {
		t.Run(test.dur, func(t *testing.T) {
			neg, ret, err := parseDuration(test.dur, base)
			if test.good {
				assert.NoError(t, err)
				assert.False(t, neg)
				assert.Eq(t, test.expected, ret)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Skipf("tz data not loaded: %v", err)
		return
	}

	base := time.Date(2022, time.November, 4, 18, 0, 0, 0, loc)

	for _, test := range []struct {
		good     bool
		time     string
		expected time.Time
	}{
		{true, "17", base.Add(-time.Hour)},
		{true, "17:01", base.Add(-time.Hour + time.Minute)},
		{true, "17:00:01", base.Add(-time.Hour + time.Second)},
		{true, "2026-12-26", base.AddDate(4, 1, 22).Add(-18 * time.Hour)},
		{true, "2026-12-26T17:01:01", base.AddDate(4, 1, 22).Add(-time.Hour + time.Minute + time.Second)},
		{true, "2026-12-26T17:01:01+08:00", base.AddDate(4, 1, 22).Add(time.Minute + time.Second)},

		// bad values
		{false, "", base},
		{false, "xx", base},
	} {
		t.Run(test.time, func(t *testing.T) {
			ret, err := parseTime(test.time, base)
			if test.good {
				assert.NoError(t, err)
				assert.Eq(t, test.expected, ret)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
		PB = 1024 * TB
		EB = 1024 * PB
	)

	for _, test := range []struct {
		good     bool
		sz       string
		expected uint64
	}{
		{true, "", 0},

		{true, "10", 10 * B},
		{true, "100B", 100 * B},
		{true, "100b", 100 * B},
		{true, "2kb", 2 * KB},
		{true, "2.25KB", 2*KB + 256*B},
		{true, "1Mb", 1 * MB},
		{true, "1.5m", 1*MB + 512*KB},
		{true, "1G", 1 * GB},
		{true, "1.5gB", 1*GB + 512*MB},
		{true, "1T", 1 * TB},
		{true, "10T", 10 * TB},
		{true, "1.5t", 1*TB + 512*GB},
		{true, "1P", 1 * PB},
		{true, "1.5p", 1*PB + 512*TB},
		{true, "1p1T1gB1m1kb1b", PB + TB + GB + MB + KB + B},
		{true, "1eb", EB},

		// bad values
		{false, "xxx", 0},
		{false, "pp", 0},
		{false, "p2p", 0},
		{false, "1.1.1", 0},
	} {
		t.Run(test.sz, func(t *testing.T) {
			neg, ret, err := parseSize(test.sz)
			if test.good {
				assert.NoError(t, err)
				assert.False(t, neg)
				assert.Eq(t, test.expected, ret)
			} else {
				assert.Error(t, err)
			}
		})

	}
}
