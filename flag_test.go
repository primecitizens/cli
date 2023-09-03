// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"regexp"
	"testing"
	"time"

	"github.com/primecitizens/cli/internal/assert"
)

func TestIsShorthand(t *testing.T) {
	for _, test := range []struct {
		s        string
		expected bool
	}{
		{"", false},
		{"s", true},
		{"ss", false},
		{"🥳", true},
		{"🥵", true},
	} {
		t.Run(test.s, func(t *testing.T) {
			assert.Eq(t, test.expected, IsShorthand(test.s))
		})
	}
}

func TestParseFlags_Limitations(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string

		good FlagTestOptions
		bad  error
	}{
		{
			name: "Implicit flag cannot be followed by value prefixed with hyphen",
			args: []string{"--IntSum", "-1"},
			bad: &ErrAmbiguousArgs{
				Name:  "IntSum",
				Value: "-1",
			},
		},
		{
			name: "Implicit flag cannot be followed by valid value prefixed with hyphen (shorthand)",
			args: []string{"-V", "-1"},
			bad: &ErrAmbiguousArgs{
				Name:  "V",
				Value: "-1",
			},
		},
		{
			name: "Workaround implicit flag limitation",
			args: []string{"-V=-1", "--IntSum=-1"},
			good: FlagTestOptions{IntSum: -2},
		},

		{
			name: "Standalone dash cannot be flag value",
			args: []string{"--String", "--"},
			bad: &ErrFlagValueMissing{
				Name: "String",
				At:   0,
			},
		},
		{
			name: "Workaround dash limitation",
			args: []string{"--String=--"},
			good: FlagTestOptions{String: "--"},
		},
	} {
		var (
			actual FlagTestOptions
			flags  TestFlags
		)

		flags.Bind(&actual)
		posArgs, dashArgs, err := ParseFlags(test.args, &flags, nil)
		assert.Eq(t, len(posArgs), 0)
		assert.Eq(t, len(dashArgs), 0)
		assertOptsEq(t, test.good, actual)
		if test.bad != nil {
			assert.ErrorIs(t, test.bad, err)
		}
	}
}

func TestParseFlags(t *testing.T) {
	const (
		durStr  = "14d"
		szStr   = "15GB"
		timeStr = "19:01:39"

		expectedDuration = 14 * 24 * time.Hour
		expectedSize     = 15 * 1024 * 1024 * 1024
	)

	opts := ParseOptions{
		HelpArgs:  []string{}, // no help flags
		StartTime: time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
		HandleParseError: func(opts *ParseOptions, args []string, i int, err error) error {
			if !assert.NoError(t, err) {
				t.Log(args[i], err)
			}
			return err
		},
	}

	expectedTime := opts.StartTime.Add(19*time.Hour + time.Minute + 39*time.Second)

	expected := FlagTestOptions{
		String:    "-",
		Bool:      true,
		Int:       'c',
		Int8:      'd',
		Int16:     -'e',
		Int32:     -'f',
		Int64:     'g',
		Uint:      'h',
		Uint8:     'i',
		Uint16:    'j',
		Uint32:    'k',
		Uint64:    'l',
		Uintptr:   'm',
		Float32:   32.32,
		Float64:   64.64,
		Duration:  expectedDuration,
		Time:      expectedTime,
		UnixSec:   expectedTime.Unix(),
		UnixMilli: expectedTime.UnixMilli(),
		UnixNano:  expectedTime.UnixNano(),
		Size:      expectedSize,

		StringSlice:    []string{"String", "String", "String"},
		BoolSlice:      []bool{true, false, true},
		IntSlice:       []int{'C', 'C', 'C'},
		Int8Slice:      []int8{'D', 'D', 'D'},
		Int16Slice:     []int16{-'E', -'E', -'E'},
		Int32Slice:     []int32{-'F', -'F', -'F'},
		Int64Slice:     []int64{'G', 'G', 'G'},
		UintSlice:      []uint{'H', 'H', 'H'},
		Uint8Slice:     []uint8{'I', 'I', 'I'},
		Uint16Slice:    []uint16{'J', 'J', 'J'},
		Uint32Slice:    []uint32{'K', 'K', 'K'},
		Uint64Slice:    []uint64{'L', 'L', 'L'},
		UintptrSlice:   []uintptr{'M', 'M', 'M'},
		Float32Slice:   []float32{32.32, 32.32, 32.32},
		Float64Slice:   []float64{64.64, 64.64, 64.64},
		DurationSlice:  []time.Duration{expectedDuration, expectedDuration, expectedDuration},
		TimeSlice:      []time.Time{expectedTime, expectedTime, expectedTime},
		SizeSlice:      []int64{expectedSize, expectedSize, expectedSize},
		UnixSecSlice:   []int64{expectedTime.Unix(), expectedTime.Unix(), expectedTime.Unix()},
		UnixMilliSlice: []int64{expectedTime.UnixMilli(), expectedTime.UnixMilli(), expectedTime.UnixMilli()},
		UnixNanoSlice:  []int64{expectedTime.UnixNano(), expectedTime.UnixNano(), expectedTime.UnixNano()},

		IntSum:      5,
		Int8Sum:     5,
		Int16Sum:    5,
		Int32Sum:    5,
		Int64Sum:    5,
		UintSum:     5,
		Uint8Sum:    5,
		Uint16Sum:   5,
		Uint32Sum:   5,
		Uint64Sum:   5,
		UintptrSum:  5,
		Float32Sum:  5,
		Float64Sum:  5,
		DurationSum: 5 * time.Second,
		SizeSum:     5,
	}

	for _, test := range []struct {
		name     string
		args     []string
		posArgs  []string
		dashArgs []string
	}{
		{
			name: "long flags",
			args: []string{
				"-", "-",
				"--String", "-",
				"--Bool", "true",
				"--Int", "99",
				"--Int8", "100",
				"--Int16", "-101",
				"--Int32", "-102",
				"--Int64", "103",
				"--Uint", "104",
				"--Uint8", "105",
				"--Uint16", "106",
				"--Uint32", "107",
				"--Uint64", "108",
				"--Uintptr", "109",
				"--Float32", "32.32",
				"--Float64", "64.64",
				"--Duration", durStr,
				"--Time", timeStr,
				"--UnixSec", timeStr,
				"--UnixMilli", timeStr,
				"--UnixNano", timeStr,
				"--Size", szStr,

				"--StringSlice", "String", "--StringSlice", "String", "--StringSlice", "String",
				"--BoolSlice", "true", "--BoolSlice", "false", "--BoolSlice", "true",
				"--IntSlice", "67", "--IntSlice", "67", "--IntSlice", "67",
				"--Int8Slice", "68", "--Int8Slice", "68", "--Int8Slice", "68",
				"--Int16Slice", "-69", "--Int16Slice", "-69", "--Int16Slice", "-69",
				"--Int32Slice", "-70", "--Int32Slice", "-70", "--Int32Slice", "-70",
				"--Int64Slice", "71", "--Int64Slice", "71", "--Int64Slice", "71",
				"--UintSlice", "72", "--UintSlice", "72", "--UintSlice", "72",
				"--Uint8Slice", "73", "--Uint8Slice", "73", "--Uint8Slice", "73",
				"--Uint16Slice", "74", "--Uint16Slice", "74", "--Uint16Slice", "74",
				"--Uint32Slice", "75", "--Uint32Slice", "75", "--Uint32Slice", "75",
				"--Uint64Slice", "76", "--Uint64Slice", "76", "--Uint64Slice", "76",
				"--UintptrSlice", "77", "--UintptrSlice", "77", "--UintptrSlice", "77",
				"--Float32Slice", "32.32", "--Float32Slice", "32.32", "--Float32Slice", "32.32",
				"--Float64Slice", "64.64", "--Float64Slice", "64.64", "--Float64Slice", "64.64",
				"--DurationSlice", durStr, "--DurationSlice", durStr, "--DurationSlice", durStr,
				"--TimeSlice", timeStr, "--TimeSlice", timeStr, "--TimeSlice", timeStr,
				"--UnixSecSlice", timeStr, "--UnixSecSlice", timeStr, "--UnixSecSlice", timeStr,
				"--UnixMilliSlice", timeStr, "--UnixMilliSlice", timeStr, "--UnixMilliSlice", timeStr,
				"--UnixNanoSlice", timeStr, "--UnixNanoSlice", timeStr, "--UnixNanoSlice", timeStr,
				"--SizeSlice", szStr, "--SizeSlice", szStr, "--SizeSlice", szStr,

				"--IntSum", "--IntSum", "--IntSum", "3",
				"--Int8Sum", "--Int8Sum", "--Int8Sum", "3",
				"--Int16Sum", "--Int16Sum", "--Int16Sum", "3",
				"--Int32Sum", "--Int32Sum", "--Int32Sum", "3",
				"--Int64Sum", "--Int64Sum", "--Int64Sum", "3",
				"--UintSum", "--UintSum", "--UintSum", "3",
				"--Uint8Sum", "--Uint8Sum", "--Uint8Sum", "3",
				"--Uint16Sum", "--Uint16Sum", "--Uint16Sum", "3",
				"--Uint32Sum", "--Uint32Sum", "--Uint32Sum", "3",
				"--Uint64Sum", "--Uint64Sum", "--Uint64Sum", "3",
				"--UintptrSum", "--UintptrSum", "--UintptrSum", "3",
				"--Float32Sum", "--Float32Sum", "--Float32Sum", "3",
				"--Float64Sum", "--Float64Sum", "--Float64Sum", "3",
				"--DurationSum", "--DurationSum", "--DurationSum", "3s",
				"--SizeSum", "--SizeSum", "--SizeSum", "2", "--SizeSum",

				"--Empty=true",
				"--Regexp", "^foo",
				"--RegexpNocase", "^Foo",
				"--RegexpSlice", "^foo", "--RegexpSlice", "^foo", "--RegexpSlice", "^foo",
				"--RegexpNocaseSlice", "^Foo", "--RegexpNocaseSlice", "^Foo", "--RegexpNocaseSlice", "^Foo",

				"--", "-", "-",
			},

			posArgs:  []string{"-", "-"},
			dashArgs: []string{"-", "-"},
		},
		{
			name: "short flags",
			args: []string{
				"-", "-",
				"-a", "-",
				"-b",
				"-c", "99",
				"-d", "100",
				"-e", "-101",
				"-f", "-102",
				"-g", "103",
				"-h", "104",
				"-i", "105",
				"-j", "106",
				"-k", "107",
				"-l", "108",
				"-m", "109",
				"-n32.32",
				"-o", "64.64",
				"-p", durStr,
				"-q", timeStr,
				"-r", timeStr,
				"-s", timeStr,
				"-t", timeStr,
				"-u", szStr,

				"-YA", "String", "-A=String", "-A", "String",
				"-BB", "false", "-B", "true",
				"-C67", "-C", "67", "-C", "67",
				"-D", "68", "-D", "68", "-D", "68",
				"-E", "-69", "-E", "-69", "-E", "-69",
				"-F", "-70", "-F", "-70", "-F", "-70",
				"-G", "71", "-G", "71", "-G", "71",
				"-H", "72", "-H", "72", "-H", "72",
				"-I", "73", "-I", "73", "-I", "73",
				"-J", "74", "-J", "74", "-J", "74",
				"-K", "75", "-K", "75", "-K", "75",
				"-L", "76", "-L", "76", "-L", "76",
				"-M", "77", "-M", "77", "-M", "77",
				"-N", "32.32", "-N", "32.32", "-N", "32.32",
				"-O", "64.64", "-O", "64.64", "-O", "64.64",
				"-P", durStr, "-P", durStr, "-P", durStr,
				"-Q", timeStr, "-Q", timeStr, "-Q", timeStr,
				"-R", timeStr, "-R", timeStr, "-R", timeStr,
				"-S", timeStr, "-S", timeStr, "-S", timeStr,
				"-T", timeStr, "-T", timeStr, "-T", timeStr,
				"-U", szStr, "-U", szStr, "-U", szStr,

				"-VV", "-V=3",
				"-WW", "-W", "3",
				"-XX", "-XX=2",
				"-YY", "-Y", "2",
				"-ZZ", "-Z", "3",
				"-vv", "-v", "3",
				"-ww", "-w", "3",
				"-xx", "-x", "3",
				"-yy", "-y", "3",
				"-zz", "-z", "3",
				"-✋✋=1", "-✋", "3",
				"-🖖🖖", "-🖖", "3",
				"-🍵🍵", "-🍵=3",
				"-🤔🤔=1s", "-🤔", "3s",
				"-🥵🥵", "-🥵", "1", "-🥵🥵",

				"--", "-f", "-foo",
			},
			posArgs:  []string{"-", "-"},
			dashArgs: []string{"-f", "-foo"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Run("Custom Flag Finder", func(t *testing.T) {
				var (
					actual FlagTestOptions
					flags  TestFlags
				)

				flags.Bind(&actual)
				posArgs, dashArgs, err := ParseFlags(test.args, &flags, &opts)
				assert.NoError(t, err)

				assert.EqS(t, test.posArgs, posArgs)
				assert.EqS(t, test.dashArgs, dashArgs)
				assertOptsEq(t, expected, actual)
			})

			t.Run("ReflectIndexer", func(t *testing.T) {
				var actual FlagTestOptions

				flags := NewReflectIndexer(DefaultReflectVPFactory{}, &actual)
				posArgs, dashArgs, err := ParseFlags(test.args, flags, &opts)
				assert.NoError(t, err)

				assert.EqS(t, test.posArgs, posArgs)
				assert.EqS(t, test.dashArgs, dashArgs)
				assertOptsEq(t, expected, actual)
			})
		})
	}
}

type FlagTestOptions struct {
	String    string        `cli:"a|String"`
	Bool      bool          `cli:"b|Bool"`
	Int       int           `cli:"c|Int"`
	Int8      int8          `cli:"d|Int8"`
	Int16     int16         `cli:"e|Int16"`
	Int32     int32         `cli:"f|Int32"`
	Int64     int64         `cli:"g|Int64"`
	Uint      uint          `cli:"h|Uint"`
	Uint8     uint8         `cli:"i|Uint8"`
	Uint16    uint16        `cli:"j|Uint16"`
	Uint32    uint32        `cli:"k|Uint32"`
	Uint64    uint64        `cli:"l|Uint64"`
	Uintptr   uintptr       `cli:"m|Uintptr"`
	Float32   float32       `cli:"n|Float32"`
	Float64   float64       `cli:"o|Float64"`
	Duration  time.Duration `cli:"p|Duration,value=dur"`
	Time      time.Time     `cli:"q|Time,value=time"`
	Size      int64         `cli:"u|Size,value=size"`
	UnixSec   int64         `cli:"r|UnixSec,value=unix-ts"`
	UnixMilli int64         `cli:"s|UnixMilli,value=unix-ms"`
	UnixNano  int64         `cli:"t|UnixNano,value=unix-ns"`

	StringSlice    []string        `cli:"A|StringSlice"`
	BoolSlice      []bool          `cli:"B|BoolSlice"`
	IntSlice       []int           `cli:"C|IntSlice"`
	Int8Slice      []int8          `cli:"D|Int8Slice"`
	Int16Slice     []int16         `cli:"E|Int16Slice"`
	Int32Slice     []int32         `cli:"F|Int32Slice"`
	Int64Slice     []int64         `cli:"G|Int64Slice"`
	UintSlice      []uint          `cli:"H|UintSlice"`
	Uint8Slice     []uint8         `cli:"I|Uint8Slice"`
	Uint16Slice    []uint16        `cli:"J|Uint16Slice"`
	Uint32Slice    []uint32        `cli:"K|Uint32Slice"`
	Uint64Slice    []uint64        `cli:"L|Uint64Slice"`
	UintptrSlice   []uintptr       `cli:"M|UintptrSlice"`
	Float32Slice   []float32       `cli:"N|Float32Slice"`
	Float64Slice   []float64       `cli:"O|Float64Slice"`
	DurationSlice  []time.Duration `cli:"P|DurationSlice,value=dur"`
	TimeSlice      []time.Time     `cli:"Q|TimeSlice,value=time"`
	SizeSlice      []int64         `cli:"U|SizeSlice,value=size"`
	UnixSecSlice   []int64         `cli:"R|UnixSecSlice,value=unix-ts"`
	UnixMilliSlice []int64         `cli:"S|UnixMilliSlice,value=unix-ms"`
	UnixNanoSlice  []int64         `cli:"T|UnixNanoSlice,value=unix-ns"`

	IntSum      int           `cli:"V|IntSum,value=sum"`
	Int8Sum     int8          `cli:"W|Int8Sum,value=sum"`
	Int16Sum    int16         `cli:"X|Int16Sum,value=sum"`
	Int32Sum    int32         `cli:"Y|Int32Sum,value=sum"`
	Int64Sum    int64         `cli:"Z|Int64Sum,value=sum"`
	UintSum     uint          `cli:"v|UintSum,value=sum"`
	Uint8Sum    uint8         `cli:"w|Uint8Sum,value=sum"`
	Uint16Sum   uint16        `cli:"x|Uint16Sum,value=sum"`
	Uint32Sum   uint32        `cli:"y|Uint32Sum,value=sum"`
	Uint64Sum   uint64        `cli:"z|Uint64Sum,value=sum"`
	UintptrSum  uintptr       `cli:"✋|UintptrSum,value=sum"`
	Float32Sum  float32       `cli:"🖖|Float32Sum,value=sum"`
	Float64Sum  float64       `cli:"🍵|Float64Sum,value=sum"`
	DurationSum time.Duration `cli:"🤔|DurationSum,value=dsum"`
	SizeSum     int64         `cli:"🥵|SizeSum,value=ssum"`

	Empty             bool            `cli:"Empty"`
	Regexp            regexp.Regexp   `cli:"Regexp,value=regexp"`
	RegexpNocase      regexp.Regexp   `cli:"RegexpNocase,value=regexp-nocase"`
	RegexpSlice       []regexp.Regexp `cli:"RegexpSlice,value=regexp"`
	RegexpNocaseSlice []regexp.Regexp `cli:"RegexpNocaseSlice,value=regexp-nocase"`
}

func assertOptsEq(t *testing.T, expected, actual FlagTestOptions) bool {
	return assert.Eq(t, expected.String, actual.String) &&
		assert.Eq(t, expected.Bool, actual.Bool) &&
		assert.Eq(t, expected.Int, actual.Int) &&
		assert.Eq(t, expected.Int8, actual.Int8) &&
		assert.Eq(t, expected.Int16, actual.Int16) &&
		assert.Eq(t, expected.Int32, actual.Int32) &&
		assert.Eq(t, expected.Int64, actual.Int64) &&
		assert.Eq(t, expected.Uint, actual.Uint) &&
		assert.Eq(t, expected.Uint8, actual.Uint8) &&
		assert.Eq(t, expected.Uint16, actual.Uint16) &&
		assert.Eq(t, expected.Uint32, actual.Uint32) &&
		assert.Eq(t, expected.Uint64, actual.Uint64) &&
		assert.Eq(t, expected.Uintptr, actual.Uintptr) &&
		assert.Eq(t, expected.Float32, actual.Float32) &&
		assert.Eq(t, expected.Float64, actual.Float64) &&
		assert.Eq(t, expected.Duration, actual.Duration) &&
		assert.Eq(t, expected.Time, actual.Time) &&
		assert.Eq(t, expected.Size, actual.Size) &&
		assert.Eq(t, expected.UnixSec, actual.UnixSec) &&
		assert.Eq(t, expected.UnixMilli, actual.UnixMilli) &&
		assert.Eq(t, expected.UnixNano, actual.UnixNano) &&
		assert.EqS(t, expected.StringSlice, actual.StringSlice) &&
		assert.EqS(t, expected.BoolSlice, actual.BoolSlice) &&
		assert.EqS(t, expected.IntSlice, actual.IntSlice) &&
		assert.EqS(t, expected.Int8Slice, actual.Int8Slice) &&
		assert.EqS(t, expected.Int16Slice, actual.Int16Slice) &&
		assert.EqS(t, expected.Int32Slice, actual.Int32Slice) &&
		assert.EqS(t, expected.Int64Slice, actual.Int64Slice) &&
		assert.EqS(t, expected.UintSlice, actual.UintSlice) &&
		assert.EqS(t, expected.Uint8Slice, actual.Uint8Slice) &&
		assert.EqS(t, expected.Uint16Slice, actual.Uint16Slice) &&
		assert.EqS(t, expected.Uint32Slice, actual.Uint32Slice) &&
		assert.EqS(t, expected.Uint64Slice, actual.Uint64Slice) &&
		assert.EqS(t, expected.UintptrSlice, actual.UintptrSlice) &&
		assert.EqS(t, expected.Float32Slice, actual.Float32Slice) &&
		assert.EqS(t, expected.Float64Slice, actual.Float64Slice) &&
		assert.EqS(t, expected.DurationSlice, actual.DurationSlice) &&
		assert.EqS(t, expected.TimeSlice, actual.TimeSlice) &&
		assert.EqS(t, expected.SizeSlice, actual.SizeSlice) &&
		assert.EqS(t, expected.UnixSecSlice, actual.UnixSecSlice) &&
		assert.EqS(t, expected.UnixMilliSlice, actual.UnixMilliSlice) &&
		assert.EqS(t, expected.UnixNanoSlice, actual.UnixNanoSlice) &&
		assert.Eq(t, expected.IntSum, actual.IntSum) &&
		assert.Eq(t, expected.Int8Sum, actual.Int8Sum) &&
		assert.Eq(t, expected.Int16Sum, actual.Int16Sum) &&
		assert.Eq(t, expected.Int32Sum, actual.Int32Sum) &&
		assert.Eq(t, expected.Int64Sum, actual.Int64Sum) &&
		assert.Eq(t, expected.UintSum, actual.UintSum) &&
		assert.Eq(t, expected.Uint8Sum, actual.Uint8Sum) &&
		assert.Eq(t, expected.Uint16Sum, actual.Uint16Sum) &&
		assert.Eq(t, expected.Uint32Sum, actual.Uint32Sum) &&
		assert.Eq(t, expected.Uint64Sum, actual.Uint64Sum) &&
		assert.Eq(t, expected.UintptrSum, actual.UintptrSum) &&
		assert.Eq(t, expected.Float32Sum, actual.Float32Sum) &&
		assert.Eq(t, expected.Float64Sum, actual.Float64Sum) &&
		assert.Eq(t, expected.DurationSum, actual.DurationSum) &&
		assert.Eq(t, expected.SizeSum, actual.SizeSum)
}

var _ FlagFinder = (*TestFlags)(nil)

type TestFlags struct {
	String String

	Bool Bool

	Int   Int
	Int8  Int8
	Int16 Int16
	Int32 Int32
	Int64 Int64

	Uint    Uint
	Uint8   Uint8
	Uint16  Uint16
	Uint32  Uint32
	Uint64  Uint64
	Uintptr Uintptr

	Float32 Float32
	Float64 Float64

	Duration  Duration
	Time      Time
	UnixSec   UnixSec
	UnixMilli UnixMilli
	UnixNano  UnixNano

	Size Size

	StringSlice StringSlice

	BoolSlice BoolSlice

	IntSlice   IntSlice
	Int8Slice  Int8Slice
	Int16Slice Int16Slice
	Int32Slice Int32Slice
	Int64Slice Int64Slice

	UintSlice    UintSlice
	Uint8Slice   Uint8Slice
	Uint16Slice  Uint16Slice
	Uint32Slice  Uint32Slice
	Uint64Slice  Uint64Slice
	UintptrSlice UintptrSlice

	Float32Slice Float32Slice
	Float64Slice Float64Slice

	DurationSlice  DurationSlice
	TimeSlice      TimeSlice
	UnixSecSlice   UnixSecSlice
	UnixMilliSlice UnixMilliSlice
	UnixNanoSlice  UnixNanoSlice

	SizeSlice SizeSlice

	IntSum   IntSum
	Int8Sum  Int8Sum
	Int16Sum Int16Sum
	Int32Sum Int32Sum
	Int64Sum Int64Sum

	UintSum    UintSum
	Uint8Sum   Uint8Sum
	Uint16Sum  Uint16Sum
	Uint32Sum  Uint32Sum
	Uint64Sum  Uint64Sum
	UintptrSum UintptrSum

	Float32Sum Float32Sum
	Float64Sum Float64Sum

	DurationSum DurationSum

	SizeSum SizeSum

	// below are unchecked values

	Empty             FlagEmptyV
	Regexp            RegexpV
	RegexpNocase      RegexpNocaseV
	RegexpSlice       RegexpSliceV
	RegexpNocaseSlice RegexpNocaseSliceV
}

// FindFlag implements [FlagFinder]
func (tf *TestFlags) FindFlag(name string) (Flag, bool) {
	switch name {
	case "String", "a":
		return &tf.String, true
	case "Bool", "b":
		return &tf.Bool, true
	case "Int", "c":
		return &tf.Int, true
	case "Int8", "d":
		return &tf.Int8, true
	case "Int16", "e":
		return &tf.Int16, true
	case "Int32", "f":
		return &tf.Int32, true
	case "Int64", "g":
		return &tf.Int64, true

	case "Uint", "h":
		return &tf.Uint, true
	case "Uint8", "i":
		return &tf.Uint8, true
	case "Uint16", "j":
		return &tf.Uint16, true
	case "Uint32", "k":
		return &tf.Uint32, true
	case "Uint64", "l":
		return &tf.Uint64, true
	case "Uintptr", "m":
		return &tf.Uintptr, true
	case "Float32", "n":
		return &tf.Float32, true
	case "Float64", "o":
		return &tf.Float64, true
	case "Duration", "p":
		return &tf.Duration, true
	case "Time", "q":
		return &tf.Time, true
	case "UnixSec", "r":
		return &tf.UnixSec, true
	case "UnixMilli", "s":
		return &tf.UnixMilli, true
	case "UnixNano", "t":
		return &tf.UnixNano, true
	case "Size", "u":
		return &tf.Size, true

	case "StringSlice", "A":
		return &tf.StringSlice, true
	case "BoolSlice", "B":
		return &tf.BoolSlice, true
	case "IntSlice", "C":
		return &tf.IntSlice, true
	case "Int8Slice", "D":
		return &tf.Int8Slice, true
	case "Int16Slice", "E":
		return &tf.Int16Slice, true
	case "Int32Slice", "F":
		return &tf.Int32Slice, true
	case "Int64Slice", "G":
		return &tf.Int64Slice, true

	case "UintSlice", "H":
		return &tf.UintSlice, true
	case "Uint8Slice", "I":
		return &tf.Uint8Slice, true
	case "Uint16Slice", "J":
		return &tf.Uint16Slice, true
	case "Uint32Slice", "K":
		return &tf.Uint32Slice, true
	case "Uint64Slice", "L":
		return &tf.Uint64Slice, true
	case "UintptrSlice", "M":
		return &tf.UintptrSlice, true
	case "Float32Slice", "N":
		return &tf.Float32Slice, true
	case "Float64Slice", "O":
		return &tf.Float64Slice, true
	case "DurationSlice", "P":
		return &tf.DurationSlice, true
	case "TimeSlice", "Q":
		return &tf.TimeSlice, true
	case "UnixSecSlice", "R":
		return &tf.UnixSecSlice, true
	case "UnixMilliSlice", "S":
		return &tf.UnixMilliSlice, true
	case "UnixNanoSlice", "T":
		return &tf.UnixNanoSlice, true
	case "SizeSlice", "U":
		return &tf.SizeSlice, true

	case "IntSum", "V":
		return &tf.IntSum, true
	case "Int8Sum", "W":
		return &tf.Int8Sum, true
	case "Int16Sum", "X":
		return &tf.Int16Sum, true
	case "Int32Sum", "Y":
		return &tf.Int32Sum, true
	case "Int64Sum", "Z":
		return &tf.Int64Sum, true
	case "UintSum", "v":
		return &tf.UintSum, true
	case "Uint8Sum", "w":
		return &tf.Uint8Sum, true
	case "Uint16Sum", "x":
		return &tf.Uint16Sum, true
	case "Uint32Sum", "y":
		return &tf.Uint32Sum, true
	case "Uint64Sum", "z":
		return &tf.Uint64Sum, true
	case "UintptrSum", "✋":
		return &tf.UintptrSum, true
	case "Float32Sum", "🖖":
		return &tf.Float32Sum, true
	case "Float64Sum", "🍵":
		return &tf.Float64Sum, true
	case "DurationSum", "🤔":
		return &tf.DurationSum, true
	case "SizeSum", "🥵":
		return &tf.SizeSum, true

	case "Empty":
		return &tf.Empty, true
	case "Regexp":
		return &tf.Regexp, true
	case "RegexpNocase":
		return &tf.RegexpNocase, true
	case "RegexpSlice":
		return &tf.RegexpSlice, true
	case "RegexpNocaseSlice":
		return &tf.RegexpNocaseSlice, true

	}

	return nil, false
}

func (tf *TestFlags) Bind(opts *FlagTestOptions) {
	tf.String.Value = &opts.String
	tf.Bool.Value = &opts.Bool
	tf.Int.Value = &opts.Int
	tf.Int8.Value = &opts.Int8
	tf.Int16.Value = &opts.Int16
	tf.Int32.Value = &opts.Int32
	tf.Int64.Value = &opts.Int64
	tf.Uint.Value = &opts.Uint
	tf.Uint8.Value = &opts.Uint8
	tf.Uint16.Value = &opts.Uint16
	tf.Uint32.Value = &opts.Uint32
	tf.Uint64.Value = &opts.Uint64
	tf.Uintptr.Value = &opts.Uintptr
	tf.Float32.Value = &opts.Float32
	tf.Float64.Value = &opts.Float64
	tf.Duration.Value = &opts.Duration
	tf.Time.Value = &opts.Time
	tf.UnixSec.Value = &opts.UnixSec
	tf.UnixMilli.Value = &opts.UnixMilli
	tf.UnixNano.Value = &opts.UnixNano
	tf.Size.Value = &opts.Size

	tf.StringSlice.Value = &opts.StringSlice
	tf.BoolSlice.Value = &opts.BoolSlice
	tf.IntSlice.Value = &opts.IntSlice
	tf.Int8Slice.Value = &opts.Int8Slice
	tf.Int16Slice.Value = &opts.Int16Slice
	tf.Int32Slice.Value = &opts.Int32Slice
	tf.Int64Slice.Value = &opts.Int64Slice
	tf.UintSlice.Value = &opts.UintSlice
	tf.Uint8Slice.Value = &opts.Uint8Slice
	tf.Uint16Slice.Value = &opts.Uint16Slice
	tf.Uint32Slice.Value = &opts.Uint32Slice
	tf.Uint64Slice.Value = &opts.Uint64Slice
	tf.UintptrSlice.Value = &opts.UintptrSlice
	tf.Float32Slice.Value = &opts.Float32Slice
	tf.Float64Slice.Value = &opts.Float64Slice
	tf.DurationSlice.Value = &opts.DurationSlice
	tf.TimeSlice.Value = &opts.TimeSlice
	tf.UnixSecSlice.Value = &opts.UnixSecSlice
	tf.UnixMilliSlice.Value = &opts.UnixMilliSlice
	tf.UnixNanoSlice.Value = &opts.UnixNanoSlice
	tf.SizeSlice.Value = &opts.SizeSlice

	tf.IntSum.Value = &opts.IntSum
	tf.Int8Sum.Value = &opts.Int8Sum
	tf.Int16Sum.Value = &opts.Int16Sum
	tf.Int32Sum.Value = &opts.Int32Sum
	tf.Int64Sum.Value = &opts.Int64Sum
	tf.UintSum.Value = &opts.UintSum
	tf.Uint8Sum.Value = &opts.Uint8Sum
	tf.Uint16Sum.Value = &opts.Uint16Sum
	tf.Uint32Sum.Value = &opts.Uint32Sum
	tf.Uint64Sum.Value = &opts.Uint64Sum
	tf.UintptrSum.Value = &opts.UintptrSum
	tf.Float32Sum.Value = &opts.Float32Sum
	tf.Float64Sum.Value = &opts.Float64Sum
	tf.DurationSum.Value = &opts.DurationSum
	tf.SizeSum.Value = &opts.SizeSum
}
