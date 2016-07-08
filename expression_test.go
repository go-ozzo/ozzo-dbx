// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExp(t *testing.T) {
	params := Params{"k2": "v2"}

	e1 := NewExp("s1").(*Exp)
	assert.Equal(t, e1.Build(nil, params), "s1", "e1.Build()")
	assert.Equal(t, len(params), 1, `len(params)@1`)

	e2 := NewExp("s2", Params{"k1": "v1"}).(*Exp)
	assert.Equal(t, e2.Build(nil, params), "s2", "e2.Build()")
	assert.Equal(t, len(params), 2, `len(params)@2`)
}

func TestHashExp(t *testing.T) {
	e1 := HashExp{}
	assert.Equal(t, e1.Build(nil, nil), "", `e1.Build()`)

	e2 := HashExp{
		"k1": nil,
		"k2": NewExp("s1", Params{"ka": "va"}),
		"k3": 1.1,
		"k4": "abc",
		"k5": []interface{}{1, 2},
	}
	db := getDB()
	params := Params{"k0": "v0"}
	expected := "`k1` IS NULL AND (s1) AND `k3`={:p2} AND `k4`={:p3} AND `k5` IN ({:p4}, {:p5})"

	assert.Equal(t, e2.Build(db, params), expected, `e2.Build()`)
	assert.Equal(t, len(params), 6, `len(params)`)
	assert.Equal(t, params["p5"].(int), 2, `params["p5"]`)
}

func TestNotExp(t *testing.T) {
	e1 := Not(NewExp("s1"))
	assert.Equal(t, e1.Build(nil, nil), "NOT (s1)", `e1.Build()`)

	e2 := Not(NewExp(""))
	assert.Equal(t, e2.Build(nil, nil), "", `e2.Build()`)
}

func TestAndOrExp(t *testing.T) {
	e1 := And(NewExp("s1", Params{"k1": "v1"}), NewExp(""), NewExp("s2", Params{"k2": "v2"}))
	params := Params{}
	assert.Equal(t, e1.Build(nil, params), "(s1) AND (s2)", `e1.Build()`)
	assert.Equal(t, len(params), 2, `len(params)`)

	e2 := Or(NewExp("s1"), NewExp("s2"))
	assert.Equal(t, e2.Build(nil, nil), "(s1) OR (s2)", `e2.Build()`)

	e3 := And()
	assert.Equal(t, e3.Build(nil, nil), "", `e3.Build()`)

	e4 := And(NewExp("s1"))
	assert.Equal(t, e4.Build(nil, nil), "s1", `e4.Build()`)

	e5 := And(NewExp("s1"), nil)
	assert.Equal(t, e5.Build(nil, nil), "s1", `e5.Build()`)
}

func TestInExp(t *testing.T) {
	db := getDB()

	e1 := In("age", 1, 2, 3)
	params := Params{}
	assert.Equal(t, e1.Build(db, params), "`age` IN ({:p0}, {:p1}, {:p2})", `e1.Build()`)
	assert.Equal(t, len(params), 3, `len(params)@1`)

	e2 := In("age", 1)
	params = Params{}
	assert.Equal(t, e2.Build(db, params), "`age`={:p0}", `e2.Build()`)
	assert.Equal(t, len(params), 1, `len(params)@2`)

	e3 := NotIn("age", 1, 2, 3)
	params = Params{}
	assert.Equal(t, e3.Build(db, params), "`age` NOT IN ({:p0}, {:p1}, {:p2})", `e3.Build()`)
	assert.Equal(t, len(params), 3, `len(params)@3`)

	e4 := NotIn("age", 1)
	params = Params{}
	assert.Equal(t, e4.Build(db, params), "`age`<>{:p0}", `e4.Build()`)
	assert.Equal(t, len(params), 1, `len(params)@4`)

	e5 := In("age")
	assert.Equal(t, e5.Build(db, nil), "0=1", `e5.Build()`)

	e6 := NotIn("age")
	assert.Equal(t, e6.Build(db, nil), "", `e6.Build()`)
}

func TestLikeExp(t *testing.T) {
	db := getDB()

	e1 := Like("name", "a", "b", "c")
	params := Params{}
	assert.Equal(t, e1.Build(db, params), "`name` LIKE {:p0} AND `name` LIKE {:p1} AND `name` LIKE {:p2}", `e1.Build()`)
	assert.Equal(t, len(params), 3, `len(params)@1`)

	e2 := Like("name", "a")
	params = Params{}
	assert.Equal(t, e2.Build(db, params), "`name` LIKE {:p0}", `e2.Build()`)
	assert.Equal(t, len(params), 1, `len(params)@2`)

	e3 := Like("name")
	assert.Equal(t, e3.Build(db, nil), "", `e3.Build()`)

	e4 := NotLike("name", "a", "b", "c")
	params = Params{}
	assert.Equal(t, e4.Build(db, params), "`name` NOT LIKE {:p0} AND `name` NOT LIKE {:p1} AND `name` NOT LIKE {:p2}", `e4.Build()`)
	assert.Equal(t, len(params), 3, `len(params)@4`)

	e5 := OrLike("name", "a", "b", "c")
	params = Params{}
	assert.Equal(t, e5.Build(db, params), "`name` LIKE {:p0} OR `name` LIKE {:p1} OR `name` LIKE {:p2}", `e5.Build()`)
	assert.Equal(t, len(params), 3, `len(params)@5`)

	e6 := OrNotLike("name", "a", "b", "c")
	params = Params{}
	assert.Equal(t, e6.Build(db, params), "`name` NOT LIKE {:p0} OR `name` NOT LIKE {:p1} OR `name` NOT LIKE {:p2}", `e6.Build()`)
	assert.Equal(t, len(params), 3, `len(params)@6`)

	e7 := Like("name", "a_\\%")
	params = Params{}
	e7.Build(db, params)
	assert.Equal(t, params["p0"], "%a\\_\\\\\\%%", `params["p0"]@1`)

	e8 := Like("name", "a").Match(false, true)
	params = Params{}
	e8.Build(db, params)
	assert.Equal(t, params["p0"], "a%", `params["p0"]@2`)

	e9 := Like("name", "a").Match(true, false)
	params = Params{}
	e9.Build(db, params)
	assert.Equal(t, params["p0"], "%a", `params["p0"]@3`)

	e10 := Like("name", "a").Match(false, false)
	params = Params{}
	e10.Build(db, params)
	assert.Equal(t, params["p0"], "a", `params["p0"]@4`)

	e11 := Like("name", "%a").Match(false, false).Escape()
	params = Params{}
	e11.Build(db, params)
	assert.Equal(t, params["p0"], "%a", `params["p0"]@5`)
}

func TestBetweenExp(t *testing.T) {
	db := getDB()

	e1 := Between("age", 30, 40)
	params := Params{}
	assert.Equal(t, e1.Build(db, params), "`age` BETWEEN {:p0} AND {:p1}", `e1.Build()`)
	assert.Equal(t, len(params), 2, `len(params)@1`)

	e2 := NotBetween("age", 30, 40)
	params = Params{}
	assert.Equal(t, e2.Build(db, params), "`age` NOT BETWEEN {:p0} AND {:p1}", `e2.Build()`)
	assert.Equal(t, len(params), 2, `len(params)@2`)
}

func TestExistsExp(t *testing.T) {
	e1 := Exists(NewExp("s1"))
	assert.Equal(t, e1.Build(nil, nil), "EXISTS (s1)", `e1.Build()`)

	e2 := NotExists(NewExp("s1"))
	assert.Equal(t, e2.Build(nil, nil), "NOT EXISTS (s1)", `e2.Build()`)

	e3 := Exists(NewExp(""))
	assert.Equal(t, e3.Build(nil, nil), "0=1", `e3.Build()`)

	e4 := NotExists(NewExp(""))
	assert.Equal(t, e4.Build(nil, nil), "", `e4.Build()`)
}
