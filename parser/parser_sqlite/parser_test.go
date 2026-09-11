package parser_sqlite

import (
	"fmt"
	"testing"

	"ariga.io/atlas/sql/schema"
	"github.com/CovenantSQL/sqlparser"
	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/parser"
	"github.com/stretchr/testify/require"
)

func TestParseSqliteSelect(t *testing.T) {
	sql := "select a,b,c from test where a=1 and b=? and c=?"
	stmtNodes, err := sqlparser.Parse(sql)
	require.NoError(t, err)

	selectStmt := stmtNodes.(*sqlparser.Select)
	/*
		ret := ParseWhereToFields(selectStmt.Where.Expr)
		for _, v := range ret {
			fmt.Printf("ret : %d %T | %d %T\n", v.left, v.left, v.right, v.right)
			fmt.Println(string(v.right.(*sqlparser.SQLVal).Val), v.right.(*sqlparser.SQLVal).Type)
		}
	*/
	// select
	switch sel := selectStmt.SelectExprs[0].(type) {
	case *sqlparser.StarExpr: // select * 일 경우 schema 의 모든 인자 추출
		fmt.Println(sel)
	case *sqlparser.AliasedExpr:
		fmt.Printf("%T\n", sel.Expr)
		switch expr := sel.Expr.(type) {
		case *sqlparser.ColName:
			fmt.Println(expr)

		}
	default:
		panic("need more programming")
	}
}

func TestParseSqliteInsert(t *testing.T) {
	sql := "insert into test (a,b,c) values (?,?,?)"
	stmtNodes, err := sqlparser.Parse(sql)
	require.NoError(t, err)

	insertStmt := stmtNodes.(*sqlparser.Insert)
	fmt.Println(insertStmt.Table.Name.String())
	fmt.Println(insertStmt.OnDup)

}

func newSQLiteTestParser(t *testing.T) parser.Parser {
	t.Helper()

	sch := schema.New("")
	sch.AddTables(&schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "int"}}},
		},
	})
	return New(&config.Schema{Schema: sch})
}

func TestInsertSourceClassification(t *testing.T) {
	p := newSQLiteTestParser(t)

	t.Run("select source is unsupported", func(t *testing.T) {
		_, err := p.Parse("INSERT INTO users(id) SELECT id FROM users")
		require.ErrorContains(t, err, "unsupported INSERT source")
	})

	t.Run("multiple values rows are bulk", func(t *testing.T) {
		_, err := p.Parse("INSERT INTO users(id) VALUES (?), (?)")
		require.ErrorContains(t, err, "bulk query")
	})
}
