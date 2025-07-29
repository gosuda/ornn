package parser_mysql

import (
	"fmt"
	"testing"

	"github.com/gosuda/ornn/atlas"
	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/db/db_mysql"
	tiparser "github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/stretchr/testify/require"
)

func TestParseMysqlSelect(t *testing.T) {
	db, err := db_mysql.New("127.0.0.1", "3306", "root", "1234", "db_name")
	require.NoError(t, err)

	atlas := atlas.New(atlas.DbTypeMaria, db)
	sc, err := atlas.InspectSchema()
	require.NoError(t, err)
	sch := config.Schema{
		DbType: atlas.DbType,
		Schema: sc,
	}

	myps := New(&sch).(*Parser)

	// parse
	// stmtNodes, _, err := tiparser.New().Parse("select seq, id from user where id = ? and seq = b limit 123 offset 456;", "", "")
	// stmtNodes, _, err := tiparser.New().Parse("SELECT Orders.OrderID, Customers.CustomerName, Orders.OrderDate FROM Orders INNER JOIN Customers ON Orders.CustomerID=Customers.CustomerID;", "", "")

	stmtNodes, _, err := tiparser.New().Parse("select * from user where id = ? and pw = abc", "", "")
	require.NoError(t, err)
	for _, stmtNode := range stmtNodes {
		selectStmt := stmtNode.(*ast.SelectStmt)
		// from
		tables := ParseJoinToTables(selectStmt.From.TableRefs)
		for _, tableExpr := range tables {
			tblName := ParseTableName(tableExpr)
			fmt.Println(tblName)
		}

		tableName := selectStmt.From.TableRefs.Left.(*ast.TableSource).Source.(*ast.TableName).Name.O
		table, _ := sch.Table(tableName)
		fmt.Println("table name :", tableName, table.Name)

		// select
		// select * 일 경우 schema 의 모든 필드 추출
		if selectStmt.Fields.Fields[0].WildCard != nil {
			tbl, ok := sch.Table(tableName)
			if ok == true {
				for _, col := range tbl.Columns {
					fmt.Printf("name : %s | db type : %s | golang Type : %s\n", col.Name, col.Type.Raw, myps.ConvType(col.Type))
				}
			}
		} else {
			// select * 외의 경우
		}

		// visit 하면서 재귀적으로 where 필드 추출
		fields := ParseWhereToFields(selectStmt.Where)
		for k, v := range fields {
			fmt.Printf("key : (%T) %v\n", k, k)
			fmt.Printf("value : (%T) %v\n", v, v)
		}
	}
}

func TestParseMysqlInsert(t *testing.T) {
	db, err := db_mysql.New("127.0.0.1", "3306", "root", "1234", "db_name")
	require.NoError(t, err)
	atlas := atlas.New(atlas.DbTypeMaria, db)
	sc, err := atlas.InspectSchema()
	require.NoError(t, err)

	myps := New(&config.Schema{
		DbType: atlas.DbType,
		Schema: sc,
	})

	parsedQuery, err := myps.Parse("insert into user VALUES(1, ?, ?, ?);")
	require.NoError(t, err)
	fmt.Println(parsedQuery)
}

func TestParseMysqlUpdate(t *testing.T) {
	db, err := db_mysql.New("127.0.0.1", "3306", "root", "1234", "db_name")
	require.NoError(t, err)
	atlas := atlas.New(atlas.DbTypeMaria, db)
	sc, err := atlas.InspectSchema()
	require.NoError(t, err)

	myps := New(&config.Schema{
		DbType: atlas.DbType,
		Schema: sc,
	})
	parsedQuery, err := myps.Parse("update user set seq = ?, id = ? where id = ?;")
	require.NoError(t, err)
	_ = parsedQuery
}
