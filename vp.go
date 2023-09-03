// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
)

// VP is the acronym for ValuePeeker, ValueParser and ValuePrinter.
type VP[T any] interface {
	// Type returns the VPType handled by the VP.
	//
	// NOTE: It SHOULD return VPTypeUnkown for types cannot represented by
	// VPType, see definition of VPType for more details.
	Type() VPType

	// HasValue returns true if v is considered having value.
	HasValue(v T) bool

	// ParseValue parses an arg string as value of type T.
	//
	// NOTE: Implementation MUST handle nil ParseOptions.
	ParseValue(opts *ParseOptions, arg string, out T, set bool) error

	// PrintValue writes text representation of v to out.
	PrintValue(out io.Writer, value T) (int, error)
}

// VPType represents the type a VP is handling.
//
// It is limited to one of following types:
//
//   - scalar type.
//   - slice of scalar.
//   - map of scalar (key) to scalar/slice (value).
//
// Memory layout:
//
//	map elem variant: VPTypeMapElemVariantSlice/VPTypeMapElemVariantSum
//	       |
//	       |
//	0x     0     0     0     0     0     0     0     0
//	           [map key scalar]    |    [ value scalar ]
//	                               |
//	   type variant: VPTypeVariantSlice/VPTypeVariantSum as variant for the
//	                 value scalar, all bits on the left should be 0.
//	                 VPTypeVariantMap as variant for the whole type, in this
//	                 case, map elem variant is for the value scalar
//
//	       [   map only bits   ]
//	  (only used when type variant is map)
type VPType uint32

const (
	VPTypeUnknown VPType = iota
	VPTypeString
	VPTypeBool
	VPTypeInt
	VPTypeUint
	VPTypeFloat
	VPTypeSize
	VPTypeDuration
	VPTypeTime
	VPTypeTimestampUnixSec
	VPTypeTimestampUnixMilli
	VPTypeTimestampUnixMicro
	VPTypeTimestampUnixNano
	VPTypeRegexp
	VPTypeRegexpNocase

	VPTypeScalarMAX

	VPTypeElemScalarMASK VPType = 0x00000fff
	VPTypeKeyScalarMASK  VPType = 0x0fff0000

	// variant mask for the whole type
	VPTypeVariantMASK  VPType = 0x0000f000
	VPTypeVariantSlice VPType = 0x00001000
	VPTypeVariantSum   VPType = 0x00002000
	VPTypeVariantMap   VPType = 0x00003000

	// variant mask for the map value type (only used when VPTypeVariantMap is set)
	VPTypeMapElemVariantMASK  VPType = 0xf0000000
	VPTypeMapElemVariantSlice VPType = 0x10000000
	VPTypeMapElemVariantSum   VPType = 0x20000000

	VPTypeVariantShift        VPType = 12
	VPTypeKeyScalarShift      VPType = 16
	VPTypeMapElemVariantShift VPType = 28
)

// String returns an empty string if t is unknown.
func (t VPType) String() string {
	switch t & VPTypeVariantMASK {
	case 0:
		switch t & VPTypeElemScalarMASK {
		case VPTypeString:
			return "str"
		case VPTypeBool:
			return "bool"
		case VPTypeInt:
			return "int"
		case VPTypeUint:
			return "uint"
		case VPTypeFloat:
			return "float"
		case VPTypeSize:
			return "size"
		case VPTypeDuration:
			return "dur"
		case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
			return "time"
		case VPTypeRegexp, VPTypeRegexpNocase:
			return "regexp"
		}
	case VPTypeVariantSlice:
		switch t & VPTypeElemScalarMASK {
		case VPTypeString:
			return "[]str"
		case VPTypeBool:
			return "[]bool"
		case VPTypeInt:
			return "[]int"
		case VPTypeUint:
			return "[]uint"
		case VPTypeFloat:
			return "[]float"
		case VPTypeSize:
			return "[]size"
		case VPTypeDuration:
			return "[]dur"
		case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
			return "[]time"
		case VPTypeRegexp, VPTypeRegexpNocase:
			return "[]regexp"
		}
	case VPTypeVariantSum:
		switch t & VPTypeElemScalarMASK {
		case VPTypeInt:
			return "isum"
		case VPTypeUint:
			return "usum"
		case VPTypeFloat:
			return "fsum"
		case VPTypeSize:
			return "ssum"
		case VPTypeDuration:
			return "dsum"
		}
	case VPTypeVariantMap:
		switch t & VPTypeMapElemVariantMASK {
		case 0:
			switch t & VPTypeElemScalarMASK {
			case VPTypeString:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]str"
				case VPTypeBool:
					return "map[bool]str"
				case VPTypeInt:
					return "map[int]str"
				case VPTypeUint:
					return "map[uint]str"
				case VPTypeFloat:
					return "map[float]str"
				case VPTypeSize:
					return "map[size]str"
				case VPTypeDuration:
					return "map[dur]str"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]str"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]str"
				}
			case VPTypeBool:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]bool"
				case VPTypeBool:
					return "map[bool]bool"
				case VPTypeInt:
					return "map[int]bool"
				case VPTypeUint:
					return "map[uint]bool"
				case VPTypeFloat:
					return "map[float]bool"
				case VPTypeSize:
					return "map[size]bool"
				case VPTypeDuration:
					return "map[dur]bool"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]bool"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]bool"
				}
			case VPTypeInt:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]int"
				case VPTypeBool:
					return "map[bool]int"
				case VPTypeInt:
					return "map[int]int"
				case VPTypeUint:
					return "map[uint]int"
				case VPTypeFloat:
					return "map[float]int"
				case VPTypeSize:
					return "map[size]int"
				case VPTypeDuration:
					return "map[dur]int"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]int"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]int"
				}
			case VPTypeUint:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]uint"
				case VPTypeBool:
					return "map[bool]uint"
				case VPTypeInt:
					return "map[int]uint"
				case VPTypeUint:
					return "map[uint]uint"
				case VPTypeFloat:
					return "map[float]uint"
				case VPTypeSize:
					return "map[size]uint"
				case VPTypeDuration:
					return "map[dur]uint"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]uint"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]uint"
				}
			case VPTypeFloat:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]float"
				case VPTypeBool:
					return "map[bool]float"
				case VPTypeInt:
					return "map[int]float"
				case VPTypeUint:
					return "map[uint]float"
				case VPTypeFloat:
					return "map[float]float"
				case VPTypeSize:
					return "map[size]float"
				case VPTypeDuration:
					return "map[dur]float"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]float"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]float"
				}
			case VPTypeSize:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]size"
				case VPTypeBool:
					return "map[bool]size"
				case VPTypeInt:
					return "map[int]size"
				case VPTypeUint:
					return "map[uint]size"
				case VPTypeFloat:
					return "map[float]size"
				case VPTypeSize:
					return "map[size]size"
				case VPTypeDuration:
					return "map[dur]size"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]size"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]size"
				}
			case VPTypeDuration:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]dur"
				case VPTypeBool:
					return "map[bool]dur"
				case VPTypeInt:
					return "map[int]dur"
				case VPTypeUint:
					return "map[uint]dur"
				case VPTypeFloat:
					return "map[float]dur"
				case VPTypeSize:
					return "map[size]dur"
				case VPTypeDuration:
					return "map[dur]dur"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]dur"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]dur"
				}
			case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]time"
				case VPTypeBool:
					return "map[bool]time"
				case VPTypeInt:
					return "map[int]time"
				case VPTypeUint:
					return "map[uint]time"
				case VPTypeFloat:
					return "map[float]time"
				case VPTypeSize:
					return "map[size]time"
				case VPTypeDuration:
					return "map[dur]time"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]time"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]time"
				}
			case VPTypeRegexp, VPTypeRegexpNocase:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]regexp"
				case VPTypeBool:
					return "map[bool]regexp"
				case VPTypeInt:
					return "map[int]regexp"
				case VPTypeUint:
					return "map[uint]regexp"
				case VPTypeFloat:
					return "map[float]regexp"
				case VPTypeSize:
					return "map[size]regexp"
				case VPTypeDuration:
					return "map[dur]regexp"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]regexp"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]regexp"
				}
			}
		case VPTypeMapElemVariantSlice:
			switch t & VPTypeElemScalarMASK {
			case VPTypeString:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]str"
				case VPTypeBool:
					return "map[bool][]str"
				case VPTypeInt:
					return "map[int][]str"
				case VPTypeUint:
					return "map[uint][]str"
				case VPTypeFloat:
					return "map[float][]str"
				case VPTypeSize:
					return "map[size][]str"
				case VPTypeDuration:
					return "map[dur][]str"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]str"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]str"
				}
			case VPTypeBool:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]bool"
				case VPTypeBool:
					return "map[bool][]bool"
				case VPTypeInt:
					return "map[int][]bool"
				case VPTypeUint:
					return "map[uint][]bool"
				case VPTypeFloat:
					return "map[float][]bool"
				case VPTypeSize:
					return "map[size][]bool"
				case VPTypeDuration:
					return "map[dur][]bool"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]bool"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]bool"
				}
			case VPTypeInt:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]int"
				case VPTypeBool:
					return "map[bool][]int"
				case VPTypeInt:
					return "map[int][]int"
				case VPTypeUint:
					return "map[uint][]int"
				case VPTypeFloat:
					return "map[float][]int"
				case VPTypeSize:
					return "map[size][]int"
				case VPTypeDuration:
					return "map[dur][]int"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]int"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]int"
				}
			case VPTypeUint:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]uint"
				case VPTypeBool:
					return "map[bool][]uint"
				case VPTypeInt:
					return "map[int][]uint"
				case VPTypeUint:
					return "map[uint][]uint"
				case VPTypeFloat:
					return "map[float][]uint"
				case VPTypeSize:
					return "map[size][]uint"
				case VPTypeDuration:
					return "map[dur][]uint"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]uint"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]uint"
				}
			case VPTypeFloat:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]float"
				case VPTypeBool:
					return "map[bool][]float"
				case VPTypeInt:
					return "map[int][]float"
				case VPTypeUint:
					return "map[uint][]float"
				case VPTypeFloat:
					return "map[float][]float"
				case VPTypeSize:
					return "map[size][]float"
				case VPTypeDuration:
					return "map[dur][]float"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]float"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]float"
				}
			case VPTypeSize:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]size"
				case VPTypeBool:
					return "map[bool][]size"
				case VPTypeInt:
					return "map[int][]size"
				case VPTypeUint:
					return "map[uint][]size"
				case VPTypeFloat:
					return "map[float][]size"
				case VPTypeSize:
					return "map[size][]size"
				case VPTypeDuration:
					return "map[dur][]size"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]size"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]size"
				}
			case VPTypeDuration:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]dur"
				case VPTypeBool:
					return "map[bool][]dur"
				case VPTypeInt:
					return "map[int][]dur"
				case VPTypeUint:
					return "map[uint][]dur"
				case VPTypeFloat:
					return "map[float][]dur"
				case VPTypeSize:
					return "map[size][]dur"
				case VPTypeDuration:
					return "map[dur][]dur"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]dur"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]dur"
				}
			case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]time"
				case VPTypeBool:
					return "map[bool][]time"
				case VPTypeInt:
					return "map[int][]time"
				case VPTypeUint:
					return "map[uint][]time"
				case VPTypeFloat:
					return "map[float][]time"
				case VPTypeSize:
					return "map[size][]time"
				case VPTypeDuration:
					return "map[dur][]time"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]time"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]time"
				}
			case VPTypeRegexp, VPTypeRegexpNocase:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str][]regexp"
				case VPTypeBool:
					return "map[bool][]regexp"
				case VPTypeInt:
					return "map[int][]regexp"
				case VPTypeUint:
					return "map[uint][]regexp"
				case VPTypeFloat:
					return "map[float][]regexp"
				case VPTypeSize:
					return "map[size][]regexp"
				case VPTypeDuration:
					return "map[dur][]regexp"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time][]regexp"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp][]regexp"
				}
			}
		case VPTypeMapElemVariantSum:
			switch t & VPTypeElemScalarMASK {
			case VPTypeInt:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]isum"
				case VPTypeBool:
					return "map[bool]isum"
				case VPTypeInt:
					return "map[int]isum"
				case VPTypeUint:
					return "map[uint]isum"
				case VPTypeFloat:
					return "map[float]isum"
				case VPTypeSize:
					return "map[size]isum"
				case VPTypeDuration:
					return "map[dur]isum"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]isum"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]isum"
				}
			case VPTypeUint:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]usum"
				case VPTypeBool:
					return "map[bool]usum"
				case VPTypeInt:
					return "map[int]usum"
				case VPTypeUint:
					return "map[uint]usum"
				case VPTypeFloat:
					return "map[float]usum"
				case VPTypeSize:
					return "map[size]usum"
				case VPTypeDuration:
					return "map[dur]usum"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]usum"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]usum"
				}
			case VPTypeFloat:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]fsum"
				case VPTypeBool:
					return "map[bool]fsum"
				case VPTypeInt:
					return "map[int]fsum"
				case VPTypeUint:
					return "map[uint]fsum"
				case VPTypeFloat:
					return "map[float]fsum"
				case VPTypeSize:
					return "map[size]fsum"
				case VPTypeDuration:
					return "map[dur]fsum"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]fsum"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]fsum"
				}
			case VPTypeSize:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]ssum"
				case VPTypeBool:
					return "map[bool]ssum"
				case VPTypeInt:
					return "map[int]ssum"
				case VPTypeUint:
					return "map[uint]ssum"
				case VPTypeFloat:
					return "map[float]ssum"
				case VPTypeSize:
					return "map[size]ssum"
				case VPTypeDuration:
					return "map[dur]ssum"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]ssum"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]ssum"
				}
			case VPTypeDuration:
				switch (t & VPTypeKeyScalarMASK) >> VPTypeKeyScalarShift {
				case VPTypeString:
					return "map[str]dsum"
				case VPTypeBool:
					return "map[bool]dsum"
				case VPTypeInt:
					return "map[int]dsum"
				case VPTypeUint:
					return "map[uint]dsum"
				case VPTypeFloat:
					return "map[float]dsum"
				case VPTypeSize:
					return "map[size]dsum"
				case VPTypeDuration:
					return "map[dur]dsum"
				case VPTypeTime, VPTypeTimestampUnixSec, VPTypeTimestampUnixMilli, VPTypeTimestampUnixMicro, VPTypeTimestampUnixNano:
					return "map[time]dsum"
				case VPTypeRegexp, VPTypeRegexpNocase:
					return "map[regexp]dsum"
				}
			}
		}
	}

	return ""
}
