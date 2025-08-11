// Example usage:
//
//	package main
//
//	import (
//		"os"
//
//	    ornn "https://github.com/gosuda/ornn/cmd/ornn"
//	)
//
//	func main() {
//		if err := ornn.Run(os.Args[1:]); err != nil {
//			os.Exit(1)
//		}
//	}
package main

import (
	"os"

	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/atlas"
	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/db"
	"github.com/gosuda/ornn/db/db_mysql"
	"github.com/gosuda/ornn/db/db_postgres"
	"github.com/gosuda/ornn/db/db_sqlite"
	"github.com/gosuda/ornn/gen"
	"github.com/gosuda/ornn/parser"
	"github.com/gosuda/ornn/parser/parser_mysql"
	"github.com/gosuda/ornn/parser/parser_postgres"
	"github.com/gosuda/ornn/parser/parser_sqlite"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Run(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

var (
	rootCmd = &cobra.Command{
		Use:   "ornn",
		Short: "ornn is a code generator for golang",
		Long:  "ornn is a code generator for golang db access",
		Run:   rootRun,
	}

	loadExistSchemaFile bool // 기존 스키마 파일에서 로딩, 스키마 파일대로 db migrate
	loadExistConfigFile bool // 기존 설정 파일에서 로딩
	configFilePath      string
)

func init() {
	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&configFilePath, "config", "c", "config.toml", "Path to config file")
	fs.BoolVar(&loadExistSchemaFile, "file_schema_load", false, "load schema from existing file and migrate database")
	fs.BoolVar(&loadExistConfigFile, "file_config_load", false, "load config from existing file")
}

func main() {
	if err := Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) {
	cfg, err := loadConfig()
	if err != nil {
		log.Panic().Err(err).Msg("Failed to load config")
	}
	atlasDbType := atlas.DbTypeStrReverse[cfg.DB.Type]

	// 1. connect db
	var conn *db.Conn
	switch atlasDbType {
	case atlas.DbTypeMySQL, atlas.DbTypeMaria, atlas.DbTypeTiDB:
		conn, err = db_mysql.New(db_mysql.Dsn(cfg.DB.User, cfg.DB.Password, cfg.DB.Addr, cfg.DB.Port, cfg.DB.Name), cfg.DB.Name)
	case atlas.DbTypePostgre, atlas.DbTypeCockroachDB:
		conn, err = db_postgres.New(db_postgres.Dsn(cfg.DB.User, cfg.DB.Password, cfg.DB.Addr, cfg.DB.Port, cfg.DB.Name), cfg.DB.Name)
	case atlas.DbTypeSQLite:
		conn, err = db_sqlite.New(cfg.DB.Path)
	default:
		log.Panic().Msgf("invalid db type: %s", cfg.DB.Type)
	}
	if err != nil {
		log.Panic().Err(err).Msg("db connect error")
	}

	// 2. init schema from atl
	var sch *schema.Schema
	atl := atlas.New(atlasDbType, conn)
	if loadExistSchemaFile { // load from existing schema file
		if sch, err = atl.Load(cfg.Gen.SchemaPath); err != nil {
			log.Panic().Err(err).Msg("schema load error")
		}
		// migrate db from file
		if err = atl.MigrateSchema(sch); err != nil {
			log.Panic().Err(err).Msg("atlas migrate error")
		}
		// inspect schema fron migrated db
		if sch, err = atl.InspectSchema(); err != nil {
			log.Panic().Err(err).Msg("atlas inspect error")
		}
	} else {
		if sch, err = atl.InspectSchema(); err != nil {
			log.Panic().Err(err).Msg("atlas inspect error")
		}
		if err = atl.Save(cfg.Gen.SchemaPath, sch); err != nil {
			log.Panic().Err(err).Msg("schema save error")
		}
	}

	// 3. set config
	var conf = &config.Config{}
	if loadExistConfigFile { // load from existing config file
		if err = conf.Load(cfg.Gen.ConfigPath); err != nil { // load
			log.Panic().Err(err).Msg("config load error")
		}
		if err = conf.Init(atlasDbType, sch, cfg.Gen.GenPath, cfg.Gen.FileName, cfg.Gen.PackageName, cfg.Gen.ClassName); err != nil { // init
			log.Panic().Err(err).Msg("config init error")
		}
	} else {
		if err = conf.Init(atlasDbType, sch, cfg.Gen.GenPath, cfg.Gen.FileName, cfg.Gen.PackageName, cfg.Gen.ClassName); err != nil { // init
			log.Panic().Err(err).Msg("config init error")
		}
		if err = conf.Save(cfg.Gen.ConfigPath); err != nil { // save
			log.Panic().Err(err).Msg("config save error")
		}
	}

	// 4. set parser
	var psr parser.Parser
	switch atlasDbType {
	case atlas.DbTypeMySQL, atlas.DbTypeMaria, atlas.DbTypeTiDB:
		psr = parser_mysql.New(&conf.Schema)
	case atlas.DbTypePostgre, atlas.DbTypeCockroachDB:
		psr = parser_postgres.New(&conf.Schema)
	case atlas.DbTypeSQLite:
		psr = parser_sqlite.New(&conf.Schema)
	default:
		log.Panic().Msgf("invalid db type: %s", cfg.DB.Type)
	}

	// 5. gen code
	var gen *gen.ORNN = &gen.ORNN{}
	{
		gen.Init(conf, psr)
		if err = gen.GenCode(); err != nil { // code generate
			log.Panic().Err(err).Msg("code generate error")
		}
	}
}
