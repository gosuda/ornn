package parser_postgres

import (
	"fmt"

	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/sem/tree"
)

func ParseDriverValue(node tree.Expr) (*tree.NumVal, *tree.StrVal, *tree.Placeholder, bool) {
	switch data := node.(type) {
	case *tree.NumVal:
		return data, nil, nil, true
	case *tree.StrVal:
		return nil, data, nil, true
	case *tree.Placeholder:
		return nil, nil, data, true
	default:
		return nil, nil, nil, false
	}
}

type binaryExpr struct {
	left  tree.Expr
	right tree.Expr
	op    string
}

func ParseWhereToFields(whereExpr tree.Expr) []*binaryExpr {
	fields, _ := parseWhereToFields(whereExpr)
	return fields
}

func parseWhereToFields(whereExpr tree.Expr) ([]*binaryExpr, error) {
	if whereExpr == nil {
		return nil, nil
	}
	fields := make([]*binaryExpr, 0, 100)

	switch data := whereExpr.(type) {
	case *tree.ComparisonExpr:
		fields = append(fields, &binaryExpr{
			left:  data.Left,
			right: data.Right,
			op:    data.Operator.String(),
		})
	case *tree.AndExpr:
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
	case *tree.OrExpr:
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
	case *tree.ParenExpr:
		return parseWhereToFields(data.Expr)
	case *tree.NotExpr:
		return parseWhereToFields(data.Expr)
	case *tree.NumVal:
		// do nothing
	case *tree.StrVal:
		// do nothing
	case *tree.Placeholder:
		// do nothing
	case *tree.Subquery:
		// do nothing
	default:
		return nil, fmt.Errorf("parser error | unsupported where type %T", whereExpr)
	}
	return fields, nil
}
