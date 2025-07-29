package parser_mysql

import (
	"fmt"

	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/parser"
	sqlparser "github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

func New(sch *config.Schema) parser.Parser {
	return &Parser{
		sch: sch,
	}
}

type Parser struct {
	sch *config.Schema
}

func (p *Parser) Parse(sql string) (*parser.ParsedQuery, error) {
	sqlParser := sqlparser.New()
	stmtNodes, _, err := sqlParser.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}

	parseQuery := &parser.ParsedQuery{}
	parseQuery.Init(sql)

	for _, stmtNode := range stmtNodes {
		switch stmt := stmtNode.(type) {
		case *ast.SelectStmt:
			err = p.parseSelect(stmt, parseQuery)
		case *ast.InsertStmt:
			err = p.parseInsert(stmt, parseQuery)
		case *ast.UpdateStmt:
			err = p.parseUpdate(stmt, parseQuery)
		case *ast.DeleteStmt:
			err = p.parseDelete(stmt, parseQuery)
		default:
			err = fmt.Errorf("parser error | not support query statement %T", stmt)
		}
		if err != nil {
			return nil, err
		}
	}

	return parseQuery, nil
}

// parseQuery 를 stmt 를 이용해 초기화, p.sch 를 이용해 타입 설정
func (p *Parser) parseSelect(stmt *ast.SelectStmt, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeSelect

	// from
	tbl, err := p.parseFrom(stmt.From)
	if err != nil {
		return err
	}

	// select
	fields := stmt.Fields.Fields
	if len(fields) == 1 && fields[0].WildCard != nil { // select * 일 경우 schema 의 모든 필드 추출
		for _, col := range tbl.Columns {
			parsedQuery.Ret = append(parsedQuery.Ret, parser.NewField(col.Name, p.ConvType(col.Type)))
		}
	} else {
		for _, field := range stmt.Fields.Fields {
			switch fieldExpr := field.Expr.(type) {
			case *ast.ColumnNameExpr:
				colName := fieldExpr.Name.Name.O
				col, ok := tbl.Column(colName)
				if ok != true {
					parsedQuery.Ret = append(parsedQuery.Ret, parser.NewField(colName, "interface{}"))
				} else {
					parsedQuery.Ret = append(parsedQuery.Ret, parser.NewField(colName, p.ConvType(col.Type)))
				}
			}
		}
	}
	// where
	err = p.parseWhere(stmt.Where, tbl, parsedQuery)
	if err != nil {
		return err
	}

	return nil
}

// parseQuery 를 stmt 를 이용해 초기화, p.sch 를 이용해 타입 설정
func (p *Parser) parseInsert(stmt *ast.InsertStmt, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeInsert

	// from
	tbl, err := p.parseFrom(stmt.Table)
	if err != nil {
		return err
	}

	// insert fields
	if len(stmt.Lists) != 1 {
		panic("bulk query is invalid, use bulk options")
	}
	colNames := make([]string, len(tbl.Columns))
	if len(stmt.Columns) == 0 { // insert all fields
		for i, col := range tbl.Columns {
			colNames[i] = col.Name
		}
		if len(tbl.Columns) != len(stmt.Lists[0]) {
			panic("not same column and value count")
		}

		for i, list := range stmt.Lists[0] {
			if _, paramMarkerExpr, ok := ParseDriverValue(list); !ok {
				panic("need more programming")
			} else if paramMarkerExpr != nil {
				parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colNames[i], p.ConvType(tbl.Columns[i].Type)))
			}
		}
	} else { // insert specific fields
		if len(stmt.Columns) != len(stmt.Lists[0]) {
			panic("not same column and value count")
		}
		for i, list := range stmt.Lists[0] {
			if _, paramMarkerExpr, ok := ParseDriverValue(list); !ok {
				panic("need more programming")
			} else if paramMarkerExpr != nil {
				colName := stmt.Columns[i].Name.O
				col, ok := tbl.Column(colName)
				if ok != true {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colName, "interface{}"))
				} else {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colName, p.ConvType(col.Type)))
				}
			}
		}
	}

	if stmt.Select != nil {
		// TODO : insert select
		panic("need more programming")
	}

	if stmt.OnDuplicate != nil {
		// TODO : on duplicate
		panic("need more programming")
	}
	return nil
}

func (p *Parser) parseUpdate(stmt *ast.UpdateStmt, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeUpdate

	tbl, err := p.parseFrom(stmt.TableRefs)
	if err != nil {
		return err
	}

	// set
	for _, set := range stmt.List {
		colName := set.Column.Name.O
		col, ok := tbl.Column(colName)
		if ok != true {
			parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("set_"+colName, "interface{}"))
		} else {
			parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("set_"+colName, p.ConvType(col.Type)))
		}
	}

	// where
	err = p.parseWhere(stmt.Where, tbl, parsedQuery)
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) parseDelete(stmt *ast.DeleteStmt, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeDelete

	// from
	tbl, err := p.parseFrom(stmt.TableRefs)
	if err != nil {
		return err
	}

	// where
	err = p.parseWhere(stmt.Where, tbl, parsedQuery)
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) parseFrom(tableClause *ast.TableRefsClause) (tbl *schema.Table, err error) {
	tableSources := ParseJoinToTables(tableClause.TableRefs)
	if len(tableSources) != 1 {
		// TODO : delete join
		panic("need more programming")
	}
	tableName := ParseTableName(tableSources[0])
	tbl, ok := p.sch.Table(tableName)
	if ok != true {
		return nil, fmt.Errorf("parser error | not found table %s", tableName)
	}
	return tbl, nil
}

func (p *Parser) parseWhere(where ast.ExprNode, tbl *schema.Table, parsedQuery *parser.ParsedQuery) error {
	// where
	whereFields := ParseWhereToFields(where)
	for _, where := range whereFields {
		if where.right == nil || where.left == nil {
			continue
		}
		// left 의 column 을 인자로 추출
		if _, paramMarkerExpr, _ := ParseDriverValue(where.right); paramMarkerExpr != nil {
			if data, ok := where.left.(*ast.ColumnNameExpr); ok == true {
				colName := data.Name.Name.O
				col, ok := tbl.Column(colName)
				if ok != true {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("where_"+colName, "interface{}"))
				} else {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("where_"+colName, p.ConvType(col.Type)))
				}
			}
		}

		// right 의 column 을 인자로 추출
		if _, paramMarkerExpr, _ := ParseDriverValue(where.left); paramMarkerExpr != nil {
			if data, ok := where.right.(*ast.ColumnNameExpr); ok == true {
				colName := data.Name.Name.O
				col, ok := tbl.Column(colName)
				if ok != true {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField(colName, "interface{}"))
				} else {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField(colName, p.ConvType(col.Type)))
				}
			}
		}
	}
	return nil
}
