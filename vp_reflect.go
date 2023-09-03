// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"math/bits"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// VPReflectBool is the reflect version of VPBool.
//
// It accepts arbitrary depth of pointers.
type VPReflectBool struct{}

func (VPReflectBool) Type() VPType                   { return VPTypeBool }
func (VPReflectBool) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectBool) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Bool()
	return VPBool[bool]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectBool) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp bool
	err = VPBool[bool]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	_, v = prepareRValue(v.Type(), value, set)
	v.SetBool(tmp)
	return
}

// VPReflectString is the reflect version of VPString.
//
// It accepts arbitrary depth of pointers.
type VPReflectString struct{}

func (VPReflectString) Type() VPType                   { return VPTypeString }
func (VPReflectString) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectString) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.String()
	return VPString[string]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectString) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp string
	err = VPString[string]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	_, v = prepareRValue(v.Type(), value, set)
	v.SetString(tmp)
	return
}

// VPReflectInt is the reflect version of VPInt.
//
// It accepts arbitrary depth of pointers.
type VPReflectInt struct{}

func (VPReflectInt) Type() VPType                   { return VPTypeInt }
func (VPReflectInt) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectInt) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Int()
	return VPInt[int64]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectInt) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	if !set {
		_, err = strconv.ParseInt(arg, 0, 64)
		return
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	tmp, err := strconv.ParseInt(arg, 0, typ.Bits())
	if err != nil {
		return
	}

	v.SetInt(tmp)
	return
}

// VPReflectUint is the reflect version of VPUint.
//
// It accepts arbitrary depth of pointers.
type VPReflectUint struct{}

func (VPReflectUint) Type() VPType                   { return VPTypeUint }
func (VPReflectUint) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectUint) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Uint()
	return VPUint[uint64]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectUint) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	if !set {
		_, err = strconv.ParseUint(arg, 0, 64)
		return
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	tmp, err := strconv.ParseUint(arg, 0, typ.Bits())
	if err != nil {
		return
	}

	v.SetUint(tmp)
	return
}

// VPReflectFloat is the reflect version of VPFloat.
//
// It accepts arbitrary depth of pointers.
type VPReflectFloat struct{}

func (VPReflectFloat) Type() VPType                   { return VPTypeFloat }
func (VPReflectFloat) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectFloat) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Float()
	return VPFloat[float64]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectFloat) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	if !set {
		_, err = strconv.ParseFloat(arg, 64)
		return
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	tmp, err := strconv.ParseFloat(arg, typ.Bits())
	if err != nil {
		return
	}

	v.SetFloat(tmp)
	return
}

// VPReflectSize is the reflect version of VPSize.
//
// It accepts arbitrary depth of pointers.
type VPReflectSize struct{}

func (VPReflectSize) Type() VPType                   { return VPTypeSize }
func (VPReflectSize) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectSize) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		tmp := v.Int()
		return VPSize[int64]{}.PrintValue(out, noescape(&tmp))
	default:
		tmp := v.Uint()
		return VPSize[uint64]{}.PrintValue(out, noescape(&tmp))
	}
}

func (VPReflectSize) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	if !set {
		var tmp int64
		err = VPSize[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
		if err == nil {
			return
		}

		var tmpu uint64
		return VPSize[uint64]{}.ParseValue(opts, arg, noescape(&tmpu), set)
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var tmp int64
		err = VPSize[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
		if err != nil {
			return
		}

		if bits.Len64(uint64(tmp)) > typ.Bits() {
			return strconv.ErrRange
		}

		v.SetInt(tmp)
	default:
		var tmp uint64
		err = VPSize[uint64]{}.ParseValue(opts, arg, noescape(&tmp), set)
		if err != nil {
			return
		}
		if bits.Len64(tmp) > typ.Bits() {
			return strconv.ErrRange
		}
		v.SetUint(tmp)
	}

	return
}

// VPReflectDuration is the reflect version of VPDuration.
//
// It accepts arbitrary depth of pointers.
type VPReflectDuration struct{}

func (VPReflectDuration) Type() VPType                   { return VPTypeDuration }
func (VPReflectDuration) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectDuration) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		tmp := v.Int()
		return VPDuration[int64]{}.PrintValue(out, noescape(&tmp))
	default:
		tmp := v.Uint()
		return VPDuration[uint64]{}.PrintValue(out, noescape(&tmp))
	}
}

func (VPReflectDuration) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	if !set {
		var tmp int64
		err = VPDuration[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
		if err == nil {
			return
		}

		var tmpu uint64
		return VPDuration[uint64]{}.ParseValue(opts, arg, noescape(&tmpu), set)
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var tmp int64
		err = VPDuration[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
		if err != nil {
			return
		}
		if bits.Len64(uint64(tmp)) > typ.Bits() {
			return strconv.ErrRange
		}
		v.SetInt(tmp)
	default:
		var tmp uint64
		err = VPDuration[uint64]{}.ParseValue(opts, arg, noescape(&tmp), set)
		if err != nil {
			return
		}
		if bits.Len64(tmp) > typ.Bits() {
			return strconv.ErrRange
		}
		v.SetUint(tmp)
	}

	return
}

// VPReflectTime is the reflect version of VPTime.
//
// It accepts arbitrary depth of pointers.
type VPReflectTime struct{}

func (VPReflectTime) Type() VPType                   { return VPTypeTime }
func (VPReflectTime) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectTime) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Convert(reflect.TypeOf((*time.Time)(nil)).Elem()).Interface().(time.Time)
	return VPTime[time.Time]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectTime) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp time.Time
	err = VPTime[time.Time]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	_, v = prepareRValue(v.Type(), value, set)
	v.Set(reflect.ValueOf(noescape(&tmp)).Elem())
	return
}

// VPReflectUnixSec is the reflect version of VPUnixSec.
//
// It accepts arbitrary depth of pointers.
type VPReflectUnixSec struct{}

func (VPReflectUnixSec) Type() VPType                   { return VPTypeTime }
func (VPReflectUnixSec) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectUnixSec) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Int()
	return VPUnixSec[int64]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectUnixSec) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp int64
	err = VPUnixSec[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	if bits.Len64(uint64(tmp)) > typ.Bits() {
		return strconv.ErrRange
	}
	v.SetInt(tmp)
	return
}

// VPReflectUnixMilli is the reflect version of VPUnixMilli.
//
// It accepts arbitrary depth of pointers.
type VPReflectUnixMilli struct{}

func (VPReflectUnixMilli) Type() VPType                   { return VPTypeTime }
func (VPReflectUnixMilli) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectUnixMilli) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Int()
	return VPUnixMilli[int64]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectUnixMilli) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp int64
	err = VPUnixMilli[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	if bits.Len64(uint64(tmp)) > typ.Bits() {
		return strconv.ErrRange
	}
	v.SetInt(tmp)
	return
}

// VPReflectUnixMicro is the reflect version of VPUnixMicro.
//
// It accepts arbitrary depth of pointers.
type VPReflectUnixMicro struct{}

func (VPReflectUnixMicro) Type() VPType                   { return VPTypeTime }
func (VPReflectUnixMicro) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectUnixMicro) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Int()
	return VPUnixMicro[int64]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectUnixMicro) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp int64
	err = VPUnixMicro[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	if bits.Len64(uint64(tmp)) > typ.Bits() {
		return strconv.ErrRange
	}
	v.SetInt(tmp)
	return
}

// VPReflectUnixNano is the reflect version of VPUnixNano.
//
// It accepts arbitrary depth of pointers.
type VPReflectUnixNano struct{}

func (VPReflectUnixNano) Type() VPType                   { return VPTypeTime }
func (VPReflectUnixNano) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectUnixNano) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Int()
	return VPUnixNano[int64]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectUnixNano) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp int64
	err = VPUnixNano[int64]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	typ, v := prepareRValue(v.Type(), value, set)
	if bits.Len64(uint64(tmp)) > typ.Bits() {
		return strconv.ErrRange
	}
	v.SetInt(tmp)
	return
}

// VPReflectRegexp is the reflect version of VPRegexp.
//
// It accepts arbitrary depth of pointers.
type VPReflectRegexp struct{}

func (VPReflectRegexp) Type() VPType                   { return VPTypeRegexp }
func (VPReflectRegexp) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectRegexp) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Convert(reflect.TypeOf((*regexp.Regexp)(nil)).Elem()).Interface().(regexp.Regexp)
	return VPRegexp[regexp.Regexp]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectRegexp) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp regexp.Regexp
	err = VPRegexp[regexp.Regexp]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	_, v = prepareRValue(v.Type(), value, set)
	v.Set(reflect.ValueOf(noescape(&tmp)).Elem())
	return
}

// VPReflectRegexpNocase is the reflect version of VPRegexpNocase.
//
// It accepts arbitrary depth of pointers.
type VPReflectRegexpNocase struct{}

func (VPReflectRegexpNocase) Type() VPType                   { return VPTypeRegexp }
func (VPReflectRegexpNocase) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectRegexpNocase) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return 0, nil
	}

	tmp := v.Convert(reflect.TypeOf((*regexp.Regexp)(nil)).Elem()).Interface().(regexp.Regexp)
	return VPRegexpNocase[regexp.Regexp]{}.PrintValue(out, noescape(&tmp))
}

func (VPReflectRegexpNocase) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var tmp regexp.Regexp
	err = VPRegexpNocase[regexp.Regexp]{}.ParseValue(opts, arg, noescape(&tmp), set)
	if err != nil || !set {
		return
	}

	v := *value
	_, v = prepareRValue(v.Type(), value, set)
	v.Set(reflect.ValueOf(noescape(&tmp)).Elem())
	return
}

// VPReflectSum is the reflect version of VPSum.
//
// It accepts arbitrary depth of pointers.
type VPReflectSum[P VP[*reflect.Value]] struct{ VP P }

func (vp VPReflectSum[P]) Type() VPType {
	ret := vp.VP.Type()
	if ret&VPTypeVariantMASK != 0 {
		return VPTypeUnknown
	}

	return ret | VPTypeVariantSum
}

func (VPReflectSum[P]) HasValue(v *reflect.Value) bool { return reflectHasValue(v) }

func (VPReflectSum[P]) PrintValue(out io.Writer, value *reflect.Value) (int, error) {
	var peeker P
	return peeker.PrintValue(out, value)
}

func (vp VPReflectSum[P]) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	var (
		v    reflect.Value
		oldI int64
		oldU uint64
		oldF float64
		oldC complex128

		isSigned, isFloat, isComplex bool
	)

	if set {
		v = *value
		_, v = prepareRValue(v.Type(), value, true)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			oldI = v.Int()
			isSigned = true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			oldU = v.Uint()
		case reflect.Float32, reflect.Float64:
			oldF = v.Float()
			isFloat = true
		case reflect.Complex64, reflect.Complex128:
			oldC = v.Complex()
			isComplex = true
		}
	}

	err = vp.VP.ParseValue(opts, arg, value, set)
	if err != nil || !set {
		return
	}

	if isComplex {
		v.SetComplex(oldC + v.Complex())
	} else if isFloat {
		v.SetFloat(oldF + v.Float())
	} else if isSigned {
		v.SetInt(oldI + v.Int())
	} else {
		v.SetUint(oldU + v.Uint())
	}
	return
}

// VPReflectSlice is the reflect version of VPSlice.
//
// It accepts arbitrary depth of pointers.
type VPReflectSlice[EP VP[*reflect.Value]] struct{ Elem EP }

func (vp VPReflectSlice[EP]) Type() VPType {
	ret := vp.Elem.Type()
	if ret&VPTypeVariantMASK != 0 {
		return VPTypeUnknown
	}

	return ret | VPTypeVariantSlice
}

func (vp VPReflectSlice[EP]) HasValue(value *reflect.Value) bool {
	v, ok := reflectBaseValue(value)
	if !ok {
		return false
	}
	return !v.IsZero() && v.Len() != 0
}

func (vp VPReflectSlice[EP]) PrintValue(out io.Writer, value *reflect.Value) (n int, err error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return
	}

	n, err = wstr(out, "[")
	if err != nil {
		return
	}

	var (
		x    int
		item reflect.Value
	)
	for i, sz := 0, v.Len(); i < sz; i++ {
		if i != 0 {
			x, err = wstr(out, ", ")
			n += x
			if err != nil {
				return
			}
		}

		item = v.Index(i)
		x, err = vp.Elem.PrintValue(out, noescape(&item))
		n += x
		if err != nil {
			return
		}
	}

	x, err = wstr(out, "]")
	n += x
	return
}

func (vp VPReflectSlice[EP]) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	if !set {
		var val reflect.Value
		return vp.Elem.ParseValue(opts, arg, noescape(&val), false)
	}

	v := *value
	_, slice := prepareRValue(v.Type(), value, set)
	n := slice.Len()
	if n == slice.Cap() {
		slice.Grow(1)
	}
	slice.SetLen(n + 1)
	val := slice.Index(n)
	err = vp.Elem.ParseValue(opts, arg, noescape(&val), true)
	if err != nil {
		slice.SetLen(n)
	}

	return nil
}

// VPReflectMap is the reflect version of VPMap.
//
// It accepts arbitrary depth of pointers.
type VPReflectMap[K, V VP[*reflect.Value]] struct {
	Key  K
	Elem V
}

func (vp VPReflectMap[K, V]) Type() VPType {
	kt := vp.Key.Type()
	if kt&VPTypeVariantMASK != 0 { // key can only be scalars (not including sum)
		return VPTypeUnknown
	}

	vt := vp.Elem.Type()
	if vt&VPTypeVariantMASK == VPTypeVariantMap { // value can only be slice or scalar (including sum)
		return VPTypeUnknown
	}

	return VPTypeVariantMap |
		kt<<VPTypeKeyScalarShift |
		(vt&VPTypeVariantMASK)<<(VPTypeMapElemVariantShift-VPTypeVariantShift) |
		vt&VPTypeElemScalarMASK
}

func (vp VPReflectMap[K, V]) HasValue(value *reflect.Value) bool {
	v, ok := reflectBaseValue(value)
	if !ok {
		return false
	}
	return !v.IsZero() && v.Len() != 0
}

func (vp VPReflectMap[K, V]) PrintValue(out io.Writer, value *reflect.Value) (n int, err error) {
	v, ok := reflectBaseValue(value)
	if !ok {
		return
	}

	var (
		x int

		key, val   reflect.Value
		iter       = v.MapRange()
		afterFirst bool
	)

	n, err = wstr(out, "[")
	if err != nil {
		return
	}

	for iter.Next() {
		if afterFirst {
			x, err = wstr(out, ", ")
			n += x
			if err != nil {
				return
			}
		}

		key.SetIterKey(iter)
		x, err = vp.Key.PrintValue(out, noescape(&key))
		n += x
		if err != nil {
			return
		}

		x, err = wstr(out, "=")
		n += x
		if err != nil {
			return
		}

		val.SetIterValue(iter)
		x, err = vp.Elem.PrintValue(out, noescape(&val))
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

func (vp VPReflectMap[K, V]) ParseValue(opts *ParseOptions, arg string, value *reflect.Value, set bool) (err error) {
	strKey, strVal, ok := strings.Cut(arg, "=")
	if !ok {
		return &ErrInvalidValue{
			Type:  "map",
			Value: arg,
		}
	}

	var (
		key, val reflect.Value
		mapval   reflect.Value
		maptyp   reflect.Type
	)

	if set {
		mapval = *value
		maptyp, mapval = prepareRValue(mapval.Type(), value, set)
		key = reflect.New(maptyp.Key()).Elem()
	}

	err = vp.Key.ParseValue(opts, strKey, noescape(&key), set)
	if err != nil {
		return
	}

	if set {
		val = reflect.New(maptyp.Elem()).Elem()
		tmp := mapval.MapIndex(key)
		if tmp.IsValid() {
			val.Set(tmp)
		}
	}

	err = vp.Elem.ParseValue(opts, strVal, noescape(&val), set)
	if err != nil || !set {
		return
	}

	mapval.SetMapIndex(key, val)
	return nil
}

func prepareRValue(typ reflect.Type, value *reflect.Value, set bool) (reflect.Type, reflect.Value) {
	var val reflect.Value
	if value != nil {
		val = *value
	}

	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		if set {
			if !val.IsValid() {
				val = reflect.New(typ)
				*value = val
			} else if val.IsNil() {
				val.Set(reflect.New(typ))
			}

			val = val.Elem()
		}
	}

	if set && (!val.IsValid() || val.IsZero()) {
		switch typ.Kind() {
		case reflect.Slice:
			val.Set(reflect.MakeSlice(typ, 0, 2))
		case reflect.Map:
			val.Set(reflect.MakeMap(typ))
		}
	}

	return typ, val
}

func reflectBaseValue(value *reflect.Value) (v reflect.Value, ok bool) {
	if value == nil {
		return
	}

	v = *value
	if !v.IsValid() {
		return
	}

	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}

		v = v.Elem()
	}

	return v, v.IsValid()
}

func reflectHasValue(value *reflect.Value) bool {
	v, ok := reflectBaseValue(value)
	return ok && !v.IsZero()
}
