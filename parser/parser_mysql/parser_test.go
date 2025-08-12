package parser_mysql

import (
	"testing"

	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/parser"
	"github.com/stretchr/testify/require"

	"github.com/gosuda/ornn/config"
)

func newTestSchema(t *testing.T) *config.Schema {
	t.Helper()

	users := &schema.Table{Name: "users"}
	users.Columns = []*schema.Column{
		{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
		{Name: "name", Type: &schema.ColumnType{Type: &schema.StringType{T: "varchar"}}},
		{Name: "age", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
	}
	orders := &schema.Table{Name: "orders"}
	orders.Columns = []*schema.Column{
		{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
		{Name: "user_id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
		{Name: "amount", Type: &schema.ColumnType{Type: &schema.DecimalType{T: "decimal"}}},
	}

	s := &config.Schema{}
	s.Schema = &schema.Schema{}
	s.AddTables(users, orders)
	return s
}

func retNames(pq *parser.ParsedQuery) []string {
	out := make([]string, len(pq.Ret))
	for i, f := range pq.Ret {
		out[i] = f.Name
	}
	return out
}

func argNames(pq *parser.ParsedQuery) []string {
	out := make([]string, len(pq.Arg))
	for i, f := range pq.Arg {
		out[i] = f.Name
	}
	return out
}

func mustParse(t *testing.T, p parser.Parser, sql string) *parser.ParsedQuery {
	t.Helper()
	pq, err := p.Parse(sql)
	require.NoError(t, err)
	return pq
}

func newParser(t *testing.T) parser.Parser {
	return New(newTestSchema(t))
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func TestSelectStar(t *testing.T) {
	p := newParser(t)
	pq := mustParse(t, p, `SELECT * FROM users WHERE id = ?`)

	require.Equal(t, parser.QueryTypeSelect, pq.QueryType)

	gotRet := retNames(pq)
	wantAny := []string{"id", "name", "age"}
	for _, w := range wantAny {
		require.Containsf(t, gotRet, w, "SELECT * missing column %q, got %v", w, gotRet)
	}

	gotArg := argNames(pq)
	require.Equal(t, []string{"where_id"}, gotArg)
}

func TestSelectColumnsAndWhereVariants(t *testing.T) {
	p := newParser(t)
	sql := `
SELECT u.id, u.name
FROM users AS u
WHERE (u.id = ? OR ? = u.age)
  AND u.age BETWEEN ? AND ?
  AND u.name LIKE ?
  AND u.id IN (?, 2, 3, ?, func())
`
	pq := mustParse(t, p, sql)

	gotRet := retNames(pq)
	require.Len(t, gotRet, 2)
	require.NotEmpty(t, gotRet[0])
	require.NotEmpty(t, gotRet[1])

	gotArg := argNames(pq)
	expect := []string{
		"where_u.id",
		"where_u.age",
		"where_u.age_from",
		"where_u.age_to",
		"where_u.name_like",
		"where_u.id_in_0",
		"where_u.id_in_3",
	}
	for _, want := range expect {
		require.Containsf(t, gotArg, want, "missing arg %q; got %v", want, gotArg)
	}
}

func TestInsertValuesSingleRow(t *testing.T) {
	p := newParser(t)
	sql := `INSERT INTO users(id, name, age) VALUES (?, 'alice', ?)`
	pq := mustParse(t, p, sql)

	require.Equal(t, parser.QueryTypeInsert, pq.QueryType)
	got := argNames(pq)
	expect := []string{"val_id", "val_age"}
	require.Equal(t, expect, got)
}

func TestInsertValuesBulkRows(t *testing.T) {
	p := newParser(t)
	sql := `
INSERT INTO users(id, name, age)
VALUES (?, ?, ?),
       (2, 'bob', ?),
       (?, 'carol', 30)
`
	pq := mustParse(t, p, sql)

	got := argNames(pq)
	expect := []string{
		"val_id", "val_name", "val_age", // row1
		"val_age_2", // row2
		"val_id_2",  // row3
	}

	require.Equal(t, expect, got, "bulk insert arg names mismatch")
}

func TestInsertSelect(t *testing.T) {
	p := newParser(t)
	sql := `
INSERT INTO orders(id, user_id, amount)
SELECT u.id, u.id, 100
FROM users u
WHERE u.age >= ?
`
	pq := mustParse(t, p, sql)

	require.Equal(t, parser.QueryTypeInsert, pq.QueryType)
	got := argNames(pq)
	require.Equal(t, []string{"where_u.age"}, got)
}

func TestInsertOnDuplicate(t *testing.T) {
	p := newParser(t)
	sql := `
INSERT INTO users(id, name, age) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
  name = ?,
  age  = age + 1
`
	pq := mustParse(t, p, sql)

	got := argNames(pq)
	expect := []string{"val_id", "val_name", "val_age", "dup_name"}
	require.Equal(t, expect, got)
}

func TestUpdateSetWithParams(t *testing.T) {
	p := newParser(t)
	sql := `UPDATE users SET name = ?, age = age + 1 WHERE id = ?`
	pq := mustParse(t, p, sql)

	require.Equal(t, parser.QueryTypeUpdate, pq.QueryType)
	got := argNames(pq)
	expect := []string{"set_name", "where_id"}
	require.Equal(t, expect, got)
}

func TestDeleteWhere(t *testing.T) {
	p := newParser(t)
	sql := `DELETE FROM users WHERE name LIKE ? AND age BETWEEN ? AND ?`
	pq := mustParse(t, p, sql)

	require.Equal(t, parser.QueryTypeDelete, pq.QueryType)
	got := argNames(pq)
	expectSet := map[string]bool{
		"where_name_like": true,
		"where_age_from":  true,
		"where_age_to":    true,
	}
	for _, a := range got {
		delete(expectSet, a)
	}
	require.Emptyf(t, expectSet, "missing args: %v; got %v", expectSet, got)
}

func TestJoinAliasResolution(t *testing.T) {
	p := newParser(t)
	sql := `
SELECT u.id, o.amount
FROM users AS u
JOIN orders AS o ON o.user_id = u.id
WHERE u.id = ? AND o.amount > ?
`
	pq := mustParse(t, p, sql)

	ret := retNames(pq)
	require.Len(t, ret, 2)

	args := argNames(pq)
	require.Contains(t, args, "where_u.id")
	require.Contains(t, args, "where_o.amount")
}
