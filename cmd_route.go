// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"io"
	"sort"
)

// Route contains the route leading from the root Cmd to the target Cmd.
type Route []*Cmd

// Root returns the first entry in the route.
//
// It returns nil if it is empty.
func (p Route) Root() *Cmd {
	if len(p) != 0 {
		return p[0]
	}

	return nil
}

// Push appends a cmd to the route.
func (p Route) Push(cmd *Cmd) Route {
	return append(p, cmd)
}

// Target returns the last *Cmd in route.
//
// It returns nil if it is empty.
func (p Route) Target() *Cmd {
	if len(p) == 0 {
		return nil
	}
	return p[len(p)-1]
}

// Up returns a Route containing all but the last *Cmd.
//
// It returns nil if it is empty.
func (p Route) Up() Route {
	if len(p) < 1 {
		return nil
	}
	return p[:len(p)-1]
}

// CheckFlagValueChanged implements [Inspector].
func (p *Route) CheckFlagValueChanged(name string) bool {
	flag, ok := p.FindFlag(name)
	if !ok {
		panic(&ErrFlagUndefined{Name: name})
	}

	return flag.State().ValueChanged()
}

// FindFlag implements [FlagFinder].
func (p *Route) FindFlag(name string) (f Flag, ok bool) {
	route := *p
	flags := p.Target().LocalFlags
	if flags != nil {
		f, ok = flags.FindFlag(name)
		if ok {
			return
		}
	}

	for i := len(route) - 1; i >= 0; i-- {
		flags = route[i].Flags
		if flags != nil {
			f, ok = flags.FindFlag(name)
			if ok {
				return
			}
		}
	}

	return nil, false
}

// NthFlag implements [FlagIter]
func (p *Route) NthFlag(i int) (info FlagInfo, ok bool) {
	route := *p
	c := route.Target()
	if c == nil {
		return
	}

	if c.LocalFlags != nil {
		var iter FlagIter
		iter, ok = c.LocalFlags.(FlagIter)
		if ok {
			info, ok = iter.NthFlag(i)
			if ok {
				return info, true
			}

			// - count of local flags
			i -= sort.Search(i, func(i int) bool {
				_, ok := iter.NthFlag(i)
				return !ok
			})
		}
	}

	for j := len(route) - 1; j >= 0; j-- {
		c = route[j]

		if c.Flags == nil {
			continue
		}

		var iter FlagIter
		iter, ok = c.Flags.(FlagIter)
		if ok {
			info, ok = iter.NthFlag(i)
			if ok {
				return
			}

			// - count of flags
			i -= sort.Search(i, func(i int) bool {
				_, ok := iter.NthFlag(i)
				return !ok
			})
		}
	}

	return FlagInfo{}, false
}

// FormatRoute writes all Cmd.Name() in the route with sep in between.
//
// For the last Cmd in route (the target), it writes the complete Cmd.Pattern.
func FormatRoute(w io.Writer, route Route, sep string) (n int, err error) {
	if len(route) == 0 {
		return
	}

	var x int
	for i := 0; i < len(route)-1; i++ {
		x, err = wstr(w, route[i].Name())
		n += x
		if err != nil {
			return
		}

		x, err = wstr(w, sep)
		n += x
		if err != nil {
			return
		}
	}

	x, err = wstr(w, route.Target().Pattern)
	n += x
	if err != nil {
		return
	}

	return
}
