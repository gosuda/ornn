// Command ornn generates database access code from a schema and query config.
//
// Build:
//
//	go build -o ornn ./cmd
//
// Run:
//
//	./ornn --config ./config.toml
package main

import (
	"errors"
	"fmt"
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
	if err := rootCmd.Execute(); err != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			return err
		}
		return NewUserError("cli", err.Error(), "run ornn --help for valid options")
	}
	return nil
}

var (
	rootCmd = &cobra.Command{
		Use:           "ornn",
		Short:         "ornn is a code generator for golang",
		Long:          "ornn is a code generator for golang db access",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          rootRun,
	}

	loadExistSchemaFile bool // 기존 스키마 파일에서 로딩, 스키마 파일대로 db migrate
	loadExistConfigFile bool // 기존 설정 파일에서 로딩
	configFilePath      string
)

func init() {
	fs := rootCmd.PersistentFlags()
	fs.StringVarP(&configFilePath, "config", "c", "config.toml", "Path to config file")
	fs.BoolVar(&loadExistSchemaFile, "load_schema", true, "load schema from existing file and migrate database")
	fs.BoolVar(&loadExistConfigFile, "load_config", false, "load config from existing file")
}

func main() {
	if err := Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewUserError("config-load", fmt.Sprintf("config file does not exist: %s", configFilePath), "provide a valid path with --config")
		}
		return NewSystemError("config-load", err, fmt.Sprintf("check config file path and format: %s", configFilePath))
	}
	if err := validateConfig(cfg, loadExistSchemaFile, loadExistConfigFile); err != nil {
		return err
	}
	atlasDbType, ok := atlas.DbTypeStrReverse[cfg.DB.Type]
	if !ok || atlasDbType == atlas.DbTypeEmpty {
		return NewUserError("config-validate", fmt.Sprintf("unsupported db type: %s", cfg.DB.Type), "use one of: mysql, mariadb, postgres, sqlite, tidb, cockroachdb")
	}

	// 1. connect db
	conn, err := connectDB(cfg, atlasDbType)
	if err != nil {
		return err
	}

	// 2. init schema from atl
	var sch *schema.Schema
	atl, err := atlas.New(atlasDbType, conn)
	if err != nil {
		return NewSystemError("atlas-init", err, "check database driver and connection settings")
	}
	if loadExistSchemaFile { // load from existing schema file
		if sch, err = atl.Load(cfg.Gen.SchemaPath); err != nil {
			return NewSystemError("schema-load", err, "check --load_schema and schema file path")
		}
		// migrate db from file
		if err = atl.MigrateSchema(sch); err != nil {
			return NewSystemError("schema-migrate", err, "check schema compatibility with target database")
		}
		// inspect schema fron migrated db
		if sch, err = atl.InspectSchema(); err != nil {
			return NewSystemError("schema-inspect", err, "check database accessibility and permissions")
		}
	} else {
		if sch, err = atl.InspectSchema(); err != nil {
			return NewSystemError("schema-inspect", err, "check database accessibility and permissions")
		}
		if err = atl.Save(cfg.Gen.SchemaPath, sch); err != nil {
			return NewSystemError("schema-save", err, "check output directory permissions for Gen.SchemaPath")
		}
	}

	// 3. set config
	var conf = &config.Config{}
	if loadExistConfigFile { // load from existing config file
		if err = conf.Load(cfg.Gen.ConfigPath); err != nil { // load
			return NewSystemError("config-load-generated", err, "check --load_config and Gen.ConfigPath")
		}
		if err = conf.Init(atlasDbType, sch, cfg.Gen.GenPath, cfg.Gen.FileName, cfg.Gen.PackageName, cfg.Gen.ClassName); err != nil { // init
			return NewSystemError("config-init", err, "check schema and generation settings")
		}
	} else {
		if err = conf.Init(atlasDbType, sch, cfg.Gen.GenPath, cfg.Gen.FileName, cfg.Gen.PackageName, cfg.Gen.ClassName); err != nil { // init
			return NewSystemError("config-init", err, "check schema and generation settings")
		}
		if err = conf.Save(cfg.Gen.ConfigPath); err != nil { // save
			return NewSystemError("config-save", err, "check output directory permissions for Gen.ConfigPath")
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
		return NewUserError("parser-init", fmt.Sprintf("unsupported db type: %s", cfg.DB.Type), "use one of: mysql, mariadb, postgres, sqlite, tidb, cockroachdb")
	}

	// 5. gen code
	var gen *gen.ORNN = &gen.ORNN{}
	{
		gen.Init(conf, psr)
		if err = gen.GenCode(); err != nil { // code generate
			return NewSystemError("code-generate", err, "check query config and generation path settings")
		}
	}
	log.Info().Str("generate path", cfg.Gen.GenPath).Msg("Code generated Succeed")
	return nil
}

func connectDB(cfg *Config, dbType atlas.DbType) (*db.Conn, error) {
	switch dbType {
	case atlas.DbTypeMySQL, atlas.DbTypeMaria, atlas.DbTypeTiDB:
		conn, err := db_mysql.New(db_mysql.Dsn(cfg.DB.User, cfg.DB.Password, cfg.DB.Addr, cfg.DB.Port, cfg.DB.Name), cfg.DB.Name)
		if err != nil {
			return nil, NewSystemError("db-connect", err, "check DB.Addr, DB.Port, DB.User, DB.Password, DB.Name")
		}
		return conn, nil
	case atlas.DbTypePostgre, atlas.DbTypeCockroachDB:
		conn, err := db_postgres.New(db_postgres.Dsn(cfg.DB.Addr, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name), cfg.DB.Name)
		if err != nil {
			return nil, NewSystemError("db-connect", err, "check DB.Addr, DB.Port, DB.User, DB.Password, DB.Name")
		}
		return conn, nil
	case atlas.DbTypeSQLite:
		conn, err := db_sqlite.New(cfg.DB.Path)
		if err != nil {
			return nil, NewSystemError("db-connect", err, "check DB.Path")
		}
		return conn, nil
	default:
		return nil, NewUserError("db-connect", fmt.Sprintf("unsupported db type: %s", cfg.DB.Type), "use one of: mysql, mariadb, postgres, sqlite, tidb, cockroachdb")
	}
}
