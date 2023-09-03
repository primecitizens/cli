// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"
)

// noescape hides a pointer from escape analysis.
//
// see $GOROOT/src/runtime/stubs.go
//
//go:nosplit
func noescape[T any](p *T) *T {
	x := uintptr(unsafe.Pointer(p)) // this line is required
	return (*T)(unsafe.Pointer(x ^ 0))
}

func noescapeSlice[T any](p []T) (ret []T) {
	ret = unsafe.Slice(noescape(unsafe.SliceData(p)), cap(p))
	return ret[:len(p)]
}

// wstr writes string s to out without escaping content of s to heap.
func wstr(out io.Writer, s string) (int, error) {
	return out.Write(unsafe.Slice(noescape(unsafe.StringData(s)), len(s)))
}

func callWriteFlagRule[R Rule](rule R, out io.Writer, args ...string) (int, error) {
	return rule.WriteFlagRule(out, noescapeSlice(args)...)
}

func replaceFuncW(
	w io.Writer,
	text string,
	match func(r rune) bool,
	writeReplace func(w io.Writer, matched string) (int, error),
) (n int, err error) {
	return replaceFuncWEx(
		w, text, writeReplace, match, func(w io.Writer, matched string, writeReplace func(w io.Writer, matched string) (int, error)) (int, error) {
			return writeReplace(w, matched)
		},
	)
}

// replaceFuncWEx is like replaceFuncWEx, but takes an extra argument.
func replaceFuncWEx[Arg any](
	w io.Writer,
	text string,
	arg Arg,
	match func(r rune) bool,
	writeReplace func(w io.Writer, matched string, arg Arg) (int, error),
) (n int, err error) {
	var (
		x, lastMatchedIdx int
		lastMismatchedIdx = len(text)
		wasMatched        = true
	)

	for i, c := range text {
		if match(c) {
			if !wasMatched {
				// this matches, prev mismatched,
				//
				// i is the end of latest match
				x, err = wstr(w, text[lastMatchedIdx:i])
				n += x
				if err != nil {
					return
				}
				wasMatched = true
				lastMismatchedIdx = i
			}
		} else {
			if wasMatched {
				// this mismatch, prev matched
				//
				// i is the end of latest mismatch
				if lastMismatchedIdx < i {
					x, err = writeReplace(w, text[lastMismatchedIdx:i], arg)
					n += x
					if err != nil {
						return
					}
				}

				lastMatchedIdx = i
				wasMatched = false
			}
		}
	}

	if wasMatched {
		x, err = writeReplace(w, text[lastMismatchedIdx:], arg)
	} else {
		x, err = wstr(w, text[lastMatchedIdx:])
	}
	n += x

	return
}

type uinteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type sinteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// integer is the type union of both signed and unsigned integers.
type integer interface {
	uinteger | sinteger
}

// float is the type union of floating-point types.
type float interface {
	~float32 | ~float64
}

// num is the type union of all number types.
type num interface {
	integer | float | ~complex64 | ~complex128
}

// parseNum parses a string as number.
func parseNum(str string, assumeFloat bool) (neg bool, uv uint64, isFloat bool, err error) {
	str, neg = strings.CutPrefix(str, "-")
	if !assumeFloat {
		uv, err = strconv.ParseUint(str, 0, 64)
		if err == nil {
			return neg, uv, false, nil
		}
	}

	// maybe it's a float?
	fv, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return neg, math.Float64bits(fv), true, nil
	}

	return neg, math.Float64bits(math.NaN()), true, strconv.ErrSyntax
}

func parseDuration(s string, base time.Time) (neg bool, dur uint64, err error) {
	var (
		n      = len(s)
		start  int
		adjust int
		unit   uint64
		uv     uint64
		x      time.Time

		uvneg   bool
		isFloat bool
		c       byte
	)

	for i := 0; i < n; i++ {
		c = s[i]

		switch c {
		case 'y': // year
			uvneg, uv, isFloat, err = parseNum(s[start:i+adjust], isFloat)
			if err != nil {
				return
			}

			if isFloat {
				err = &ErrInvalidValue{
					Type:    "year (integer-only)",
					Partial: true,
					Value:   s,
				}
				return
			}

			if uvneg {
				x = base.AddDate(-int(uv), 0, 0)
				neg, dur = uintPlus(neg, dur, true, uint64(base.Sub(x)))
			} else {
				x = base.AddDate(int(uv), 0, 0)
				neg, dur = uintPlus(neg, dur, false, uint64(x.Sub(base)))
			}
			base = x
			if i+1 < n && s[i+1] == 'r' {
				// allow `yr`
				i++
			}

			goto Reset
		case 'm': // minute, month or ms
			if i+1 < n {
				i++
				if s[i] == 's' {
					// ms
					unit = uint64(time.Millisecond)
					adjust = -1
					break
				} else if s[i] != 't' {
					// minute
					unit = uint64(time.Minute)
					i--
					break
				}
			} else {
				// minute
				unit = uint64(time.Minute)
				break
			}

			// month (mt)
			adjust = -1
			fallthrough
		case 'M': // month
			uvneg, uv, isFloat, err = parseNum(s[start:i+adjust], isFloat)
			if err != nil {
				return
			}

			if isFloat {
				err = &ErrInvalidValue{
					Type:    "month (integer-only)",
					Partial: true,
					Value:   s,
				}
				return
			}

			if uvneg {
				x = base.AddDate(0, -int(uv), 0)
				neg, dur = uintPlus(neg, dur, true, uint64(base.Sub(x)))
			} else {
				x = base.AddDate(0, int(uv), 0)
				neg, dur = uintPlus(neg, dur, false, uint64(x.Sub(base)))
			}
			base = x

			goto Reset
		case '.': // 1.1s
			if isFloat {
				return false, 0, &ErrInvalidValue{
					Type:    "floating-point",
					Partial: true,
					Value:   s,
				}
			}

			isFloat = true
			continue
		default:
			if c < '0' || c > '9' {
				return false, 0, &ErrInvalidValue{
					Type:    "numeric",
					Partial: true,
					Value:   s,
				}
			}

			if i+1 != n {
				continue
			}

			// no unit at the end, assume seconds
			i++
			fallthrough
		case 's':
			unit = uint64(time.Second)
		case 'n':
			if i+1 < n && s[i+1] == 's' {
				unit = uint64(time.Nanosecond)
				i++
				adjust = -1
				break
			}

			err = &ErrInvalidValue{
				Type:    "unit (ns)",
				Partial: true,
				Value:   s,
			}
			return
		case 'u':
			if i+1 < n && s[i+1] == 's' {
				unit = uint64(time.Microsecond)
				i++
				adjust = -1
				break
			}

			err = &ErrInvalidValue{
				Type:    "unit (us)",
				Partial: true,
				Value:   s,
			}
			return
		case 'h':
			unit = uint64(time.Hour)
			if i+1 < n && s[i+1] == 'r' {
				// allow `hr`
				i++
				adjust = -1
			}
		case 'd':
			unit = uint64(time.Hour) * 24
		case 'w':
			unit = uint64(time.Hour) * 24 * 7
		}

		// got a unit
		uvneg, uv, isFloat, err = parseNum(s[start:i+adjust], isFloat)
		if err != nil {
			return
		}

		if isFloat {
			neg, dur = uintPlus(neg, dur, uvneg, uint64(math.Float64frombits(uv)*float64(unit)))
		} else {
			neg, dur = uintPlus(neg, dur, uvneg, uv*unit)
		}

	Reset:
		start, isFloat, adjust = i+1, false, 0
	}

	return
}

// parseTime parses time string s in following format preference
//
//   - 15:04
//   - 2006-01-02
//   - 2006-01-02T15:04:05
//   - 2006-01-02T15:04:05Z07:00
//   - 15:04:05
//   - 15
func parseTime(s string, base time.Time) (t time.Time, err error) {
	loc := base.Location()

	// clock: hour:min
	t, err = time.ParseInLocation("15:04", s, loc)
	if err == nil {
		y, mt, d := base.Date()
		hr, m, sec := t.Clock()
		return time.Date(y, mt, d, hr, m, sec, 0, loc), nil
	}

	// date without time
	t, err = time.ParseInLocation("2006-01-02", s, loc)
	if err == nil {
		return
	}

	// date without timezone
	t, err = time.ParseInLocation("2006-01-02T15:04:05", s, loc)
	if err == nil {
		return
	}

	// date with timezone
	t, err = time.ParseInLocation(time.RFC3339, s, loc)
	if err == nil {
		return t.In(loc), nil
	}

	// clock: hour:min:sec
	t, err = time.ParseInLocation("15:04:05", s, loc)
	if err == nil {
		y, mt, d := base.Date()
		hr, m, sec := t.Clock()
		return time.Date(y, mt, d, hr, m, sec, 0, loc), nil
	}

	// clock: hour
	t, err = time.ParseInLocation("15", s, loc)
	if err == nil {
		y, mt, d := base.Date()
		return time.Date(y, mt, d, t.Hour(), 0, 0, 0, loc), nil
	}

	err = &ErrInvalidValue{
		Type:  "time",
		Value: s,
	}
	return
}

// parseSize parses a size string with units like K, G, TB... to an int64 with
// byte as the unit.
func parseSize(s string) (neg bool, sz uint64, err error) {
	var (
		start int
		uv    uint64

		uvneg   bool
		isFloat bool
		hasDot  bool
		c       byte
	)

	for i := 0; i < len(s); i++ {
		c = s[i]

		switch c {
		default:
			if c < '0' || c > '9' {
				return false, 0, &ErrInvalidValue{
					Type:    "numeric",
					Partial: true,
					Value:   s,
				}
			}

			if i+1 != len(s) {
				continue
			}

			// no unit at the end, assume byte
			c = 'b'
			i++
			fallthrough
		case 'B', 'b', 'K', 'k', 'M', 'm', 'G', 'g', 'T', 't', 'P', 'p', 'E', 'e':
			if i == 0 {
				return false, 0, &ErrInvalidValue{
					Type:    "numeric",
					Partial: true,
					Value:   s,
				}
			}

			uvneg, uv, isFloat, err = parseNum(s[start:i], hasDot)
			if err != nil {
				return
			}

			if isFloat {
				neg, sz = uintPlus(neg, sz, uvneg, uint64(math.Float64frombits(uv)*float64(sizeUnitInteger(c))))
			} else {
				neg, sz = uintPlus(neg, sz, uvneg, uv*sizeUnitInteger(c))
			}

			if i+1 < len(s) {
				switch s[i+1] {
				case 'B', 'b':
					i++
				}
			}

			start, hasDot = i+1, false
		case '.': // 1.1g
			if hasDot {
				return false, 0, &ErrInvalidValue{
					Type:    "floating-point",
					Partial: true,
					Value:   s,
				}
			}

			hasDot = true
		}
	}

	return
}

func uintPlus(aNeg bool, aVal uint64, bNeg bool, bVal uint64) (bool, uint64) {
	switch {
	case aNeg && bNeg, !aNeg && !bNeg:
		return aNeg, aVal + bVal
	case aNeg && !bNeg:
		if bVal > aVal {
			return false, bVal - aVal
		}

		return true, aVal - bVal
	case !aNeg && bNeg:
		if aVal > bVal {
			return false, aVal - bVal
		}

		return true, bVal - aVal
	}

	panic("unreachable")
}

func sizeUnitInteger(c byte) uint64 {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
		PB = 1024 * TB
		EB = 1024 * PB
	)

	switch c {
	case 'B', 'b':
		return B
	case 'K', 'k':
		return KB
	case 'M', 'm':
		return MB
	case 'G', 'g':
		return GB
	case 'T', 't':
		return TB
	case 'P', 'p':
		return PB
	case 'E', 'e':
		return EB
	default:
		return 0 // overflow
	}
}

func sizeUnitText(i int) byte {
	switch i {
	case 1:
		return 'K'
	case 2:
		return 'M'
	case 3:
		return 'G'
	case 4:
		return 'T'
	case 5:
		return 'P'
	case 6:
		return 'E'
	default:
		return 0 // overflow
	}
}

const (
	sizeStaticLev = 64
)

// isSimilar returns true if the Levenshtein distance between known and
// toCompare is less than min(3, len(known)).
//
// The result can only be true when min(len(known), min(toCompare)) < 64.
func isSimilar(known, toCompare string, nocase bool) bool {
	switch {
	case len(known) == 0:
		return len(toCompare) == 0
	case len(toCompare) == 0:
		return len(known) < 3
	case len(known) < sizeStaticLev:
		// optimize for more inner loop.
		if len(toCompare) < len(known) {
			return max63Lev(toCompare, known, nocase) < min(3, len(known))
		} else {
			return max63Lev(known, toCompare, nocase) < min(3, len(known))
		}
	case len(toCompare) < sizeStaticLev:
		return max63Lev(known, toCompare, nocase) < min(3, len(known))
	default:
		return false
	}
}

// levenshteinDistance returns the Levenshtein distance between two strings.
//
// len(max63) < 64 is assumed.
func max63Lev(y, max63 string, nocase bool) int {
	var (
		runeX, runeY      rune
		dX, dY, col, offX int

		buf = [2][sizeStaticLev]int{
			0: {
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
				10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
				20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
				30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
				40, 41, 42, 43, 44, 45, 46, 47, 48, 49,
				50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
				60, 61, 62, 63,
			},
			1: {0: 1},
		}

		mat = [2][]int{
			0: buf[0][:],
			1: buf[1][:],
		}
	)

	for row, offY := 1, 0; offY < len(y); row, offY = row+1, offY+dY {
		if runeY = rune(y[offY]); runeY < utf8.RuneSelf {
			dY = 1
		} else {
			runeY, dY = utf8.DecodeRuneInString(y[offY:])
		}

		mat[0][0] = row
		for col, offX = 1, 0; offX < len(max63); col, offX = col+1, offX+dX {
			if runeX = rune(max63[offX]); runeX < utf8.RuneSelf {
				dX = 1
			} else {
				runeX, dX = utf8.DecodeRuneInString(max63[offX:])
			}

			if (!nocase && runeX == runeY) ||
				(nocase && strings.EqualFold(y[offY:offY+dY], max63[offX:offX+dX])) {
				mat[1][col] = mat[0][col-1]
				continue
			}

			mat[1][col] = min(min(mat[0][col], mat[1][col-1]), mat[0][col-1]) + 1
		}

		mat[0], mat[1] = mat[1], mat[0]
	}

	return mat[0][col-1]
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
