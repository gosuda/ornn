package parser_sqlite

import (
	"fmt"

	"github.com/CovenantSQL/sqlparser"
)

type binaryExpr struct {
	left  sqlparser.Expr
	right sqlparser.Expr
	op    string
}

func ParseWhereToFields(whereExpr sqlparser.Expr) []*binaryExpr {
	fields, _ := parseWhereToFields(whereExpr)
	return fields
}

func parseWhereToFields(whereExpr sqlparser.Expr) ([]*binaryExpr, error) {
	if whereExpr == nil {
		return nil, nil
	}
	fields := make([]*binaryExpr, 0, 100)

	switch data := whereExpr.(type) {
	case *sqlparser.BinaryExpr:
		fields = append(fields, &binaryExpr{
			left:  data.Left,
			right: data.Right,
			op:    data.Operator,
		})
	case *sqlparser.AndExpr:
		left, err := parseWhereToFields(data.Left)
		if err != nil {
			return nil, err
		}
		right, err := parseWhereToFields(data.Right)
		if err != nil {
			return nil, err
		}
		fields = append(fields, left...)
		fields = append(fields, right...)
	case *sqlparser.OrExpr:
		left, err := parseWhereToFields(data.Left)
		if err != nil {
			return nil, err
		}
		right, err := parseWhereToFields(data.Right)
		if err != nil {
			return nil, err
		}
		fields = append(fields, left...)
		fields = append(fields, right...)
	case *sqlparser.ComparisonExpr:
		fields = append(fields, &binaryExpr{
			left:  data.Left,
			right: data.Right,
		})
	case *sqlparser.ParenExpr:
		return parseWhereToFields(data.Expr)
	case *sqlparser.NotExpr:
		return parseWhereToFields(data.Expr)
	case *sqlparser.ExistsExpr:
		return parseWhereToFields(data.Subquery)
	case *sqlparser.SQLVal:
		// do nothing
	case *sqlparser.NullVal:
		// do nothing
	case *sqlparser.ColName:
		// do nothing
	case *sqlparser.Subquery:
		// do nothing
	case *sqlparser.ListArg:
		// do nothing
	default:
		return nil, fmt.Errorf("parser error | unsupported where type %T", whereExpr)
	}
	return fields, nil
}

func ParseDriverValue(node sqlparser.Expr) (*sqlparser.ColName, *sqlparser.SQLVal, bool) {
	switch data := node.(type) {
	case *sqlparser.SQLVal:
		return nil, data, true
	default:
		return nil, nil, false
	}
}
