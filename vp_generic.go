// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// VPNop does absolutely nothing.
type VPNop[T any] struct{}

func (VPNop[T]) Type() VPType                                    { return VPTypeUnknown }
func (VPNop[T]) HasValue(v T) bool                               { return false }
func (VPNop[T]) PrintValue(io.Writer, T) (int, error)            { return 0, nil }
func (VPNop[T]) ParseValue(*ParseOptions, string, T, bool) error { return nil }

// VPString for types compatible with string.
//
// It does nothing validating args, so all kinds of strings are accepted.
type VPString[T ~string] struct{}

func (VPString[T]) Type() VPType                                { return VPTypeString }
func (VPString[T]) HasValue(v *T) bool                          { return v != nil && len(*v) != 0 }
func (VPString[T]) PrintValue(out io.Writer, v *T) (int, error) { return wstr(out, string(*v)) }

func (VPString[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) error {
	if set {
		*out = T(arg)
	}
	return nil
}

// VPInt for types compatible with int{, 8, 16, 32, 64}.
//
// It uses strconv.ParseInt to parse args.
type VPInt[T sinteger] struct{}

func (VPInt[T]) Type() VPType       { return VPTypeInt }
func (VPInt[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPInt[T]) PrintValue(out io.Writer, value *T) (n int, err error) {
	v := *value
	if v < 0 {
		n, err = wstr(out, "-")
		if err != nil {
			return
		}

		v = -v
	}

	// at most 20 decimal digits for uint64
	// max: 18446744073709551615
	var buf [20]byte
	i := len(buf) - 1
	for ; i > 0; i-- {
		buf[i] = byte(v%10 + '0')
		if v < 10 {
			break
		}
		v /= 10
	}

	x, err := wstr(out, unsafe.String(noescape(&buf[i]), 20-i))
	n += x
	return
}

func (VPInt[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var tmp T
	x, err := strconv.ParseInt(arg, 0, int(unsafe.Sizeof(tmp))*8)
	if err != nil {
		return
	}

	if set {
		*out = T(x)
	}

	return
}

// VPUint for types compatible with uint{, 8, 16, 32, 64, ptr}.
//
// It uses strconv.ParseUint to parse args.
type VPUint[T uinteger] struct{}

func (VPUint[T]) Type() VPType       { return VPTypeUint }
func (VPUint[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPUint[T]) PrintValue(out io.Writer, value *T) (n int, err error) {
	v := *value
	// at most 20 decimal digits for uint64
	// max: 18446744073709551615
	var buf [20]byte
	i := len(buf) - 1
	for ; i > 0; i-- {
		buf[i] = byte(v%10 + '0')
		if v < 10 {
			break
		}
		v /= 10
	}

	x, err := wstr(out, unsafe.String(noescape(&buf[i]), 20-i))
	n += x
	return
}

func (VPUint[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var tmp T
	x, err := strconv.ParseUint(arg, 0, int(unsafe.Sizeof(tmp))*8)
	if err != nil {
		return
	}

	if set {
		*out = T(x)
	}

	return
}

// VPFloat for types compatible with float{32, 64}.
//
// It uses strconv.ParseFloat to parse args.
type VPFloat[T float] struct{}

func (VPFloat[T]) Type() VPType       { return VPTypeFloat }
func (VPFloat[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPFloat[T]) PrintValue(out io.Writer, value *T) (int, error) {
	v := *value
	return wstr(out, strconv.FormatFloat(float64(v), 'f', -1, int(unsafe.Sizeof(v)*8)))
}

func (VPFloat[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) error {
	var tmp T
	x, err := strconv.ParseFloat(arg, int(unsafe.Sizeof(tmp))*8)
	if err != nil {
		return err
	}

	test := T(x)
	if !math.IsNaN(x) && !math.IsInf(x, 0) && math.Abs(float64(test)-x) > 0.1 {
		return strconv.ErrRange
	}

	if set {
		*out = test
	}

	return nil
}

// VPBool for types compatible with bool.
//
// These args are considered true: "true", "yes", "y", "on", "1"
//
// These args are considered false: "false", "no", "n", "off", "0"
//
// All other values are invalid.
type VPBool[T ~bool] struct{}

func (VPBool[T]) Type() VPType       { return VPTypeBool }
func (VPBool[T]) HasValue(v *T) bool { return v != nil && bool(*v) }

func (VPBool[T]) PrintValue(out io.Writer, v *T) (int, error) {
	if *v {
		return wstr(out, "true")
	}
	return wstr(out, "false")
}

func (VPBool[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) error {
	switch arg {
	case "true", "yes", "y", "on", "1":
		if set {
			*out = true
		}
	case "false", "no", "n", "off", "0":
		if set {
			*out = false
		}
	default:
		return &ErrInvalidValue{
			Type:  "bool",
			Value: arg,
		}
	}

	return nil
}

// VPRegexp for types compatible regexp.Regexp.
//
// When parsing, it compiles the arg as a regular expression using
// regexp.Compile.
type VPRegexp[T regexp.Regexp] struct{}

func (VPRegexp[T]) Type() VPType { return VPTypeRegexp }

func (VPRegexp[T]) HasValue(v *T) bool {
	return v != nil && len((*regexp.Regexp)(unsafe.Pointer(v)).String()) != 0
}

func (VPRegexp[T]) PrintValue(out io.Writer, v *T) (int, error) {
	return wstr(out, (*regexp.Regexp)(unsafe.Pointer(v)).String())
}

func (VPRegexp[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var x *regexp.Regexp
	x, err = regexp.Compile(arg)
	if err != nil {
		return
	}

	if set {
		*out = T(*x)
	}

	return
}

// VPRegexpNocase is VPRegexp but compiles pattern as case-insensitive.
//
// By adding prefix `(?i:` and suffix `)`
type VPRegexpNocase[T regexp.Regexp] struct{}

func (VPRegexpNocase[T]) Type() VPType { return VPTypeRegexp }

func (VPRegexpNocase[T]) HasValue(v *T) bool {
	return v != nil && len((*regexp.Regexp)(unsafe.Pointer(v)).String()) != 0
}

func (VPRegexpNocase[T]) PrintValue(out io.Writer, v *T) (int, error) {
	ptn := (*regexp.Regexp)(unsafe.Pointer(v)).String()
	ptn, ok := strings.CutPrefix(ptn, "(?i:")
	if ok {
		ptn = strings.TrimSuffix(ptn, ")")
	}
	return wstr(out, ptn)
}

func (VPRegexpNocase[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var x *regexp.Regexp
	x, err = regexp.Compile("(?i:" + arg + ")")
	if err != nil {
		return
	}

	if set {
		*out = T(*x)
	}
	return
}

// VPSum sums numeric values.
//
// NOTE: It is not recommended to use VPSum with VPSlice, unless you known
// what you are doing.
type VPSum[T num, P VP[*T]] struct{}

func (VPSum[T, P]) Type() VPType {
	var peeker P
	ret := peeker.Type()
	if ret&VPTypeVariantMASK != 0 {
		return VPTypeUnknown
	}

	return ret | VPTypeVariantSum
}

func (VPSum[T, P]) HasValue(v *T) bool {
	return v != nil && *v != 0
}

func (VPSum[T, P]) PrintValue(out io.Writer, v *T) (int, error) {
	var printer P
	return printer.PrintValue(out, v)
}

func (VPSum[T, P]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var (
		tmp    T
		parser P
	)

	// here we parse value into out instead of &x to avoid allocation.
	err = parser.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil {
		return
	}

	if set {
		*out = *out + tmp
	}
	return
}

// VPSize for size strings with suffix:
//
//	b, B, k, KB, g, GB, t, TB, p, PB, e, EB
//
// parsed value is the size in bytes, size overflow will cause error
type VPSize[T integer] struct{}

func (VPSize[T]) Type() VPType       { return VPTypeSize }
func (VPSize[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPSize[T]) PrintValue(out io.Writer, value *T) (n int, err error) {
	v := *value
	if v < 0 {
		n, err = wstr(out, "-")
		if err != nil {
			return
		}

		v = -v
	}

	// at most 39 bytes for size text in uint64
	// max: 15EB1023PB1023TB1023GB1023MB1023KB1023B
	var buf [40]byte

	i := len(buf) - 1
	for x, j, chunk := uint64(v), 0, uint64(0); ; j++ {
		chunk = x % 1024
		if chunk > 0 {
			buf[i] = 'B'
			i--
			if j != 0 {
				buf[i] = sizeUnitText(j)
				i--
			}

			for ; i > 0; i-- {
				buf[i] = byte(chunk%10 + '0')
				if chunk < 10 {
					if x > 1024 { // not the last write, prepare for next write
						i--
					}
					break
				}
				chunk /= 10
			}
		} else if x < 1024 {
			break
		}

		x /= 1024
	}

	x, err := wstr(out, unsafe.String(noescape(&buf[i]), 40-i))
	n += x
	return
}

func (VPSize[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	neg, x, err := parseSize(arg)
	if err != nil {
		return
	}

	if neg {
		if test := T(-int64(x)); int64(test) != -int64(x) {
			return strconv.ErrRange
		}
	} else {
		if test := T(x); uint64(test) != x {
			return strconv.ErrRange
		}
	}

	if set {
		if neg {
			*out = T(-int64(x))
		} else {
			*out = T(x)
		}
	}

	return
}

// VPDuration decoding duration values to nanoseconds integer.
// Only decimal numbers are supported, other number format may cause
// silent error.
//
// Supported units:
//
//   - `ns` (nanoseconds)
//   - `us` (microseconds)
//   - `ms` (milliseconds)
//   - `s` (seconds)
//   - `m` (minutes)
//   - `h` or `hr` (hours)
//   - `d` (days)
//   - `w` (weeks)
//   - `M` or `mt` (months)
//   - `y` or `yr` (years)
//
// For example:
//
//	1000ns, 200us, 100ms, 1s, 2m, 3h, 4d, 5w, 6M, 6y, 1d2h, 5w3d
//
// NOTE: Months and years MUST be integer values, they are non-deterministic
// and are based on the time from opts.StartTime or time.Now(). All other
// units are deterministic.
type VPDuration[T integer] struct{}

func (VPDuration[T]) Type() VPType       { return VPTypeDuration }
func (VPDuration[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPDuration[T]) PrintValue(out io.Writer, v *T) (int, error) {
	return wstr(out, time.Duration(*v).String())
}

func (VPDuration[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var t time.Time
	if opts == nil || opts.StartTime.IsZero() {
		t = time.Now()
	} else {
		t = opts.StartTime
	}

	neg, x, err := parseDuration(arg, t)
	if err != nil {
		return
	}

	if neg {
		if test := T(-int64(x)); int64(test) != -int64(x) {
			return strconv.ErrRange
		}
	} else {
		if test := T(x); uint64(test) != x {
			return strconv.ErrRange
		}
	}

	if set {
		if neg {
			*out = T(-int64(x))
		} else {
			*out = T(x)
		}
	}
	return
}

// VPUnixSec is like VPTime but the target value is seconds since the
// unix epoch.
type VPUnixSec[T ~int64] struct{}

func (VPUnixSec[T]) Type() VPType       { return VPTypeTime }
func (VPUnixSec[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPUnixSec[T]) PrintValue(out io.Writer, value *T) (int, error) {
	return wstr(out, time.Unix(int64(*value), 0).Format(time.RFC3339Nano))
}

func (VPUnixSec[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var t time.Time
	if opts == nil || opts.StartTime.IsZero() {
		t = time.Now()
	} else {
		t = opts.StartTime
	}

	t, err = parseTime(arg, t)
	if err != nil {
		return
	}

	if set {
		*out = T(t.Unix())
	}
	return
}

// VPUnixMilli is like VPTime but the target value is milliseconds since the
// unix epoch.
type VPUnixMilli[T ~int64] struct{}

func (VPUnixMilli[T]) Type() VPType       { return VPTypeTime }
func (VPUnixMilli[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPUnixMilli[T]) PrintValue(out io.Writer, value *T) (int, error) {
	return wstr(out, time.UnixMilli(int64(*value)).Format(time.RFC3339Nano))
}

func (VPUnixMilli[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var t time.Time
	if opts != nil || opts.StartTime.IsZero() {
		t = opts.StartTime
	} else {
		t = time.Now()
	}

	t, err = parseTime(arg, t)
	if err != nil {
		return
	}

	if set {
		*out = T(t.UnixMilli())
	}
	return
}

// VPUnixMicro is like VPTime but the target value is microseconds since the
// unix epoch.
type VPUnixMicro[T ~int64] struct{}

func (VPUnixMicro[T]) Type() VPType       { return VPTypeTime }
func (VPUnixMicro[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPUnixMicro[T]) PrintValue(out io.Writer, value *T) (int, error) {
	return wstr(out, time.UnixMicro(int64(*value)).Format(time.RFC3339Nano))
}

func (VPUnixMicro[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var t time.Time
	if opts != nil || opts.StartTime.IsZero() {
		t = opts.StartTime
	} else {
		t = time.Now()
	}

	t, err = parseTime(arg, t)
	if err != nil {
		return
	}

	if set {
		*out = T(t.UnixMicro())
	}
	return
}

// VPUnixNano is like VPTime but the target value is nanoseconds since the
// unix epoch.
type VPUnixNano[T ~int64] struct{}

func (VPUnixNano[T]) Type() VPType       { return VPTypeTime }
func (VPUnixNano[T]) HasValue(v *T) bool { return v != nil && *v != 0 }

func (VPUnixNano[T]) PrintValue(out io.Writer, value *T) (int, error) {
	return wstr(out, time.Unix(0, int64(*value)).Format(time.RFC3339Nano))
}

func (VPUnixNano[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var t time.Time
	if opts != nil || opts.StartTime.IsZero() {
		t = opts.StartTime
	} else {
		t = time.Now()
	}

	t, err = parseTime(arg, t)
	if err != nil {
		return
	}

	if set {
		*out = T(t.UnixNano())
	}
	return
}

// VPTime for time values like:
//
//   - 15:04
//   - 2006-01-02
//   - 2006-01-02T15:04:05
//   - 2006-01-02T15:04:05Z07:00
//   - 15:04:05
//   - 15
//
// To parse value, it uses opts.StartTime or time.Now() to fill missing
// date parts.
type VPTime[T time.Time] struct{}

func (VPTime[T]) Type() VPType       { return VPTypeTime }
func (VPTime[T]) HasValue(v *T) bool { return v != nil && !time.Time(*v).IsZero() }

func (VPTime[T]) PrintValue(out io.Writer, v *T) (int, error) {
	return wstr(out, time.Time(*v).Format(time.RFC3339Nano))
}

func (VPTime[T]) ParseValue(opts *ParseOptions, arg string, out *T, set bool) (err error) {
	var t time.Time
	if opts == nil || opts.StartTime.IsZero() {
		t = time.Now()
	} else {
		t = opts.StartTime
	}

	t, err = parseTime(arg, t)
	if err != nil {
		return
	}

	if set {
		*out = T(t)
	}
	return
}

// VPMap wraps other VPs for parsing map[K]E types.
//
// It parses args as "key=value" pairs.
type VPMap[K comparable, E any, KP VP[*K], EP VP[*E]] struct {
	Key   KP
	Value EP
}

func (m VPMap[K, E, KP, EP]) Type() VPType {
	kt := m.Key.Type()
	if kt&VPTypeVariantMASK != 0 { // key can only be scalars (not including sum)
		return VPTypeUnknown
	}

	vt := m.Value.Type()
	if vt&VPTypeVariantMASK == VPTypeVariantMap { // value can only be slice or scalar (including sum)
		return VPTypeUnknown
	}

	return VPTypeVariantMap |
		kt<<VPTypeKeyScalarShift |
		(vt&VPTypeVariantMASK)<<(VPTypeMapElemVariantShift-VPTypeVariantShift) |
		vt&VPTypeElemScalarMASK
}

func (m VPMap[K, E, KP, EP]) HasValue(v *map[K]E) bool { return v != nil && len(*v) != 0 }

func (m VPMap[K, E, KP, EP]) PrintValue(out io.Writer, value *map[K]E) (n int, err error) {
	var (
		x int

		afterFirst bool
	)

	n, err = wstr(out, "[")
	if err != nil {
		return
	}

	for k, v := range *value {
		if afterFirst {
			x, err = wstr(out, ", ")
			n += x
			if err != nil {
				return
			}
		}

		x, err = m.Key.PrintValue(out, noescape(&k))
		n += x
		if err != nil {
			return
		}

		x, err = wstr(out, "=")
		n += x
		if err != nil {
			return
		}

		x, err = m.Value.PrintValue(out, noescape(&v))
		n += x
		if err != nil {
			return
		}

		afterFirst = true
	}

	x, err = wstr(out, "]")
	n += x
	return
}

func (m VPMap[K, E, KP, EP]) ParseValue(opts *ParseOptions, arg string, out *map[K]E, set bool) (err error) {
	strKey, strVal, ok := strings.Cut(arg, "=")
	if !ok {
		return &ErrInvalidValue{
			Type:  "map",
			Value: arg,
		}
	}

	var (
		key K
		val E
	)

	err = m.Key.ParseValue(opts, strKey, noescape(&key), set)
	if err != nil {
		return
	}

	if set {
		m := *out
		if m != nil {
			val = m[key]
		}
	}

	err = m.Value.ParseValue(opts, strVal, noescape(&val), set)
	if err != nil || !set {
		return
	}

	v := *out
	if v == nil {
		v = map[K]E{}
	}

	v[key] = val
	*out = v

	return
}

// VPSlice wraps other VP for parsing []T types.
//
// It appends value parsed by the inner VP to []T.
type VPSlice[Elem any, EP VP[*Elem]] struct{ Elem EP }

func (s VPSlice[E, EP]) Type() VPType {
	ret := s.Elem.Type()
	if ret&VPTypeVariantMASK != 0 {
		return VPTypeUnknown
	}

	return ret | VPTypeVariantSlice
}

func (s VPSlice[E, EP]) HasValue(v *[]E) bool { return v != nil && len(*v) != 0 }

func (s VPSlice[E, EP]) PrintValue(out io.Writer, v *[]E) (n int, err error) {
	var (
		x     int
		slice = *v
	)

	n, err = wstr(out, "[")
	if err != nil {
		return
	}

	for i := range slice {
		if i != 0 {
			x, err = wstr(out, ", ")
			n += x
			if err != nil {
				return
			}
		}

		x, err = s.Elem.PrintValue(out, &slice[i])
		n += x
		if err != nil {
			return
		}
	}

	x, err = wstr(out, "]")
	n += x
	return
}

func (s VPSlice[E, EP]) ParseValue(opts *ParseOptions, arg string, out *[]E, set bool) (err error) {
	var (
		tmp E
	)

	if set {
		err = s.Elem.ParseValue(opts, arg, noescape(&tmp), true)
		if err != nil {
			return
		}

		*out = append(*out, tmp)
		return
	}

	return s.Elem.ParseValue(opts, arg, noescape(&tmp), set)
}

// VPPointer wraps other VP for parsing *T types.
type VPPointer[T any, P VP[*T]] struct{ Elem P }

func (p VPPointer[T, P]) Type() VPType        { return p.Elem.Type() }
func (p VPPointer[T, P]) HasValue(v **T) bool { return v != nil && p.Elem.HasValue(*v) }

func (p VPPointer[T, P]) PrintValue(out io.Writer, v **T) (n int, err error) {
	return p.Elem.PrintValue(out, *v)
}

func (p VPPointer[T, P]) ParseValue(opts *ParseOptions, arg string, out **T, set bool) (err error) {
	if set {
		if *out != nil {
			return p.Elem.ParseValue(opts, arg, noescape(*out), true)
		}

		var tmp T
		err = p.Elem.ParseValue(opts, arg, noescape(&tmp), true)
		if err != nil {
			return
		}

		alloc := tmp
		*out = &alloc
		return
	}

	var tmp T
	return p.Elem.ParseValue(opts, arg, noescape(&tmp), false)
}
