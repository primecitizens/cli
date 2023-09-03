// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"sort"
	"strings"
)

// FlagInfo is a pack of the flag name, shorthand, default value and current
// state.
type FlagInfo struct {
	// Name is the long flag name (the long one).
	Name string
	// Shorthand is the flag shorthand.
	Shorthand string

	// DefaultValue is the default value used for the flag.
	//
	// Please use following format to provide non-scalar default value:
	//	'[' + entry1 + ', ' + entry2 ... ']'
	//
	// NOTE: For non-scalar values, when default value is being assigned
	// by Cmd.Exec, it checks prefix '[' and suffix ']', and splits
	// elements by cutting around ', '.
	DefaultValue string

	// State is the current state of the flag.
	State FlagState
}

// FlagIter
type FlagIter interface {
	// NthFlag returns the i-th flag's info this iterator can find.
	//
	// The bool return value indicates whether there is i-th flag, on
	// returning false, call with any value greater than i should return
	// false as well.
	NthFlag(i int) (FlagInfo, bool)
}

// FlagFinder
type FlagFinder interface {
	// FindFlag searches flags known to this FlagFinder by name.
	//
	// The name can be either a full flag name or a flag shorthand, and
	// doesn't contain the POSIX & GNU flag prefix (`-` and `--`).
	FindFlag(name string) (Flag, bool)
}

// FindFlag tries to find a Flag from the FlagFinder with a list of flag names.
//
// On returning found = true, name is the one used to find the flag.
func FindFlag[F FlagFinder](flags F, names ...string) (name string, flag Flag, found bool) {
	names = noescapeSlice(names)
	for _, name = range names {
		if len(name) == 0 {
			continue
		}

		flag, found = flags.FindFlag(name)
		if found {
			return
		}
	}

	return "", nil, false
}

// FlagFinderMaybeIter is an alias of FlagFinder but indicates the
// implementation may have additional FlagIter support.
type FlagFinderMaybeIter = FlagFinder

// FlagIndexer is the combination of FlagFinder and FlagIter.
type FlagIndexer interface {
	FlagFinder
	FlagIter
}

// FuncIndexer wraps a function as FlagIndexer.
//
// when index < 0, act as FlagFinder, otherwise act as FlagIter.
type FuncIndexer func(flag string, index int) (f Flag, info FlagInfo, ok bool)

// FindFlag implements [FlagFinder].
func (fn FuncIndexer) FindFlag(name string) (Flag, bool) {
	if f, _, ok := fn(name, -1); ok {
		return f, true
	}

	return nil, false
}

// NthFlag implements [FlagIter].
func (fn FuncIndexer) NthFlag(i int) (info FlagInfo, ok bool) {
	if i < 0 {
		return info, false
	}

	_, info, ok = fn("", i)
	return
}

// NewMapIndexer creates a new MapIndexer.
func NewMapIndexer() *MapIndexer {
	return &MapIndexer{
		n2i: map[string]int{},
		i2f: map[int]*flagBundle{},
	}
}

type flagBundle struct {
	flag Flag
	info FlagInfo
}

// MapIndexer implements [FlagIndexer] using built-in maps.
type MapIndexer struct {
	next int

	n2i map[string]int      // name to index
	i2f map[int]*flagBundle // index to flagBundle
}

// Add adds a flag with its names.
//
// It panics when name is empty or there is flag with same name shorthand.
func (m *MapIndexer) Add(flag Flag, names ...string) *MapIndexer {
	return m.AddWithDefaultValue("", flag, names...)
}

// AddWithDefaultValue is Add but provides default value to the flag.
func (m *MapIndexer) AddWithDefaultValue(defaultValue string, flag Flag, names ...string) *MapIndexer {
	if len(names) == 0 {
		panic("invalid empty name.")
	}

	index := m.next
	m.next++

	fb := flagBundle{
		flag: flag,
		info: FlagInfo{
			DefaultValue: defaultValue,
		},
	}
	for _, name := range names {
		if len(name) == 0 {
			panic("invalid empty name.")
		}

		_, alreadyHave := m.n2i[name]
		if alreadyHave {
			panic(&ErrDuplicateFlag{name})
		}

		m.n2i[name] = index
		if IsShorthand(name) {
			if len(fb.info.Shorthand) == 0 {
				fb.info.Shorthand = name
			}
		} else {
			if len(fb.info.Name) == 0 {
				fb.info.Name = name
			}
		}
	}

	m.i2f[index] = &fb
	return m
}

// FindFlag implements [FlagFinder].
func (m *MapIndexer) FindFlag(name string) (Flag, bool) {
	i, ok := m.n2i[name]
	if ok {
		if fb := m.i2f[i]; fb != nil {
			return fb.flag, true
		}
	}

	return nil, false
}

// NthFlag implements [FlagIter].
func (m *MapIndexer) NthFlag(i int) (info FlagInfo, ok bool) {
	fb := m.i2f[i]
	if fb == nil {
		return
	}

	ok = true
	info = fb.info
	info.State = fb.flag.State()
	return
}

// MultiIndexer combines multiple FlagFinders into one.
type MultiIndexer struct {
	Flags []FlagFinderMaybeIter
}

// FindFlag implements [FlagFinder].
func (m *MultiIndexer) FindFlag(name string) (f Flag, ok bool) {
	for _, fi := range m.Flags {
		f, ok = fi.FindFlag(name)
		if ok {
			return
		}
	}

	return nil, false
}

// NthFlag implements [FlagIter].
func (m *MultiIndexer) NthFlag(i int) (info FlagInfo, ok bool) {
	for _, fi := range m.Flags {
		var iter FlagIter
		iter, ok = fi.(FlagIter)
		if ok {
			info, ok = iter.NthFlag(i)
			if ok {
				return
			}

			i -= sort.Search(i, func(i int) bool {
				_, ok := iter.NthFlag(i)
				return !ok
			})
		}
	}

	return
}

// FlagLevel
type FlagLevel interface {
	// TrimAllLevelPrefixes tirms all prefixes belonging to each level.
	TrimAllLevelPrefixes(name string) string

	// GetFullFlagName adds all prefixes to name.
	GetFullFlagName(name string) string
}

// LevelIndexer is a FlagFinder wrapper to build multi-level flag hierarchy.
type LevelIndexer struct {
	// Up points to the up level, if any.
	Up FlagLevel

	// Prefix is the prefix to identify this level.
	//
	// Should not be empty for non-root level.
	Prefix string

	// Flags
	Flags FlagFinderMaybeIter

	fullnames map[string]string
}

// TrimAllLevelPrefixes implements [FlagLevel].
func (l *LevelIndexer) TrimAllLevelPrefixes(name string) string {
	if l.Up != nil {
		name = l.Up.TrimAllLevelPrefixes(name)
	}

	return strings.TrimPrefix(name, l.Prefix)
}

// GetFullFlagName implements [FlagLevel].
func (l *LevelIndexer) GetFullFlagName(name string) string {
	if IsShorthand(name) {
		return name
	}

	if len(l.Prefix) != 0 {
		name = l.Prefix + name
	}

	if l.Up != nil {
		name = l.Up.GetFullFlagName(name)
	}

	return name
}

// NthFlag implements [FlagIter].
func (l *LevelIndexer) NthFlag(i int) (info FlagInfo, ok bool) {
	if l.Flags == nil {
		return
	}

	iter, ok := l.Flags.(FlagIter)
	if !ok {
		return
	}

	info, ok = iter.NthFlag(i)
	if ok && len(info.Name) != 0 {
		info.Name = l.GetFullFlagName(info.Name)
	}

	return
}

// FindFlag implements [FlagFinder].
func (l *LevelIndexer) FindFlag(name string) (Flag, bool) {
	if l.Flags == nil {
		return nil, false
	}

	if IsShorthand(name) {
		return l.Flags.FindFlag(name)
	}

	return l.Flags.FindFlag(l.TrimAllLevelPrefixes(name))
}
