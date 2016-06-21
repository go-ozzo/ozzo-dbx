// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbx

import (
	"fmt"
	"sort"
	"strings"
)

// Expression represents a DB expression that can be embedded in a SQL statement.
type Expression interface {
	// Build converts an expression into a SQL fragment.
	// If the expression contains binding parameters, they will be added to the given Params.
	Build(*DB, Params) string
}

// HashExp represents a hash expression.
//
// A hash expression is a map whose keys are DB column names which need to be filtered according
// to the corresponding values. For example, HashExp{"level": 2, "dept": 10} will generate
// the SQL: "level"=2 AND "dept"=10.
//
// HashExp also handles nil values and slice values. For example, HashExp{"level": []interface{}{1, 2}, "dept": nil}
// will generate: "level" IN (1, 2) AND "dept" IS NULL.
type HashExp map[string]interface{}

// NewExp generates an expression with the specified SQL fragment and the optional binding parameters.
func NewExp(e string, params ...Params) Expression {
	if len(params) > 0 {
		return &Exp{e, params[0]}
	}
	return &Exp{e, nil}
}

// Not generates a NOT expression which prefixes "NOT" to the specified expression.
func Not(e Expression) Expression {
	return &NotExp{e}
}

// And generates an AND expression which concatenates the given expressions with "AND".
func And(exps ...Expression) Expression {
	return &AndOrExp{exps, "AND"}
}

// Or generates an OR expression which concatenates the given expressions with "OR".
func Or(exps ...Expression) Expression {
	return &AndOrExp{exps, "OR"}
}

// In generates an IN expression for the specified column and the list of allowed values.
// If values is empty, a SQL "0=1" will be generated which represents a false expression.
func In(col string, values ...interface{}) Expression {
	return &InExp{col, values, false}
}

// NotIn generates an NOT IN expression for the specified column and the list of disallowed values.
// If values is empty, an empty string will be returned indicating a true expression.
func NotIn(col string, values ...interface{}) Expression {
	return &InExp{col, values, true}
}

// DefaultLikeEscape specifies the default special character escaping for LIKE expressions
// The strings at 2i positions are the special characters to be escaped while those at 2i+1 positions
// are the corresponding escaped versions.
var DefaultLikeEscape = []string{"\\", "\\\\", "%", "\\%", "_", "\\_"}

// Like generates a LIKE expression for the specified column and the possible strings that the column should be like.
// If multiple values are present, the column should be like *all* of them. For example, Like("name", "key", "word")
// will generate a SQL expression: "name" LIKE "%key%" AND "name" LIKE "%word%".
//
// By default, each value will be surrounded by "%" to enable partial matching. If a value contains special characters
// such as "%", "\", "_", they will also be properly escaped.
//
// You may call Escape() and/or Match() to change the default behavior. For example, Like("name", "key").Match(false, true)
// generates "name" LIKE "key%".
func Like(col string, values ...string) *LikeExp {
	return &LikeExp{
		left:   true,
		right:  true,
		col:    col,
		values: values,
		escape: DefaultLikeEscape,
		Like:   "LIKE",
	}
}

// NotLike generates a NOT LIKE expression.
// For example, NotLike("name", "key", "word") will generate a SQL expression:
// "name" NOT LIKE "%key%" AND "name" NOT LIKE "%word%". Please see Like() for more details.
func NotLike(col string, values ...string) *LikeExp {
	return &LikeExp{
		left:   true,
		right:  true,
		col:    col,
		values: values,
		escape: DefaultLikeEscape,
		Like:   "NOT LIKE",
	}
}

// OrLike generates an OR LIKE expression.
// This is similar to Like() except that the column should be like one of the possible values.
// For example, OrLike("name", "key", "word") will generate a SQL expression:
// "name" LIKE "%key%" OR "name" LIKE "%word%". Please see Like() for more details.
func OrLike(col string, values ...string) *LikeExp {
	return &LikeExp{
		or:     true,
		left:   true,
		right:  true,
		col:    col,
		values: values,
		escape: DefaultLikeEscape,
		Like:   "LIKE",
	}
}

// OrNotLike generates an OR NOT LIKE expression.
// For example, OrNotLike("name", "key", "word") will generate a SQL expression:
// "name" NOT LIKE "%key%" OR "name" NOT LIKE "%word%". Please see Like() for more details.
func OrNotLike(col string, values ...string) *LikeExp {
	return &LikeExp{
		or:     true,
		left:   true,
		right:  true,
		col:    col,
		values: values,
		escape: DefaultLikeEscape,
		Like:   "NOT LIKE",
	}
}

// Exists generates an EXISTS expression by prefixing "EXISTS" to the given expression.
func Exists(exp Expression) Expression {
	return &ExistsExp{exp, false}
}

// NotExists generates an EXISTS expression by prefixing "NOT EXISTS" to the given expression.
func NotExists(exp Expression) Expression {
	return &ExistsExp{exp, true}
}

// Between generates a BETWEEN expression.
// For example, Between("age", 10, 30) generates: "age" BETWEEN 10 AND 30
func Between(col string, from, to interface{}) Expression {
	return &BetweenExp{col, from, to, false}
}

// NotBetween generates a NOT BETWEEN expression.
// For example, NotBetween("age", 10, 30) generates: "age" NOT BETWEEN 10 AND 30
func NotBetween(col string, from, to interface{}) Expression {
	return &BetweenExp{col, from, to, true}
}

// Exp represents an expression with a SQL fragment and a list of optional binding parameters.
type Exp struct {
	e      string
	params Params
}

// Build converts an expression into a SQL fragment.
func (e *Exp) Build(db *DB, params Params) string {
	if len(e.params) == 0 {
		return e.e
	}
	for k, v := range e.params {
		params[k] = v
	}
	return e.e
}

// Build converts an expression into a SQL fragment.
func (e HashExp) Build(db *DB, params Params) string {
	if len(e) == 0 {
		return ""
	}

	// ensure the hash exp generates the same SQL for different runs
	names := []string{}
	for name := range e {
		names = append(names, name)
	}
	sort.Strings(names)

	var parts []string
	for _, name := range names {
		value := e[name]
		switch value.(type) {
		case nil:
			name = db.QuoteColumnName(name)
			parts = append(parts, name+" IS NULL")
		case Expression:
			if sql := value.(Expression).Build(db, params); sql != "" {
				parts = append(parts, "("+sql+")")
			}
		case []interface{}:
			in := In(name, value.([]interface{})...)
			if sql := in.Build(db, params); sql != "" {
				parts = append(parts, sql)
			}
		default:
			pn := fmt.Sprintf("p%v", len(params))
			name = db.QuoteColumnName(name)
			parts = append(parts, name+"={:"+pn+"}")
			params[pn] = value
		}
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, " AND ")
}

// NotExp represents an expression that should prefix "NOT" to a specified expression.
type NotExp struct {
	e Expression
}

// Build converts an expression into a SQL fragment.
func (e *NotExp) Build(db *DB, params Params) string {
	if sql := e.e.Build(db, params); sql != "" {
		return "NOT (" + sql + ")"
	}
	return ""
}

// AndOrExp represents an expression that concatenates multiple expressions using either "AND" or "OR".
type AndOrExp struct {
	exps []Expression
	op   string
}

// Build converts an expression into a SQL fragment.
func (e *AndOrExp) Build(db *DB, params Params) string {
	if len(e.exps) == 0 {
		return ""
	}

	var parts []string
	for _, a := range e.exps {
		if a == nil {
			continue
		}
		if sql := a.Build(db, params); sql != "" {
			parts = append(parts, sql)
		}
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, ") "+e.op+" (") + ")"
}

// InExp represents an "IN" or "NOT IN" expression.
type InExp struct {
	col    string
	values []interface{}
	not    bool
}

// Build converts an expression into a SQL fragment.
func (e *InExp) Build(db *DB, params Params) string {
	if len(e.values) == 0 {
		if e.not {
			return ""
		}
		return "0=1"
	}

	var values []string
	for _, value := range e.values {
		switch value.(type) {
		case nil:
			values = append(values, "NULL")
		case Expression:
			sql := value.(Expression).Build(db, params)
			values = append(values, sql)
		default:
			name := fmt.Sprintf("p%v", len(params))
			params[name] = value
			values = append(values, "{:"+name+"}")
		}
	}
	col := db.QuoteColumnName(e.col)
	if len(values) == 1 {
		if e.not {
			return col + "<>" + values[0]
		}
		return col + "=" + values[0]
	}
	in := "IN"
	if e.not {
		in = "NOT IN"
	}
	return fmt.Sprintf("%v %v (%v)", col, in, strings.Join(values, ", "))
}

// LikeExp represents a variant of LIKE expressions.
type LikeExp struct {
	or          bool
	left, right bool
	col         string
	values      []string
	escape      []string

	// Like stores the LIKE operator. It can be "LIKE", "NOT LIKE".
	// It may also be customized as something like "ILIKE".
	Like string
}

// Escape specifies how a LIKE expression should be escaped.
// Each string at position 2i represents a special character and the string at position 2i+1 is
// the corresponding escaped version.
func (e *LikeExp) Escape(chars ...string) *LikeExp {
	e.escape = chars
	return e
}

// Match specifies whether to do wildcard matching on the left and/or right of given strings.
func (e *LikeExp) Match(left, right bool) *LikeExp {
	e.left, e.right = left, right
	return e
}

// Build converts an expression into a SQL fragment.
func (e *LikeExp) Build(db *DB, params Params) string {
	if len(e.values) == 0 {
		return ""
	}

	if len(e.escape)%2 != 0 {
		panic("LikeExp.Escape must be a slice of even number of strings")
	}

	var parts []string
	col := db.QuoteColumnName(e.col)
	for _, value := range e.values {
		name := fmt.Sprintf("p%v", len(params))
		for i := 0; i < len(e.escape); i += 2 {
			value = strings.Replace(value, e.escape[i], e.escape[i+1], -1)
		}
		if e.left {
			value = "%" + value
		}
		if e.right {
			value += "%"
		}
		params[name] = value
		parts = append(parts, fmt.Sprintf("%v %v {:%v}", col, e.Like, name))
	}

	if e.or {
		return strings.Join(parts, " OR ")
	}
	return strings.Join(parts, " AND ")
}

// ExistsExp represents an EXISTS or NOT EXISTS expression.
type ExistsExp struct {
	exp Expression
	not bool
}

// Build converts an expression into a SQL fragment.
func (e *ExistsExp) Build(db *DB, params Params) string {
	sql := e.exp.Build(db, params)
	if sql == "" {
		if e.not {
			return ""
		}
		return "0=1"
	}
	if e.not {
		return "NOT EXISTS (" + sql + ")"
	}
	return "EXISTS (" + sql + ")"
}

// BetweenExp represents a BETWEEN or a NOT BETWEEN expression.
type BetweenExp struct {
	col      string
	from, to interface{}
	not      bool
}

// Build converts an expression into a SQL fragment.
func (e *BetweenExp) Build(db *DB, params Params) string {
	between := "BETWEEN"
	if e.not {
		between = "NOT BETWEEN"
	}
	name1 := fmt.Sprintf("p%v", len(params))
	name2 := fmt.Sprintf("p%v", len(params)+1)
	params[name1] = e.from
	params[name2] = e.to
	col := db.QuoteColumnName(e.col)
	return fmt.Sprintf("%v %v {:%v} AND {:%v}", col, between, name1, name2)
}
