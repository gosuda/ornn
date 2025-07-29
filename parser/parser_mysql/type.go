package parser_mysql

import (
	"regexp"

	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/parser"
)

func (p *Parser) ConvType(colType *schema.ColumnType) (genType string) {
	parseType := parser.ParseType(colType.Raw)
	parseType.Nullable = colType.Null
	switch parseType.Type {
	case "bit":
		switch {
		case parseType.Prec == 1:
			genType = "bool"
		case parseType.Prec <= 8:
			genType = "uint8"
		case parseType.Prec <= 16:
			genType = "uint16"
		case parseType.Prec <= 32:
			genType = "uint32"
		default:
			genType = "uint64"
		}
		if parseType.Nullable {
			genType = "*" + genType
		}

	case "bool", "boolean":
		genType = "bool"
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext":
		genType = "string"
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "tinyint":
		switch {
		case parseType.Prec == 1:
			genType = "bool"
		default:
			genType = "int8"
			if parseType.Unsigned {
				genType = "uint8"
			}
		}
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "smallint", "year":
		genType = "int16"
		if parseType.Unsigned {
			genType = "uint16"
		}
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "mediumint", "int", "integer":
		genType = "int32"
		if parseType.Unsigned {
			genType = "uint32"
		}
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "bigint":
		genType = "int64"
		if parseType.Unsigned {
			genType = "uint64"
		}
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "float":
		genType = "float32"
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "decimal", "double":
		genType = "float64"
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "binary", "blob", "longblob", "mediumblob", "tinyblob", "varbinary":
		genType = "[]byte"
	case "json":
		genType = "json.RawMessage"
	case "timestamp", "datetime", "date":
		genType = "time.Time"
		if parseType.Nullable {
			genType = "*" + genType
		}
	case "time":
		genType = "string"
		if parseType.Nullable {
			genType = "*" + genType
		}
	default:
		genType = "interface{}"
	}
	if regexp.MustCompile(`(?i)^set\([^)]*\)$`).MatchString(parseType.Type) {
		genType = "[]byte"
	}

	return genType
}
