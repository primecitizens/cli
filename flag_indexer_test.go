// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 The Prime Citizens

package cli

import (
	"strconv"
	"strings"
	"testing"

	"github.com/primecitizens/cli/internal/assert"
)

var (
	_ FlagIndexer = FuncIndexer(nil)
	_ FlagIndexer = (*MapIndexer)(nil)
	_ FlagIndexer = (*MultiIndexer)(nil)
	_ FlagIndexer = (*LevelIndexer)(nil)
	_ FlagLevel   = (*LevelIndexer)(nil)
	_ FlagIndexer = (*ReflectIndexer)(nil)
)

func assertFlagTrue(t *testing.T, f Flag, ok bool) {
	if !ok || f == nil {
		t.Errorf("want (nil=false, true), got (nil=%v, %v)", f == nil, ok)
	}
}

func assertNoflagFalse(t *testing.T, f Flag, ok bool) {
	if ok || f != nil {
		t.Errorf("want (nil=true, false), got (nil=%v, %v)", f == nil, ok)
	}
}

func TestFuncIndexer(t *testing.T) {
	t.Run("Search", func(t *testing.T) {
		const (
			FLAG      = "foo"
			shorthand = "f"
		)
		indexer := FuncIndexer(func(flag string, index int) (f Flag, info FlagInfo, ok bool) {
			assert.Eq(t, -1, index)

			switch flag {
			case FLAG:
				return &Bool{}, FlagInfo{Name: FLAG, Shorthand: shorthand}, true
			default:
				return
			}
		})

		f, ok := indexer.FindFlag(FLAG)
		assertFlagTrue(t, f, ok)

		f, ok = indexer.FindFlag("non-existing")
		assertNoflagFalse(t, f, ok)
	})

	t.Run("Index", func(t *testing.T) {
		const (
			N = 5
		)
		indexer := FuncIndexer(func(flag string, index int) (f Flag, info FlagInfo, ok bool) {
			assert.Eq(t, "", flag)

			if index < N {
				name := strconv.FormatInt(int64(index), 10)
				return &Bool{}, FlagInfo{Name: "foo" + name, Shorthand: name[:1]}, true
			}

			return
		})

		for i := 0; i < N; i++ {
			info, ok := indexer.NthFlag(i)
			assert.True(t, ok)
			assert.Eq(t, "foo"+strconv.FormatInt(int64(i), 10), info.Name)
			assert.Eq(t, strconv.FormatInt(int64(i), 10)[:1], info.Shorthand)
		}
	})
}

func TestMapIndexer(t *testing.T) {
	indexer := NewMapIndexer().
		Add(&Bool{}, "foo", "f").
		Add(&Bool{}, "bar", "b")

	testIndexer(t, indexer)
}

func TestMultiIndexer(t *testing.T) {
	indexer := &MultiIndexer{
		Flags: []FlagFinderMaybeIter{
			NewMapIndexer().Add(&Bool{}, "foo", "f"),
			FuncIndexer(func(flag string, index int) (f Flag, info FlagInfo, ok bool) {
				if flag == "bar" || index == 0 {
					return &Bool{}, FlagInfo{Name: "bar", Shorthand: "b"}, true
				}

				return
			}),
		},
	}

	testIndexer(t, indexer)
}

func testIndexer(t *testing.T, indexer FlagIndexer) {
	f, ok := indexer.FindFlag("f")
	assertFlagTrue(t, f, ok)

	f, ok = indexer.FindFlag("bar")
	assertFlagTrue(t, f, ok)

	f, ok = indexer.FindFlag("non-existing")
	assertNoflagFalse(t, f, ok)

	info, ok := indexer.NthFlag(0)
	assert.True(t, ok)
	assert.Eq(t, "foo", info.Name)
	assert.Eq(t, "f", info.Shorthand)

	info, ok = indexer.NthFlag(1)
	assert.True(t, ok)
	assert.Eq(t, "bar", info.Name)
	assert.Eq(t, "b", info.Shorthand)

	info, ok = indexer.NthFlag(2)
	assert.False(t, ok)
	assert.Eq(t, "", info.Name)
	assert.Eq(t, "", info.Shorthand)
}

func TestLevelIndexer(t *testing.T) {
	var (
		root   = LevelIndexer{}
		levelA = LevelIndexer{
			Up:     &root,
			Prefix: "a-",
			Flags: NewMapIndexer().
				Add(&Bool{}, "fx", "x"),
		}
		levelAB = LevelIndexer{
			Up:     &levelA,
			Prefix: "b-",
		}
		levelAC = LevelIndexer{
			Up:     &levelA,
			Prefix: "c-",
		}
		levelABD = LevelIndexer{
			Up:     &levelAB,
			Prefix: "d-",
			Flags: NewMapIndexer().
				Add(&Bool{}, "fy", "y").
				Add(&Bool{}, "fz", "z"),
		}
	)

	t.Run("TrimAllLevelPrefixes", func(t *testing.T) {
		assert.Eq(t, "foo", root.TrimAllLevelPrefixes("foo"))
		assert.Eq(t, "foo", levelA.TrimAllLevelPrefixes("a-foo"))
		assert.Eq(t, "foo", levelAC.TrimAllLevelPrefixes("a-c-foo"))
		assert.Eq(t, "foo", levelABD.TrimAllLevelPrefixes("a-b-d-foo"))
	})

	t.Run("GetFullFlagName", func(t *testing.T) {
		assert.Eq(t, "a-b-d-foo", levelABD.GetFullFlagName("foo"))
	})

	t.Run("FindFlag", func(t *testing.T) {
		f, ok := levelABD.FindFlag("a-b-d-fy")
		assertFlagTrue(t, f, ok)

		f, ok = levelABD.FindFlag("a-b-d-fz")
		assertFlagTrue(t, f, ok)

		f, ok = levelABD.FindFlag("y")
		assertFlagTrue(t, f, ok)

		f, ok = levelABD.FindFlag("z")
		assertFlagTrue(t, f, ok)

		f, ok = levelABD.FindFlag("foo")
		assertNoflagFalse(t, f, ok)
	})

	t.Run("Integration", func(t *testing.T) {
		rootCmd := Cmd{
			Pattern: "root",
			Flags:   &root,
			Children: []*Cmd{
				{
					Pattern: "A",
					Flags:   &levelA,
					Children: []*Cmd{
						{
							Pattern: "B",
							Flags:   &levelAB,
							Children: []*Cmd{
								{
									Pattern: "D",
									Flags:   &levelABD,
								},
							},
						},
						{
							Pattern: "C",
							Flags:   &levelAC,
						},
					},
				},
			},
		}

		var sb strings.Builder
		_ = HandleArgErrorAsHelpRequest(
			&CmdOptions{
				Stderr: &sb,
			},
			Route{
				&rootCmd,
				rootCmd.Children[0],
				rootCmd.Children[0].Children[0],
				rootCmd.Children[0].Children[0].Children[0],
			},
			nil,
			0,
			ErrTimeout{},
		)

		assert.Eq(
			t,
			""+
				"Error: timeout\n"+
				"\n"+
				"root A B D\n"+
				"\n"+
				"Flags:\n"+
				"  -y --a-b-d-fy bool\n"+
				"  -z --a-b-d-fz bool\n"+
				"  -x --a-fx bool\n",
			sb.String(),
		)
	})
}
