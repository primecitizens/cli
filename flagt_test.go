// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/primecitizens/cli/internal/assert"
)

func TestFlagTypes(t *testing.T) {
	var (
		TRUE = true
		Str  = "str"

		I   int   = -123
		I8  int8  = -123
		I16 int16 = -123
		I32 int32 = -123
		I64 int64 = -123

		U    uint    = 123
		U8   uint8   = 123
		U16  uint16  = 123
		U32  uint32  = 123
		U64  uint64  = 123
		Uptr uintptr = 123

		F32 float32 = -123
		F64 float64 = -123

		SliceStr  = []string{Str, Str}
		SliceTrue = []bool{TRUE, TRUE}

		SliceI   = []int{-123, -123}
		SliceI8  = []int8{-123, -123}
		SliceI16 = []int16{-123, -123}
		SliceI32 = []int32{-123, -123}
		SliceI64 = []int64{-123, -123}

		SliceU    = []uint{123, 123}
		SliceU8   = []uint8{123, 123}
		SliceU16  = []uint16{123, 123}
		SliceU32  = []uint32{123, 123}
		SliceU64  = []uint64{123, 123}
		SliceUptr = []uintptr{123, 123}

		SliceF32 = []float32{-123, -123}
		SliceF64 = []float64{-123, -123}

		Sz      int64 = 1024*1024 + 1024
		SliceSz       = []int64{Sz, Sz}

		Dur      time.Duration = time.Minute + time.Second
		SliceDur               = []time.Duration{Dur, Dur}
	)

	flagTypeTests := []struct {
		flag                   Flag
		typ, value, implyValue string
	}{
		// scalar types
		{&String{Value: &Str}, "str", "str", ""},
		{&Bool{Value: &TRUE}, "bool", "true", "true"},
		{&Int{Value: &I}, "int", "-123", ""},
		{&Int8{Value: &I8}, "int", "-123", ""},
		{&Int16{Value: &I16}, "int", "-123", ""},
		{&Int32{Value: &I32}, "int", "-123", ""},
		{&Int64{Value: &I64}, "int", "-123", ""},
		{&Uint{Value: &U}, "uint", "123", ""},
		{&Uint8{Value: &U8}, "uint", "123", ""},
		{&Uint16{Value: &U16}, "uint", "123", ""},
		{&Uint32{Value: &U32}, "uint", "123", ""},
		{&Uint64{Value: &U64}, "uint", "123", ""},
		{&Uintptr{Value: &Uptr}, "uint", "123", ""},
		{&Float32{Value: &F32}, "float", "-123", ""},
		{&Float64{Value: &F64}, "float", "-123", ""},
		{&StringV{Value: Str}, "str", "str", ""},
		{&BoolV{Value: TRUE}, "bool", "true", "true"},
		{&IntV{Value: I}, "int", "-123", ""},
		{&Int8V{Value: I8}, "int", "-123", ""},
		{&Int16V{Value: I16}, "int", "-123", ""},
		{&Int32V{Value: I32}, "int", "-123", ""},
		{&Int64V{Value: I64}, "int", "-123", ""},
		{&UintV{Value: U}, "uint", "123", ""},
		{&Uint8V{Value: U8}, "uint", "123", ""},
		{&Uint16V{Value: U16}, "uint", "123", ""},
		{&Uint32V{Value: U32}, "uint", "123", ""},
		{&Uint64V{Value: U64}, "uint", "123", ""},
		{&UintptrV{Value: Uptr}, "uint", "123", ""},
		{&Float32V{Value: F32}, "float", "-123", ""},
		{&Float64V{Value: F64}, "float", "-123", ""},
		{&Size{Value: &Sz}, "size", "1MB1KB", ""},
		{&Duration{Value: &Dur}, "dur", "1m1s", ""},
		{&Time{}, "time", "", ""},
		{&UnixSec{}, "time", "", ""},
		{&UnixMilli{}, "time", "", ""},
		{&UnixNano{}, "time", "", ""},
		{&Regexp{}, "regexp", "", ""},
		{&RegexpNocase{}, "regexp", "", ""},
		{&SizeV{Value: Sz}, "size", "1MB1KB", ""},
		{&DurationV{Value: Dur}, "dur", "1m1s", ""},
		{&TimeV{}, "time", "", ""},
		{&UnixSecV{}, "time", "", ""},
		{&UnixMilliV{}, "time", "", ""},
		{&UnixNanoV{}, "time", "", ""},
		{&RegexpV{}, "regexp", "", ""},
		{&RegexpNocaseV{}, "regexp", "", ""},

		// slice types
		{&StringSlice{Value: &SliceStr}, "[]str", "[str, str]", ""},
		{&BoolSlice{Value: &SliceTrue}, "[]bool", "[true, true]", "true"},
		{&IntSlice{Value: &SliceI}, "[]int", "[-123, -123]", ""},
		{&Int8Slice{Value: &SliceI8}, "[]int", "[-123, -123]", ""},
		{&Int16Slice{Value: &SliceI16}, "[]int", "[-123, -123]", ""},
		{&Int32Slice{Value: &SliceI32}, "[]int", "[-123, -123]", ""},
		{&Int64Slice{Value: &SliceI64}, "[]int", "[-123, -123]", ""},
		{&UintSlice{Value: &SliceU}, "[]uint", "[123, 123]", ""},
		{&Uint8Slice{Value: &SliceU8}, "[]uint", "[123, 123]", ""},
		{&Uint16Slice{Value: &SliceU16}, "[]uint", "[123, 123]", ""},
		{&Uint32Slice{Value: &SliceU32}, "[]uint", "[123, 123]", ""},
		{&Uint64Slice{Value: &SliceU64}, "[]uint", "[123, 123]", ""},
		{&UintptrSlice{Value: &SliceUptr}, "[]uint", "[123, 123]", ""},
		{&Float32Slice{Value: &SliceF32}, "[]float", "[-123, -123]", ""},
		{&Float64Slice{Value: &SliceF64}, "[]float", "[-123, -123]", ""},
		{&StringSliceV{Value: SliceStr}, "[]str", "[str, str]", ""},
		{&BoolSliceV{Value: SliceTrue}, "[]bool", "[true, true]", "true"},
		{&IntSliceV{Value: SliceI}, "[]int", "[-123, -123]", ""},
		{&Int8SliceV{Value: SliceI8}, "[]int", "[-123, -123]", ""},
		{&Int16SliceV{Value: SliceI16}, "[]int", "[-123, -123]", ""},
		{&Int32SliceV{Value: SliceI32}, "[]int", "[-123, -123]", ""},
		{&Int64SliceV{Value: SliceI64}, "[]int", "[-123, -123]", ""},
		{&UintSliceV{Value: SliceU}, "[]uint", "[123, 123]", ""},
		{&Uint8SliceV{Value: SliceU8}, "[]uint", "[123, 123]", ""},
		{&Uint16SliceV{Value: SliceU16}, "[]uint", "[123, 123]", ""},
		{&Uint32SliceV{Value: SliceU32}, "[]uint", "[123, 123]", ""},
		{&Uint64SliceV{Value: SliceU64}, "[]uint", "[123, 123]", ""},
		{&UintptrSliceV{Value: SliceUptr}, "[]uint", "[123, 123]", ""},
		{&Float32SliceV{Value: SliceF32}, "[]float", "[-123, -123]", ""},
		{&Float64SliceV{Value: SliceF64}, "[]float", "[-123, -123]", ""},
		{&SizeSlice{Value: &SliceSz}, "[]size", "[1MB1KB, 1MB1KB]", ""},
		{&DurationSlice{Value: &SliceDur}, "[]dur", "[1m1s, 1m1s]", ""},
		{&TimeSlice{}, "[]time", "", ""},
		{&UnixSecSlice{}, "[]time", "", ""},
		{&UnixMilliSlice{}, "[]time", "", ""},
		{&UnixNanoSlice{}, "[]time", "", ""},
		{&RegexpSlice{}, "[]regexp", "", ""},
		{&RegexpNocaseSlice{}, "[]regexp", "", ""},
		{&SizeSliceV{Value: SliceSz}, "[]size", "[1MB1KB, 1MB1KB]", ""},
		{&DurationSliceV{Value: SliceDur}, "[]dur", "[1m1s, 1m1s]", ""},
		{&TimeSliceV{}, "[]time", "", ""},
		{&UnixSecSliceV{}, "[]time", "", ""},
		{&UnixMilliSliceV{}, "[]time", "", ""},
		{&UnixNanoSliceV{}, "[]time", "", ""},
		{&RegexpSliceV{}, "[]regexp", "", ""},
		{&RegexpNocaseSliceV{}, "[]regexp", "", ""},

		// sum types
		{&IntSum{Value: &I}, "isum", "-123", "1"},
		{&Int8Sum{Value: &I8}, "isum", "-123", "1"},
		{&Int16Sum{Value: &I16}, "isum", "-123", "1"},
		{&Int32Sum{Value: &I32}, "isum", "-123", "1"},
		{&Int64Sum{Value: &I64}, "isum", "-123", "1"},
		{&UintSum{Value: &U}, "usum", "123", "1"},
		{&Uint8Sum{Value: &U8}, "usum", "123", "1"},
		{&Uint16Sum{Value: &U16}, "usum", "123", "1"},
		{&Uint32Sum{Value: &U32}, "usum", "123", "1"},
		{&Uint64Sum{Value: &U64}, "usum", "123", "1"},
		{&UintptrSum{Value: &Uptr}, "usum", "123", "1"},
		{&Float32Sum{Value: &F32}, "fsum", "-123", "1"},
		{&Float64Sum{Value: &F64}, "fsum", "-123", "1"},
		{&IntSumV{Value: I}, "isum", "-123", "1"},
		{&Int8SumV{Value: I8}, "isum", "-123", "1"},
		{&Int16SumV{Value: I16}, "isum", "-123", "1"},
		{&Int32SumV{Value: I32}, "isum", "-123", "1"},
		{&Int64SumV{Value: I64}, "isum", "-123", "1"},
		{&UintSumV{Value: U}, "usum", "123", "1"},
		{&Uint8SumV{Value: U8}, "usum", "123", "1"},
		{&Uint16SumV{Value: U16}, "usum", "123", "1"},
		{&Uint32SumV{Value: U32}, "usum", "123", "1"},
		{&Uint64SumV{Value: U64}, "usum", "123", "1"},
		{&UintptrSumV{Value: Uptr}, "usum", "123", "1"},
		{&Float32SumV{Value: F32}, "fsum", "-123", "1"},
		{&Float64SumV{Value: F64}, "fsum", "-123", "1"},
		{&SizeSum{Value: &Sz}, "ssum", "1MB1KB", "1"},
		{&DurationSum{Value: &Dur}, "dsum", "1m1s", "1s"},
		{&SizeSumV{Value: Sz}, "ssum", "1MB1KB", "1"},
		{&DurationSumV{Value: Dur}, "dsum", "1m1s", "1s"},

		// map[string]*
		{&MapStringStringV{Value: map[string]string{Str: Str}}, "map[str]str", "[str=str]", ""},
		{&MapStringBoolV{Value: map[string]bool{Str: TRUE}}, "map[str]bool", "[str=true]", ""},
		{&MapStringIntV{Value: map[string]int{Str: I}}, "map[str]int", "[str=-123]", ""},
		{&MapStringInt8V{Value: map[string]int8{Str: I8}}, "map[str]int", "[str=-123]", ""},
		{&MapStringInt16V{Value: map[string]int16{Str: I16}}, "map[str]int", "[str=-123]", ""},
		{&MapStringInt32V{Value: map[string]int32{Str: I32}}, "map[str]int", "[str=-123]", ""},
		{&MapStringInt64V{Value: map[string]int64{Str: I64}}, "map[str]int", "[str=-123]", ""},
		{&MapStringUintV{Value: map[string]uint{Str: U}}, "map[str]uint", "[str=123]", ""},
		{&MapStringUint8V{Value: map[string]uint8{Str: U8}}, "map[str]uint", "[str=123]", ""},
		{&MapStringUint16V{Value: map[string]uint16{Str: U16}}, "map[str]uint", "[str=123]", ""},
		{&MapStringUint32V{Value: map[string]uint32{Str: U32}}, "map[str]uint", "[str=123]", ""},
		{&MapStringUint64V{Value: map[string]uint64{Str: U64}}, "map[str]uint", "[str=123]", ""},
		{&MapStringUintptrV{Value: map[string]uintptr{Str: Uptr}}, "map[str]uint", "[str=123]", ""},
		{&MapStringFloat32V{Value: map[string]float32{Str: F32}}, "map[str]float", "[str=-123]", ""},
		{&MapStringFloat64V{Value: map[string]float64{Str: F64}}, "map[str]float", "[str=-123]", ""},
		{&MapStringSizeV{Value: map[string]int64{Str: Sz}}, "map[str]size", "[str=1MB1KB]", ""},
		{&MapStringDurationV{Value: map[string]time.Duration{Str: Dur}}, "map[str]dur", "[str=1m1s]", ""},
		{&MapStringTimeV{Value: map[string]time.Time{}}, "map[str]time", "", ""},
		{&MapStringUnixSecV{Value: map[string]int64{}}, "map[str]time", "", ""},
		{&MapStringUnixMilliV{Value: map[string]int64{}}, "map[str]time", "", ""},
		{&MapStringUnixMicroV{Value: map[string]int64{}}, "map[str]time", "", ""},
		{&MapStringUnixNanoV{Value: map[string]int64{}}, "map[str]time", "", ""},
		{&MapStringRegexpV{Value: map[string]*regexp.Regexp{}}, "map[str]regexp", "", ""},
		{&MapStringRegexpNocaseV{Value: map[string]*regexp.Regexp{}}, "map[str]regexp", "", ""},

		{&MapStringStringSliceV{Value: map[string][]string{Str: SliceStr}}, "map[str][]str", "[str=[str, str]]", ""},
		{&MapStringBoolSliceV{Value: map[string][]bool{Str: SliceTrue}}, "map[str][]bool", "[str=[true, true]]", ""},
		{&MapStringIntSliceV{Value: map[string][]int{Str: SliceI}}, "map[str][]int", "[str=[-123, -123]]", ""},
		{&MapStringInt8SliceV{Value: map[string][]int8{Str: SliceI8}}, "map[str][]int", "[str=[-123, -123]]", ""},
		{&MapStringInt16SliceV{Value: map[string][]int16{Str: SliceI16}}, "map[str][]int", "[str=[-123, -123]]", ""},
		{&MapStringInt32SliceV{Value: map[string][]int32{Str: SliceI32}}, "map[str][]int", "[str=[-123, -123]]", ""},
		{&MapStringInt64SliceV{Value: map[string][]int64{Str: SliceI64}}, "map[str][]int", "[str=[-123, -123]]", ""},
		{&MapStringUintSliceV{Value: map[string][]uint{Str: SliceU}}, "map[str][]uint", "[str=[123, 123]]", ""},
		{&MapStringUint8SliceV{Value: map[string][]uint8{Str: SliceU8}}, "map[str][]uint", "[str=[123, 123]]", ""},
		{&MapStringUint16SliceV{Value: map[string][]uint16{Str: SliceU16}}, "map[str][]uint", "[str=[123, 123]]", ""},
		{&MapStringUint32SliceV{Value: map[string][]uint32{Str: SliceU32}}, "map[str][]uint", "[str=[123, 123]]", ""},
		{&MapStringUint64SliceV{Value: map[string][]uint64{Str: SliceU64}}, "map[str][]uint", "[str=[123, 123]]", ""},
		{&MapStringUintptrSliceV{Value: map[string][]uintptr{Str: SliceUptr}}, "map[str][]uint", "[str=[123, 123]]", ""},
		{&MapStringFloat32SliceV{Value: map[string][]float32{Str: SliceF32}}, "map[str][]float", "[str=[-123, -123]]", ""},
		{&MapStringFloat64SliceV{Value: map[string][]float64{Str: SliceF64}}, "map[str][]float", "[str=[-123, -123]]", ""},
		{&MapStringSizeSliceV{Value: map[string][]int64{Str: SliceSz}}, "map[str][]size", "[str=[1MB1KB, 1MB1KB]]", ""},
		{&MapStringDurationSliceV{Value: map[string][]time.Duration{Str: SliceDur}}, "map[str][]dur", "[str=[1m1s, 1m1s]]", ""},
		{&MapStringTimeSliceV{Value: map[string][]time.Time{Str: {}}}, "map[str][]time", "[str=[]]", ""},
		{&MapStringUnixSecSliceV{Value: map[string][]int64{Str: {}}}, "map[str][]time", "[str=[]]", ""},
		{&MapStringUnixMilliSliceV{Value: map[string][]int64{Str: {}}}, "map[str][]time", "[str=[]]", ""},
		{&MapStringUnixMicroSliceV{Value: map[string][]int64{Str: {}}}, "map[str][]time", "[str=[]]", ""},
		{&MapStringUnixNanoSliceV{Value: map[string][]int64{Str: {}}}, "map[str][]time", "[str=[]]", ""},
		{&MapStringRegexpSliceV{Value: map[string][]*regexp.Regexp{Str: {}}}, "map[str][]regexp", "[str=[]]", ""},
		{&MapStringRegexpNocaseSliceV{Value: map[string][]*regexp.Regexp{Str: {}}}, "map[str][]regexp", "[str=[]]", ""},

		{&MapStringIntSumV{Value: map[string]int{Str: I}}, "map[str]isum", "[str=-123]", ""},
		{&MapStringInt8SumV{Value: map[string]int8{Str: I8}}, "map[str]isum", "[str=-123]", ""},
		{&MapStringInt16SumV{Value: map[string]int16{Str: I16}}, "map[str]isum", "[str=-123]", ""},
		{&MapStringInt32SumV{Value: map[string]int32{Str: I32}}, "map[str]isum", "[str=-123]", ""},
		{&MapStringInt64SumV{Value: map[string]int64{Str: I64}}, "map[str]isum", "[str=-123]", ""},
		{&MapStringUintSumV{Value: map[string]uint{Str: U}}, "map[str]usum", "[str=123]", ""},
		{&MapStringUint8SumV{Value: map[string]uint8{Str: U8}}, "map[str]usum", "[str=123]", ""},
		{&MapStringUint16SumV{Value: map[string]uint16{Str: U16}}, "map[str]usum", "[str=123]", ""},
		{&MapStringUint32SumV{Value: map[string]uint32{Str: U32}}, "map[str]usum", "[str=123]", ""},
		{&MapStringUint64SumV{Value: map[string]uint64{Str: U64}}, "map[str]usum", "[str=123]", ""},
		{&MapStringUintptrSumV{Value: map[string]uintptr{Str: Uptr}}, "map[str]usum", "[str=123]", ""},
		{&MapStringFloat32SumV{Value: map[string]float32{Str: F32}}, "map[str]fsum", "[str=-123]", ""},
		{&MapStringFloat64SumV{Value: map[string]float64{Str: F64}}, "map[str]fsum", "[str=-123]", ""},
		{&MapStringSizeSumV{Value: map[string]int64{Str: Sz}}, "map[str]ssum", "[str=1MB1KB]", ""},
		{&MapStringDurationSumV{Value: map[string]time.Duration{Str: Dur}}, "map[str]dsum", "[str=1m1s]", ""},
	}

	for _, test := range flagTypeTests {
		t.Run(test.typ, func(t *testing.T) {
			assert.Eq(t, "", test.flag.Usage())
			assert.Eq(t, 0, test.flag.State())
			assert.Eq(t, nil, test.flag.Extra())

			typ, ok := test.flag.Type()
			if len(test.typ) != 0 {
				assert.True(t, ok)
				assert.Eq(t, test.typ, typ)
			} else {
				assert.False(t, ok)
				assert.Eq(t, "", typ)
			}

			ok = test.flag.HasValue()
			if len(test.value) != 0 {
				assert.True(t, ok)

				var buf strings.Builder
				n, err := test.flag.PrintValue(&buf)
				assert.NoError(t, err)
				assert.Eq(t, test.value, buf.String())
				assert.True(t, n > 0)
			} else {
				assert.False(t, ok)
			}

			iv, ok := test.flag.ImplyValue()
			if len(test.implyValue) != 0 {
				assert.True(t, ok)
				assert.Eq(t, test.implyValue, iv)
			} else {
				assert.False(t, ok)
				assert.Eq(t, "", iv)
			}
		})
	}
}
