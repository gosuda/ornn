package atlas

type DbType int8

const (
	DbTypeEmpty DbType = iota
	DbTypeMySQL
	DbTypeMaria
	DbTypePostgre
	DbTypeSQLite
	DbTypeTiDB
	DbTypeCockroachDB
)

var DbTypeStr map[DbType]string = map[DbType]string{
	DbTypeEmpty:       "empty",
	DbTypeMySQL:       "mysql",
	DbTypeMaria:       "mariadb",
	DbTypePostgre:     "postgres",
	DbTypeSQLite:      "sqlite",
	DbTypeTiDB:        "tidb",
	DbTypeCockroachDB: "cockroachdb",
}

var DbTypeStrReverse map[string]DbType = map[string]DbType{
	"empty":       DbTypeEmpty,
	"mysql":       DbTypeMySQL,
	"mariadb":     DbTypeMaria,
	"postgres":    DbTypePostgre,
	"sqlite":      DbTypeSQLite,
	"tidb":        DbTypeTiDB,
	"cockroachdb": DbTypeCockroachDB,
}
