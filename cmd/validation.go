package main

import (
	"errors"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosuda/ornn/atlas"
)

func validateConfig(cfg *Config, loadSchema, loadGeneratedConfig bool) error {
	if cfg == nil {
		return NewUserError("config-validate", "config is empty", "provide a valid config file with [DB] and [Gen] sections")
	}
	if strings.TrimSpace(cfg.DB.Type) == "" {
		return NewUserError("config-validate", "DB.Type is required", "set DB.Type in config.toml")
	}
	dbType, ok := atlas.DbTypeStrReverse[cfg.DB.Type]
	if !ok || dbType == atlas.DbTypeEmpty {
		return NewUserError("config-validate", fmt.Sprintf("unsupported db type: %s", cfg.DB.Type), "use one of: mysql, mariadb, postgres, sqlite, tidb, cockroachdb")
	}

	if err := validateDBConfig(cfg, dbType); err != nil {
		return err
	}
	if err := validateGenConfig(cfg, loadSchema, loadGeneratedConfig); err != nil {
		return err
	}
	return nil
}

func validateDBConfig(cfg *Config, dbType atlas.DbType) error {
	switch dbType {
	case atlas.DbTypeSQLite:
		if strings.TrimSpace(cfg.DB.Path) == "" {
			return NewUserError("config-validate", "DB.Path is required for sqlite", "set DB.Path to the sqlite file path")
		}
	default:
		missing := make([]string, 0, 5)
		if strings.TrimSpace(cfg.DB.Addr) == "" {
			missing = append(missing, "DB.Addr")
		}
		if strings.TrimSpace(cfg.DB.Port) == "" {
			missing = append(missing, "DB.Port")
		}
		if strings.TrimSpace(cfg.DB.User) == "" {
			missing = append(missing, "DB.User")
		}
		if strings.TrimSpace(cfg.DB.Name) == "" {
			missing = append(missing, "DB.Name")
		}
		if len(missing) > 0 {
			return NewUserError("config-validate", "missing required DB fields: "+strings.Join(missing, ", "), "fill required DB fields in config.toml")
		}
	}
	return nil
}

func validateGenConfig(cfg *Config, loadSchema, loadGeneratedConfig bool) error {
	if strings.TrimSpace(cfg.Gen.SchemaPath) == "" {
		return NewUserError("config-validate", "Gen.SchemaPath is required", "set Gen.SchemaPath in config.toml")
	}
	if strings.TrimSpace(cfg.Gen.ConfigPath) == "" {
		return NewUserError("config-validate", "Gen.ConfigPath is required", "set Gen.ConfigPath in config.toml")
	}
	if strings.TrimSpace(cfg.Gen.GenPath) == "" {
		return NewUserError("config-validate", "Gen.GenPath is required", "set Gen.GenPath in config.toml")
	}
	if strings.TrimSpace(cfg.Gen.FileName) == "" {
		return NewUserError("config-validate", "Gen.FileName is required", "set Gen.FileName in config.toml")
	}
	if cfg.Gen.FileName != filepath.Base(cfg.Gen.FileName) {
		return NewUserError("config-validate", "Gen.FileName must not contain a path", "move directory path into Gen.GenPath and keep only file name in Gen.FileName")
	}
	if !isValidGoIdentifier(cfg.Gen.PackageName) {
		return NewUserError("config-validate", "Gen.PackageName must be a valid Go identifier", "use letters, digits, and underscores; do not use a Go keyword")
	}
	if !isValidGoIdentifier(cfg.Gen.ClassName) {
		return NewUserError("config-validate", "Gen.ClassName must be a valid Go identifier", "use letters, digits, and underscores; do not use a Go keyword")
	}
	if loadSchema {
		if _, err := os.Stat(cfg.Gen.SchemaPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return NewUserError("config-validate", fmt.Sprintf("schema file does not exist: %s", cfg.Gen.SchemaPath), "set Gen.SchemaPath to an existing schema file or use --load_schema=false")
			}
			return NewSystemError("config-validate", err, "when --load_schema=true, ensure Gen.SchemaPath is readable")
		}
	}
	if loadGeneratedConfig {
		if _, err := os.Stat(cfg.Gen.ConfigPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return NewUserError("config-validate", fmt.Sprintf("generated config file does not exist: %s", cfg.Gen.ConfigPath), "set Gen.ConfigPath to an existing config file or use --load_config=false")
			}
			return NewSystemError("config-validate", err, "when --load_config=true, ensure Gen.ConfigPath is readable")
		}
	}
	return nil
}

func isValidGoIdentifier(name string) bool {
	return name != "_" && token.IsIdentifier(name)
}
