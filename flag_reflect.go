// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// FlagReflect
//
// It is used internally by the ReflectIndexer.
type FlagReflect struct {
	VP           VP[*reflect.Value]
	Value        reflect.Value
	BriefUsage   string
	DefaultValue string
	Comp         []string
	State_       FlagState
}

func (f *FlagReflect) Type() (string, bool) {
	t := f.VP.Type().String()
	return t, len(t) != 0
}

func (f *FlagReflect) ImplyValue() (string, bool) {
	return implyFromVPType(f.VP.Type())
}

func (f *FlagReflect) Extra() any       { return nil }
func (f *FlagReflect) HasValue() bool   { return f.VP.HasValue(&f.Value) }
func (f *FlagReflect) State() FlagState { return f.State_ }
func (f *FlagReflect) Default() string  { return f.DefaultValue }
func (f *FlagReflect) Usage() string    { return f.BriefUsage }

func (f *FlagReflect) PrintValue(out io.Writer) (int, error) {
	return f.VP.PrintValue(out, &f.Value)
}

func (f *FlagReflect) Decode(opts *ParseOptions, name, arg string, set bool) error {
	err := f.VP.ParseValue(opts, arg, &f.Value, set)
	if err != nil {
		return err
	}

	if set {
		f.State_ |= FlagStateValueChanged
	}

	return nil
}

// Suggest implements [CompAction].
func (f *FlagReflect) Suggest(tsk *CompTask) (added int, _ CompState) {
	for _, v := range f.Comp {
		added += tsk.AddMatched(false, CompItem{
			Value: v,
			Kind:  CompKindFlagValue,
		})
	}

	return
}

// ReflectVPFactory handles creation of VP[*reflect.Value] for struct fields.
type ReflectVPFactory interface {
	GetVPReflectFor(fieldType reflect.Type, keyType, valueType string) (VP[*reflect.Value], error)
}

type ErrUnsupportedType struct {
	Type      reflect.Type
	KeyType   string
	ValueType string
}

func (e *ErrUnsupportedType) Error() string {
	switch {
	case len(e.KeyType) == 0 && len(e.ValueType) == 0:
		return "unsupported type: " + e.Type.String()
	case len(e.KeyType) != 0 && len(e.ValueType) == 0:
		return "unsupported type (key=" + e.KeyType + "): " + e.Type.String()
	case len(e.KeyType) == 0 && len(e.ValueType) != 0:
		return "unsupported type (value=" + e.ValueType + "): " + e.Type.String()
	case len(e.KeyType) != 0 && len(e.ValueType) != 0:
		return "unsupported type (key=" + e.KeyType + ", value=" + e.ValueType + "): " + e.Type.String()
	}

	return ""
}

// DefaultReflectVPFactory is the ReflectVPFactory implementation referenced
// from comments of ReflectIndexer.
type DefaultReflectVPFactory struct{}

func (DefaultReflectVPFactory) GetVPReflectFor(fieldType reflect.Type, keyType, valueType string) (vp VP[*reflect.Value], err error) {
	ft := noptr(fieldType)
	switch ft.Kind() {
	case reflect.Map:
		kp := getScalarOrSliceVP(keyType, ft.Key(), false)
		if kp == nil {
			break
		}

		rawVt := ft.Elem()
		vt := noptr(rawVt)
		if vt.Kind() == reflect.Slice {
			if svp := getScalarOrSliceVP(valueType, vt.Elem(), true); svp != nil {
				vp = &VPReflectMap[VP[*reflect.Value], VP[*reflect.Value]]{
					Key:  kp,
					Elem: svp,
				}
			}
		} else {
			if svp := getScalarOrSliceVP(valueType, rawVt, false); svp != nil {
				vp = &VPReflectMap[VP[*reflect.Value], VP[*reflect.Value]]{
					Key:  kp,
					Elem: svp,
				}
			}
		}

	case reflect.Slice:
		vp = getScalarOrSliceVP(valueType, ft.Elem(), true)
	default:
		vp = getScalarOrSliceVP(valueType, fieldType, false)
	}

	if vp == nil {
		return nil, &ErrUnsupportedType{
			Type:      fieldType,
			KeyType:   keyType,
			ValueType: valueType,
		}
	}

	return vp, nil
}

func getScalarOrSliceVP(req string, rawFt reflect.Type, slice bool) VP[*reflect.Value] {
	ft := noptr(rawFt)
	sum := strings.HasSuffix(req, "sum")

	switch req {
	case "dur", "dsum":
		switch ft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		default:
			return nil
		}
		if sum {
			return VPReflectSum[VPReflectDuration]{}
		}
		if slice {
			return VPReflectSlice[VPReflectDuration]{}
		}
		return VPReflectDuration{}
	case "size", "ssum":
		switch ft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		default:
			return nil
		}
		if sum {
			return VPReflectSum[VPReflectSize]{}
		}
		if slice {
			return VPReflectSlice[VPReflectSize]{}
		}
		return VPReflectSize{}
	case "unix-ts":
		if sum || ft.Kind() != reflect.Int64 {
			return nil
		}
		if slice {
			return VPReflectSlice[VPReflectUnixSec]{}
		}
		return VPReflectUnixSec{}
	case "unix-ms":
		if sum || ft.Kind() != reflect.Int64 {
			return nil
		}
		if slice {
			return VPReflectSlice[VPReflectUnixMilli]{}
		}
		return VPReflectUnixMilli{}
	case "unix-us":
		if sum || ft.Kind() != reflect.Int64 {
			return nil
		}
		if slice {
			return VPReflectSlice[VPReflectUnixMicro]{}
		}
		return VPReflectUnixMicro{}
	case "unix-ns":
		if sum || ft.Kind() != reflect.Int64 {
			return nil
		}
		if slice {
			return VPReflectSlice[VPReflectUnixNano]{}
		}
		return VPReflectUnixNano{}
	case "time":
		if sum || !ft.ConvertibleTo(reflect.TypeOf((*time.Time)(nil)).Elem()) {
			return nil
		}
		if slice {
			return VPReflectSlice[VPReflectTime]{}
		}
		return VPReflectTime{}
	case "regexp":
		if sum || !ft.ConvertibleTo(reflect.TypeOf((*regexp.Regexp)(nil)).Elem()) {
			return nil
		}
		if slice {
			return VPReflectSlice[VPReflectRegexp]{}
		}
		return VPReflectRegexp{}
	case "regexp-nocase":
		if sum || !ft.ConvertibleTo(reflect.TypeOf((*regexp.Regexp)(nil)).Elem()) {
			return nil
		}
		if slice {
			return VPReflectSlice[VPReflectRegexpNocase]{}
		}
		return VPReflectRegexpNocase{}
	case "", "sum":
	default:
		return nil
	}

	switch ft.Kind() {
	case reflect.Bool:
		switch {
		case slice:
			return VPReflectSlice[VPReflectBool]{}
		case sum:
			return nil
		default:
			return VPReflectBool{}
		}
	case reflect.String:
		switch {
		case slice:
			return VPReflectSlice[VPReflectString]{}
		case sum:
			return nil
		default:
			return VPReflectString{}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch {
		case slice:
			return VPReflectSlice[VPReflectInt]{}
		case sum:
			return VPReflectSum[VPReflectInt]{}
		default:
			return VPReflectInt{}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch {
		case slice:
			return VPReflectSlice[VPReflectUint]{}
		case sum:
			return VPReflectSum[VPReflectUint]{}
		default:
			return VPReflectUint{}
		}
	case reflect.Float32, reflect.Float64:
		switch {
		case slice:
			return VPReflectSlice[VPReflectFloat]{}
		case sum:
			return VPReflectSum[VPReflectFloat]{}
		default:
			return VPReflectFloat{}
		}
	}

	return nil
}

// noptr returns the first non-pointer type from typ.
func noptr(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	return typ
}

// NewReflectIndexer creates a new ReflectIndexer for the pointer to struct.
func NewReflectIndexer(factory ReflectVPFactory, pStruct any) *ReflectIndexer {
	val := reflect.Indirect(reflect.ValueOf(pStruct))
	if !val.IsValid() || !val.CanAddr() {
		panic("invalid flag collection value: not addressable")
	}

	for val.Kind() == reflect.Pointer {
		if !val.IsValid() || val.IsNil() {
			panic("invalid flag collection value: not initialized")
		}

		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		panic("invalid flag collection value: not a struct")
	}

	return &ReflectIndexer{
		StructV: val,
		Factory: factory,
		Names:   nil,
		Refs:    nil,

		TotalFlags: 0,
	}
}

type ReflectFlagRef struct {
	Field   int
	Options string
	Flag    *FlagReflect
	Info    FlagInfo
}

// A ReflectIndexer creates and caches flags on request using reflection.
//
// Under most circumstances, an application runs with no more than 15 args
// but can have way more flags available, when dynamic allocation and
// reflection (binary size) is not a concern, a ReflectIndexer can speedup
// both your development and application startup.
//
// Struct field tag specification
//
//	`cli:"<long name>|<shorthand>[,comp=<completion>][,value=<type>][,key=<type>][,def=<default>][,hide][,once][,#<brief usage>]"`
//
// Text before the first comma (',') is interpreted as flag name section, it
// SHOULD contain at most two names (one long name and one shorthand),
// use pipe ('|') to separate names.
//
// Text after the first comma and before the sharp ('#') is interpreted as
// flag options, currently there are six options available:
//
//   - comp=<completion>
//   - value=<type>
//   - key=<type>
//   - def=<value>
//   - hide
//   - once
//
// Option `comp` defines completion values, multiple `comp` option creates
// multiple CompItems, for example:
//
//	type Example struct{
//	    Foo string `cli:"foo,comp=value1,comp=value2"`
//	}
//
// The ReflectIndexer will produce two CompItems for field Foo's command-line
// flag `--foo`:
//   - value1
//   - value2
//
// Option `value` is used to change the method used to decode text arg and
// can have one of following `<type>` values when using
// DefaultReflectVPFactory:
//
//   - size     (size value, example command-line arg: "1TB", "1g1M")
//   - dur      (duration value, example command-line arg: "1yr", "1m10s")
//   - sum      (sums numeric values)
//   - ssum     (sums size values)
//   - dsum     (sums duration values)
//   - regexp
//   - regexp-nocase
//   - time    (decode time string, example command-line arg: "15:00", "21")
//   - unix-ts (decode time string to seconds since the unix epoch)
//   - unix-ms (decode time string to milliseconds since the unix epoch)
//   - unix-us (decode time string to microseconds since the unix epoch)
//   - unix-ns (decode time string to nanoseconds since the unix epoch)
//
// Option `value`'s meaning varies depending on the field type:
//
//   - scalar field: for that scalar field (e.g. `value=dur` for int64)
//   - slice field: for slice element type (e.g. `value=unix-ts` for uint64 in []uint64)
//   - map field: for map value type (e.g. `value=regexp` for *regexp.Regexp in map[K]*regexp.Regexp)
//
// Option `key` is meant for map key, and works the same way as `value`.
// It can have one of following values when using DefaultReflectVPFactory:
//
//   - size
//   - dur
//   - time
//   - unix-ts
//   - unix-ms
//   - unix-us
//   - unix-ns
//
// Both option `value` and option `key` can present at most once in the tag value.
//
// Option `def` defines a default value for the flag when flag is not set.
// There can be multiple `def` options.
//
// Option `hide` marks the FlagState with FlagStateHidden. There can be no
// more than one `hide` option.
//
// Option `once` marks the FlagState with FlagStateSetAtMostOnce. There can be
// no more than one `once` option.
//
// The remaining text after the sharp sign ('#') after the first comma, is
// interpreted as the brief usage of the flag.
//
// NOTE: Unexported fields and fields without a `cli` tag value are ignored.
type ReflectIndexer struct {
	// StructV is the reflect value of the addressable struct.
	StructV reflect.Value

	// Factory handles creation of VP[*reflect.Value]
	Factory ReflectVPFactory

	// Names maps name to index into Refs
	Names map[string]int

	// Refs are cached flags in struct field order
	Refs []ReflectFlagRef

	// TotalFlags is the count of flags in StructV
	//
	// When = 0: count unknown
	// When < 0: no flag
	// When > 0: n flags, len(Refs) = TotalFlags
	TotalFlags int
}

func (r *ReflectIndexer) FindFlag(s string) (Flag, bool) {
	if r == nil || r.TotalFlags < 0 {
		return nil, false
	}

	if r.Names != nil {
		i, ok := r.Names[s]
		if ok {
			return r.getFieldFlag(i), true
		}
	}

	if r.TotalFlags != 0 { // has checked all fields
		return nil, false
	}

	typ := r.StructV.Type()
	totalFlags := 0
	for fieldIndex, flagIndex, n := 0, -1, typ.NumField(); fieldIndex < n; fieldIndex++ {
		f := typ.Field(fieldIndex)
		if !f.IsExported() {
			continue
		}

		pos := indexTag(string(f.Tag), "cli")
		if pos < 0 {
			continue
		}

		flagIndex++
		totalFlags++
		if flagIndex < len(r.Refs) { // has been checked
			continue
		}

		ref, found := r.createRefFromTag(fieldIndex, f.Tag[pos:].Get("cli"), flagIndex, s)
		r.Refs = append(r.Refs, ref)

		if found {
			if fieldIndex == n-1 {
				r.TotalFlags = totalFlags
			}

			return r.getFieldFlag(flagIndex), true
		}
	}

	if totalFlags == 0 {
		r.TotalFlags = -1
	} else {
		r.TotalFlags = totalFlags
	}

	return nil, false
}

func (r *ReflectIndexer) NthFlag(i int) (FlagInfo, bool) {
	if r == nil || r.TotalFlags < 0 {
		return FlagInfo{}, false
	}

	if i < len(r.Refs) {
		return r.Refs[i].Info, true
	}

	if r.TotalFlags > 0 {
		return FlagInfo{}, false
	}

	typ := r.StructV.Type()
	totalFlags := 0
	for fieldIndex, flagIndex, n := 0, -1, typ.NumField(); fieldIndex < n; fieldIndex++ {
		f := typ.Field(fieldIndex)
		if !f.IsExported() {
			continue
		}

		pos := indexTag(string(f.Tag), "cli")
		if pos < 0 {
			continue
		}

		flagIndex++
		totalFlags++
		if flagIndex < len(r.Refs) { // has been checked
			if i == 0 {
				return r.Refs[flagIndex].Info, true
			}
			i--
			continue
		}

		ref, _ := r.createRefFromTag(fieldIndex, f.Tag[pos:].Get("cli"), flagIndex, "")
		r.Refs = append(r.Refs, ref)

		if i == 0 {
			if fieldIndex == n-1 {
				r.TotalFlags = totalFlags
			}

			return r.Refs[flagIndex].Info, true
		}

		i--
	}

	if totalFlags == 0 {
		r.TotalFlags = -1
	} else {
		r.TotalFlags = totalFlags
	}

	return FlagInfo{}, false
}

func (r *ReflectIndexer) createRefFromTag(
	fieldIndex int, tag string, flagIndex int, matchName string,
) (ref ReflectFlagRef, nameMatch bool) {
	opt, tag, _ := strings.Cut(tag, ",")
	for len(opt) > 0 {
		var name string
		name, opt, _ = strings.Cut(opt, "|")
		if len(name) == 0 {
			continue
		}

		if r.Names == nil {
			r.Names = map[string]int{}
		}
		r.Names[name] = flagIndex

		if IsShorthand(name) {
			if len(ref.Info.Shorthand) == 0 {
				ref.Info.Shorthand = name
			}
		} else {
			if len(ref.Info.Name) == 0 {
				ref.Info.Name = name
			}
		}

		if !nameMatch && len(matchName) != 0 && matchName == name {
			nameMatch = true
		}
	}

	ref.Field = fieldIndex
	ref.Options = tag
	var (
		def  string
		defs strings.Builder
	)

	tag, _, _ = strings.Cut(tag, "#")
	for len(tag) != 0 {
		var opt string
		opt, tag, _ = strings.Cut(tag, ",")

		key, value, _ := strings.Cut(opt, "=")
		switch key {
		case "comp", "value", "key": // used when creating flag
		case "def":
			if defs.Len() != 0 {
				defs.WriteString(", ")
				defs.WriteString(value)
			} else if len(def) != 0 {
				defs.WriteByte('[')
				defs.WriteString(def)
				defs.WriteString(", ")
				defs.WriteString(value)
			} else {
				def = value
			}
		case "hide":
			if ref.Info.State&FlagStateHidden != 0 {
				panic("invalid duplicate `hide` option")
			}

			ref.Info.State |= FlagStateHidden
		case "once":
			if ref.Info.State&FlagStateSetAtMostOnce != 0 {
				panic("invalid duplicate `once` option")
			}

			ref.Info.State |= FlagStateSetAtMostOnce
		}
	}

	if defs.Len() != 0 {
		defs.WriteByte(']')
		ref.Info.DefaultValue = defs.String()
	} else {
		ref.Info.DefaultValue = def
	}

	return
}

// indexTag returns the offset of the key in tag, if there is no such key
// in tag, return -1.
func indexTag(tag, key string) int {
	var offset, i int
	for len(tag) != 0 {
		// skip leading spaces
		for i = 0; i < len(tag) && tag[i] == ' '; i++ {
		}
		offset += i

		i = strings.Index(tag, ":\"")
		if i == -1 {
			return -1
		}

		if i == len(key) && tag[:i] == key {
			return offset
		}

		offset += 2
		tag = tag[i+2:]

		// we don't care about the value, just find the first unescaped double quote
		for i = 0; i < len(tag); i++ {
			if tag[i] == '\\' {
				i++
			} else if tag[i] == '"' {
				i++
				break
			}
		}

		offset += i
		tag = tag[i:]
	}

	return -1
}

func (r *ReflectIndexer) getFieldFlag(ref int) Flag {
	if r.Refs[ref].Flag != nil {
		return r.Refs[ref].Flag
	}

	var (
		comp []string

		keyType, valueType string
	)

	options, usage, _ := strings.Cut(r.Refs[ref].Options, "#")
	for len(options) != 0 {
		var opt string
		opt, options, _ = strings.Cut(options, ",")

		key, value, _ := strings.Cut(opt, "=")
		switch key {
		case "comp":
			if len(value) != 0 {
				comp = append(comp, value)
			}
		case "value":
			if len(valueType) != 0 {
				panic("invalid multiple value types: " + opt)
			}
			valueType = value
		case "key":
			if len(keyType) != 0 {
				panic("invalid multiple key types: " + opt)
			}
			keyType = value
		case "def", "hide", "once": // reuse value in FlagInfo
		default:
			// TODO: panic on unknown option?
		}
	}

	fieldIdx := r.Refs[ref].Field
	fieldType := r.StructV.Type().Field(fieldIdx).Type
	vp, err := r.Factory.GetVPReflectFor(fieldType, keyType, valueType)
	if err != nil {
		panic(err)
	}

	if vp == nil { // defensive check
		panic("unsupported field type: " + fieldType.String())
	}

	r.Refs[ref].Flag = &FlagReflect{
		VP:           vp,
		Value:        r.StructV.Field(fieldIdx),
		BriefUsage:   usage,
		DefaultValue: r.Refs[ref].Info.DefaultValue,
		Comp:         comp,
		State_:       r.Refs[ref].Info.State,
	}
	return r.Refs[ref].Flag
}
