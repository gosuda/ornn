package parser_postgres

import (
	"errors"
	"fmt"

	"ariga.io/atlas/sql/schema"
	sqlparser "github.com/cockroachdb/cockroachdb-parser/pkg/sql/parser"
	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/sem/tree"
	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/parser"
)

func New(sch *config.Schema) parser.Parser {
	return &Parser{
		sch: sch,
	}
}

type Parser struct {
	sch *config.Schema
}

func parseSQL(sql string) (stmtNodes sqlparser.Statements, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parser backend panic: %v", r)
		}
	}()

	return sqlparser.Parse(sql)
}

func (p *Parser) Parse(sql string) (parsedQuery *parser.ParsedQuery, err error) {
	stmtNodes, err := parseSQL(sql)
	if err != nil {
		return nil, err
	} else if len(stmtNodes) != 1 {
		return nil, fmt.Errorf("parser error | only one statement is supported")
	}

	parsedQuery = &parser.ParsedQuery{}
	parsedQuery.Init(sql)
	switch stmt := stmtNodes[0].AST.(type) {
	case *tree.Select:
		err = p.parseSelect(stmt, parsedQuery)
	case *tree.Insert:
		err = p.parseInsert(stmt, parsedQuery)
	case *tree.Update:
		err = p.parseUpdate(stmt, parsedQuery)
	case *tree.Delete:
		err = p.parseDelete(stmt, parsedQuery)
	default:
		err = fmt.Errorf("parser error | not support query statement %T", stmt)
	}
	if err != nil {
		return nil, err
	}

	return parsedQuery, nil
}

func (p *Parser) parseSelect(stmt *tree.Select, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeSelect

	selectStmt, ok := stmt.Select.(*tree.SelectClause)
	if !ok {
		return fmt.Errorf("parser error | unsupported select clause type %T", stmt.Select)
	}
	// from
	if len(selectStmt.From.Tables) != 1 {
		return fmt.Errorf("parser error | only single table select is supported")
	}
	tbl, err := p.parseFrom(selectStmt.From.Tables[0])
	if err != nil {
		return err
	}

	// select
	if len(selectStmt.Exprs) > 0 {
		if _, ok := selectStmt.Exprs[0].Expr.(tree.UnqualifiedStar); ok == true {
			for _, col := range tbl.Columns {
				parsedQuery.Ret = append(parsedQuery.Ret, &parser.ParsedQueryField{
					Name:   col.Name,
					GoType: p.ConvType(col.Type.Raw),
				})
			}
		} else {
			for _, selectExpr := range selectStmt.Exprs {
				switch fieldExpr := selectExpr.Expr.(type) {
				case *tree.ColumnItem:
					colName := fieldExpr.ColumnName.String()
					col, ok := tbl.Column(colName)
					if ok != true {
						parsedQuery.Ret = append(parsedQuery.Ret, parser.NewField(colName, "any"))
					} else {
						parsedQuery.Ret = append(parsedQuery.Ret, parser.NewField(colName, p.ConvType(col.Type.Raw)))
					}
				}
			}
		}
	}
	// where
	if selectStmt.Where != nil {
		err = p.parseWhere(selectStmt.Where, tbl, parsedQuery)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parseInsert(stmt *tree.Insert, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeInsert

	// into
	tbl, err := p.parseFrom(stmt.Table)
	if err != nil {
		return err
	}

	// values
	valuesClause, ok := stmt.Rows.Select.(*tree.ValuesClause)
	if !ok {
		return fmt.Errorf("parser error | INSERT supports VALUES clause only")
	}
	rows := valuesClause.Rows
	if len(rows) != 1 {
		return errors.New("bulk query is invalid, use bulk options")
	}
	colNames := make([]string, len(tbl.Columns))
	if len(stmt.Columns) == 0 { // insert all fields
		for i, col := range tbl.Columns {
			colNames[i] = col.Name
		}
		if len(tbl.Columns) != len(rows[0]) {
			return fmt.Errorf("parser error | column count and values count mismatch")
		}
		for i, list := range rows[0] {
			if _, _, placeHolder, ok := ParseDriverValue(list); !ok {
				return fmt.Errorf("parser error | unsupported insert value type")
			} else if placeHolder != nil {
				parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colNames[i], p.ConvType(tbl.Columns[i].Type.Raw)))
			}
		}
	} else { // insert specific fields
		if len(stmt.Columns) != len(rows[0]) {
			return fmt.Errorf("parser error | column count and values count mismatch")
		}
		for i, list := range rows[0] {
			if _, _, paramMarkerExpr, ok := ParseDriverValue(list); !ok {
				return fmt.Errorf("parser error | unsupported insert value type")
			} else if paramMarkerExpr != nil {
				colName := stmt.Columns[i].String()
				col, ok := tbl.Column(colName)
				if ok != true {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colName, "any"))
				} else {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colName, p.ConvType(col.Type.Raw)))
				}
			}
		}
	}
	return nil
}

func (p *Parser) parseUpdate(stmt *tree.Update, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeUpdate
	tbl, err := p.parseFrom(stmt.Table)
	if err != nil {
		return err
	}

	// set
	for _, setExpr := range stmt.Exprs {
		if len(setExpr.Names) != 1 {
			return fmt.Errorf("parser error | multiple columns in one set clause are not supported")
		}
		colName := setExpr.Names[0].String()
		col, ok := tbl.Column(colName)
		if ok != true {
			parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colName, "any"))
		} else {
			parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("val_"+colName, p.ConvType(col.Type.Raw)))
		}
	}

	// where
	if stmt.Where != nil {
		err = p.parseWhere(stmt.Where, tbl, parsedQuery)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parseDelete(stmt *tree.Delete, parsedQuery *parser.ParsedQuery) error {
	parsedQuery.QueryType = parser.QueryTypeDelete

	// from
	tbl, err := p.parseFrom(stmt.Table)
	if err != nil {
		return err
	}

	// where
	if stmt.Where != nil {
		err = p.parseWhere(stmt.Where, tbl, parsedQuery)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parseFrom(tableClause tree.TableExpr) (tbl *schema.Table, err error) {

	var tableName string
	switch data := tableClause.(type) {
	case *tree.TableName:
		tableName = data.Table()
	case *tree.AliasedTableExpr:
		tableNameExpr, ok := data.Expr.(*tree.TableName)
		if !ok {
			return nil, fmt.Errorf("parser error | unsupported aliased table expression %T", data.Expr)
		}
		tableName = tableNameExpr.Table()
	default:
		return nil, fmt.Errorf("parser error | unsupported table clause %T", tableClause)
	}
	tbl, ok := p.sch.Table(tableName)
	if ok != true {
		return nil, fmt.Errorf("parser error | not found table %s", tableName)
	}
	return tbl, nil
}

func (p *Parser) parseWhere(where *tree.Where, tbl *schema.Table, parsedQuery *parser.ParsedQuery) (err error) {
	if where == nil || where.Expr == nil {
		return nil
	}
	whereFields, err := parseWhereToFields(where.Expr)
	if err != nil {
		return err
	}
	for _, where := range whereFields {
		// left 의 column 을 인자로 추출
		if placeHolder, _ := where.right.(*tree.Placeholder); placeHolder != nil {
			if data, ok := where.left.(*tree.UnresolvedName); ok == true {
				colName := data.Parts[0]
				col, ok := tbl.Column(colName)
				if ok != true {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("where_"+colName, "any"))
				} else {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("where_"+colName, p.ConvType(col.Type.Raw)))
				}
			}
		}
		// right 의 column 을 인자로 추출
		if placeHolder, _ := where.left.(*tree.Placeholder); placeHolder != nil {
			if data, ok := where.right.(*tree.UnresolvedName); ok == true {
				colName := data.Parts[0]
				col, ok := tbl.Column(colName)
				if ok != true {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("where_"+colName, "any"))
				} else {
					parsedQuery.Arg = append(parsedQuery.Arg, parser.NewField("where_"+colName, p.ConvType(col.Type.Raw)))
				}
			}
		}
	}

	return nil
}
