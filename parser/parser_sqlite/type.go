package parser_sqlite

import (
	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/parser"
)

func (t *Parser) ConvType(colType *schema.ColumnType) (genType string) {
	parseType := parser.ParseType(colType.Raw)
	switch parseType.Type {
	case "bool", "boolean":
		genType = "bool"
		if parseType.Nullable {
			genType = "sql.NullBool"
		}
	case "int", "integer", "tinyint", "smallint", "mediumint":
		genType = "int32"
		if parseType.Nullable {
			genType = "sql.NullInt32"
		} else if parseType.Unsigned {
			genType = "uint32"
		}
	case "bigint":
		genType = "int64"
		if parseType.Nullable {
			genType = "sql.NullInt64"
		} else if parseType.Unsigned {
			genType = "uint64"
		}
	case "numeric", "real", "double", "float", "decimal":
		genType = "float64"
		if parseType.Nullable {
			genType = "sql.NullFloat64"
		}
	case "blob":
		genType = "[]byte"
	case "timestamp", "datetime", "date", "timestamp with timezone", "time with timezone", "time without timezone", "timestamp without timezone":
		genType = "Time"
		if parseType.Nullable {
			genType = "*Time"
		}
	case "varchar", "character", "varying character", "nchar", "native character", "nvarchar", "text", "clob", "time":
		genType = "string"
		if parseType.Nullable {
			genType = "sql.NullString"
		}
	default:
		genType = "any"
	}
	return genType
}
