package parser_mysql

import (
	"fmt"

	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/parser"
	sqlparser "github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/test_driver"
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

	pq := &parser.ParsedQuery{}
	pq.Init(sql)

	for _, stmtNode := range stmtNodes {
		switch stmt := stmtNode.(type) {
		case *ast.SelectStmt:
			err = p.parseSelect(stmt, pq)
		case *ast.InsertStmt:
			err = p.parseInsert(stmt, pq)
		case *ast.UpdateStmt:
			err = p.parseUpdate(stmt, pq)
		case *ast.DeleteStmt:
			err = p.parseDelete(stmt, pq)
		default:
			err = fmt.Errorf("parser error | unsupported statement %T", stmt)
		}
		if err != nil {
			return nil, err
		}
	}
	return pq, nil
}

func (p *Parser) parseSelect(stmt *ast.SelectStmt, pq *parser.ParsedQuery) error {
	pq.QueryType = parser.QueryTypeSelect

	// FROM
	tbl, err := p.parseFrom(stmt.From)
	if err != nil {
		return err
	}

	// SELECT list
	p.collectSelectFields(tbl, stmt.Fields, pq)

	// WHERE
	return p.parseWhere(stmt.Where, tbl, pq)
}

func (p *Parser) collectSelectFields(tbl *schema.Table, fields *ast.FieldList, pq *parser.ParsedQuery) {
	if fields == nil || len(fields.Fields) == 0 {
		return
	}

	// SELECT *
	if len(fields.Fields) == 1 && fields.Fields[0].WildCard != nil {
		for _, col := range tbl.Columns {
			pq.Ret = append(pq.Ret, parser.NewField(col.Name, p.ConvType(col.Type)))
		}
		return
	}

	// Explicit fields
	for _, f := range fields.Fields {
		switch fe := f.Expr.(type) {
		case *ast.ColumnNameExpr:
			name, typ := p.resolveColumn(tbl, fe)
			if f.AsName.O != "" {
				name = f.AsName.O
			}
			pq.Ret = append(pq.Ret, parser.NewField(name, typ))
		default:
			name := f.AsName.O
			if name == "" {
				name = "expr"
			}
			pq.Ret = append(pq.Ret, parser.NewField(name, "any"))
		}
	}
}

func (p *Parser) parseInsert(stmt *ast.InsertStmt, pq *parser.ParsedQuery) error {
	pq.QueryType = parser.QueryTypeInsert

	// FROM
	tbl, err := p.parseFrom(stmt.Table)
	if err != nil {
		return err
	}

	// VALUES와 SELECT를 동시에 쓰는 건 비지원
	if stmt.Select != nil && len(stmt.Lists) > 0 {
		return fmt.Errorf("parser error | INSERT ... SELECT with VALUES together is not supported")
	}

	// ===== CASE A: INSERT ... VALUES (...), (...), ... =====
	if len(stmt.Lists) > 0 {
		// 대상 컬럼 집합
		targetCols, err := p.targetColumns(tbl, stmt.Columns)
		if err != nil {
			return err
		}

		// 행별 길이 검증
		for rowIdx, values := range stmt.Lists {
			if len(values) != len(targetCols) {
				return fmt.Errorf("parser error | columns (%d) != values (%d) at row %d",
					len(targetCols), len(values), rowIdx+1)
			}
		}

		// 중복 이름 관리: "val_<col>"가 반복될 때만 _2, _3 ...
		countSeen := make(map[string]int)

		for _, row := range stmt.Lists {
			for colIdx, v := range row {
				if !isParam(v) {
					continue
				}
				colName := targetCols[colIdx]
				if p.IsReservedKeyword(colName) {
					return fmt.Errorf("parser error | reserved keyword used as identifier: %s", colName)
				}
				base := "val_" + colName
				countSeen[base]++
				argName := base
				if countSeen[base] > 1 {
					argName = fmt.Sprintf("%s_%d", base, countSeen[base])
				}

				if col, ok := tbl.Column(colName); ok {
					pq.Arg = append(pq.Arg, parser.NewField(argName, p.ConvType(col.Type)))
				} else {
					pq.Arg = append(pq.Arg, parser.NewField(argName, "any"))
				}
			}
		}
	}

	// ===== CASE B: INSERT ... SELECT =====
	if stmt.Select != nil {
		targetCols, err := p.targetColumns(tbl, stmt.Columns)
		if err != nil {
			return err
		}

		subPQ := &parser.ParsedQuery{}
		subPQ.Init("")
		sel, ok := stmt.Select.(*ast.SelectStmt)
		if !ok {
			return fmt.Errorf("parser error | unsupported SELECT node type %T", stmt.Select)
		}
		if err := p.parseSelect(sel, subPQ); err != nil {
			return fmt.Errorf("parser error | parse inner SELECT: %w", err)
		}
		if len(targetCols) != len(subPQ.Ret) {
			return fmt.Errorf("parser error | INSERT target columns (%d) and SELECT fields (%d) mismatch",
				len(targetCols), len(subPQ.Ret))
		}
		// 내부 WHERE 등에서 나온 ?를 흡수
		pq.Arg = append(pq.Arg, subPQ.Arg...)
	}

	// ===== ON DUPLICATE KEY UPDATE =====
	if stmt.OnDuplicate != nil {
		for _, asg := range stmt.OnDuplicate {
			colName := asg.Column.Name.O
			if p.IsReservedKeyword(colName) {
				return fmt.Errorf("parser error | reserved keyword used as identifier: %s", colName)
			}
			if !hasParamMarker(asg.Expr) {
				continue
			}
			if col, ok := tbl.Column(colName); ok {
				pq.Arg = append(pq.Arg, parser.NewField("dup_"+colName, p.ConvType(col.Type)))
			} else {
				pq.Arg = append(pq.Arg, parser.NewField("dup_"+colName, "any"))
			}
		}
	}

	// 둘 다 없으면 에러
	if len(stmt.Lists) == 0 && stmt.Select == nil {
		return fmt.Errorf("parser error | INSERT has neither VALUES nor SELECT")
	}
	return nil
}

func (p *Parser) targetColumns(tbl *schema.Table, cols []*ast.ColumnName) ([]string, error) {
	if len(cols) == 0 {
		out := make([]string, len(tbl.Columns))
		for i, c := range tbl.Columns {
			out[i] = c.Name
		}
		return out, nil
	}
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Name.O
	}
	return out, nil
}

func (p *Parser) parseUpdate(stmt *ast.UpdateStmt, pq *parser.ParsedQuery) error {
	pq.QueryType = parser.QueryTypeUpdate

	tbl, err := p.parseFrom(stmt.TableRefs)
	if err != nil {
		return err
	}

	// SET (값이 ? 일 때만 Arg 추가)
	for _, set := range stmt.List {
		if !isParam(set.Expr) {
			continue
		}
		colName := set.Column.Name.O
		if p.IsReservedKeyword(colName) {
			return fmt.Errorf("parser error | reserved keyword used as identifier: %s", colName)
		}
		if col, ok := tbl.Column(colName); ok {
			p.addArg(pq, "set_", colName, p.ConvType(col.Type))
		} else {
			p.addArg(pq, "set_", colName, "any")
		}
	}

	// WHERE
	return p.parseWhere(stmt.Where, tbl, pq)
}

func (p *Parser) parseDelete(stmt *ast.DeleteStmt, pq *parser.ParsedQuery) error {
	pq.QueryType = parser.QueryTypeDelete

	tbl, err := p.parseFrom(stmt.TableRefs)
	if err != nil {
		return err
	}
	return p.parseWhere(stmt.Where, tbl, pq)
}
func (p *Parser) parseFrom(tableClause *ast.TableRefsClause) (*schema.Table, error) {
	if tableClause == nil || tableClause.TableRefs == nil {
		return nil, fmt.Errorf("parser error | missing FROM clause")
	}
	tableSources := ParseJoinToTables(tableClause.TableRefs)

	// 단일 테이블
	if len(tableSources) == 1 {
		tableName := ParseTableName(tableSources[0])
		tbl, ok := p.sch.Table(tableName)
		if !ok {
			return nil, fmt.Errorf("parser error | not found table %s", tableName)
		}
		return tbl, nil
	}

	// JOIN → 가상 테이블 구성 (alias.col > table.col > col)
	joined := &schema.Table{Name: "__joined__"}
	exists := map[string]bool{}

	for _, ts := range tableSources {
		tname := ParseTableName(ts)
		baseTbl, ok := p.sch.Table(tname)
		if !ok {
			return nil, fmt.Errorf("parser error | not found table %s", tname)
		}

		var alias string
		if ts.AsName.O != "" {
			alias = ts.AsName.O
		}

		for _, c := range baseTbl.Columns {
			cands := []string{}
			if alias != "" {
				cands = append(cands, fmt.Sprintf("%s.%s", alias, c.Name))
			}
			cands = append(cands, fmt.Sprintf("%s.%s", tname, c.Name))
			cands = append(cands, c.Name)

			var chosen string
			for _, cand := range cands {
				if !exists[cand] {
					chosen = cand
					break
				}
			}
			if chosen == "" {
				if alias != "" {
					chosen = fmt.Sprintf("%s.%s", alias, c.Name)
				} else {
					chosen = fmt.Sprintf("%s.%s", tname, c.Name)
				}
			}
			exists[chosen] = true
			joined.Columns = append(joined.Columns, &schema.Column{Name: chosen, Type: c.Type})
		}
	}
	return joined, nil
}

func ParseTableName(table *ast.TableSource) string {
	switch data := table.Source.(type) {
	case *ast.TableName:
		return data.Name.String()
	case *ast.SelectStmt:
		return data.Text()
	default:
		panic("parser error | not support table type")
	}
}

// 왼/오 재귀로 JOIN 내 테이블 소스 수집
func ParseJoinToTables(join *ast.Join) []*ast.TableSource {
	if join == nil {
		return nil
	}
	nodes := make([]*ast.TableSource, 0, 8)
	if join.Left != nil {
		switch data := join.Left.(type) {
		case *ast.Join:
			nodes = append(nodes, ParseJoinToTables(data)...)
		case *ast.TableSource:
			nodes = append(nodes, data)
		default:
			panic("parser error | not support join-left type")
		}
	}
	if join.Right != nil {
		switch data := join.Right.(type) {
		case *ast.Join:
			nodes = append(nodes, ParseJoinToTables(data)...)
		case *ast.TableSource:
			nodes = append(nodes, data)
		default:
			panic("parser error | not support join-right type")
		}
	}
	return nodes
}
func (p *Parser) parseWhere(where ast.ExprNode, tbl *schema.Table, pq *parser.ParsedQuery) error {
	if where == nil {
		return nil
	}

	var walk func(e ast.ExprNode) error
	walk = func(e ast.ExprNode) error {
		switch n := e.(type) {
		case *ast.ParenthesesExpr:
			return walk(n.Expr)

		case *ast.BinaryOperationExpr:
			if err := walk(n.L); err != nil {
				return err
			}
			if err := walk(n.R); err != nil {
				return err
			}
			// L = ?, R = col
			if isParam(n.L) {
				if col, ok := n.R.(*ast.ColumnNameExpr); ok {
					name, typ := p.resolveColumn(tbl, col)
					pq.Arg = append(pq.Arg, parser.NewField("where_"+name, typ))
				}
			}
			// L = col, R = ?
			if isParam(n.R) {
				if col, ok := n.L.(*ast.ColumnNameExpr); ok {
					name, typ := p.resolveColumn(tbl, col)
					pq.Arg = append(pq.Arg, parser.NewField("where_"+name, typ))
				}
			}
			return nil

		case *ast.BetweenExpr:
			// col BETWEEN low AND high
			if err := walk(n.Expr); err != nil {
				return err
			}
			if err := walk(n.Left); err != nil {
				return err
			}
			if err := walk(n.Right); err != nil {
				return err
			}
			if col, ok := n.Expr.(*ast.ColumnNameExpr); ok {
				name, typ := p.resolveColumn(tbl, col)
				if isParam(n.Left) {
					pq.Arg = append(pq.Arg, parser.NewField("where_"+name+"_from", typ))
				}
				if isParam(n.Right) {
					pq.Arg = append(pq.Arg, parser.NewField("where_"+name+"_to", typ))
				}
			}
			return nil

		case *ast.PatternInExpr:
			// col IN (?, ?, 1, func(), ...)
			if err := walk(n.Expr); err != nil {
				return err
			}
			if n.List != nil {
				if col, ok := n.Expr.(*ast.ColumnNameExpr); ok {
					name, typ := p.resolveColumn(tbl, col)
					for i, it := range n.List {
						if err := walk(it); err != nil {
							return err
						}
						if isParam(it) {
							pq.Arg = append(pq.Arg, parser.NewField(fmt.Sprintf("where_%s_in_%d", name, i), typ))
						}
					}
				} else {
					for _, it := range n.List {
						if err := walk(it); err != nil {
							return err
						}
					}
				}
			}
			return nil

		case *ast.PatternLikeExpr:
			// col LIKE ?
			if err := walk(n.Expr); err != nil {
				return err
			}
			if err := walk(n.Pattern); err != nil {
				return err
			}
			if col, ok := n.Expr.(*ast.ColumnNameExpr); ok && isParam(n.Pattern) {
				name, typ := p.resolveColumn(tbl, col)
				pq.Arg = append(pq.Arg, parser.NewField("where_"+name+"_like", typ))
			}
			return nil

		case *ast.UnaryOperationExpr:
			return walk(n.V)

		case *ast.FuncCallExpr:
			for _, a := range n.Args {
				if err := walk(a); err != nil {
					return err
				}
			}
			return nil

		default:
			return nil
		}
	}
	return walk(where)
}
func (p *Parser) resolveColumn(tbl *schema.Table, c *ast.ColumnNameExpr) (name string, typ string) {
	col := c.Name.Name.O
	tblName := c.Name.Table.O
	display := col
	if tblName != "" {
		display = fmt.Sprintf("%s.%s", tblName, col)
	}

	// 탐색 후보
	candidates := []string{}
	if tblName != "" {
		candidates = append(candidates, fmt.Sprintf("%s.%s", tblName, col))
	}
	candidates = append(candidates, col)

	for _, cand := range candidates {
		if real, ok := tbl.Column(cand); ok {
			if tblName == "" {
				return cand, p.ConvType(real.Type)
			}
			return display, p.ConvType(real.Type)
		}
	}
	return display, "any"
}

func parseDriverValue(node ast.ExprNode) (*test_driver.ValueExpr, *test_driver.ParamMarkerExpr, bool) {
	switch data := node.(type) {
	case *test_driver.ValueExpr:
		return data, nil, true
	case *test_driver.ParamMarkerExpr:
		return nil, data, true
	default:
		return nil, nil, false
	}
}

func isParam(e ast.ExprNode) bool {
	_, pm, _ := parseDriverValue(e)
	return pm != nil
}

func hasParamMarker(e ast.ExprNode) bool {
	switch n := e.(type) {
	case *ast.ParenthesesExpr:
		return hasParamMarker(n.Expr)
	case *ast.UnaryOperationExpr:
		return hasParamMarker(n.V)
	case *ast.BinaryOperationExpr:
		return hasParamMarker(n.L) || hasParamMarker(n.R)
	case *ast.BetweenExpr:
		return hasParamMarker(n.Expr) || hasParamMarker(n.Left) || hasParamMarker(n.Right)
	case *ast.PatternInExpr:
		if hasParamMarker(n.Expr) {
			return true
		}
		for _, it := range n.List {
			if hasParamMarker(it) {
				return true
			}
		}
		return false
	case *ast.PatternLikeExpr:
		return hasParamMarker(n.Expr) || hasParamMarker(n.Pattern)
	case *ast.FuncCallExpr:
		for _, a := range n.Args {
			if hasParamMarker(a) {
				return true
			}
		}
		return false
	default:
		return isParam(e)
	}
}

func (p *Parser) addArg(pq *parser.ParsedQuery, prefix, name, typ string) {
	pq.Arg = append(pq.Arg, parser.NewField(prefix+name, typ))
}
