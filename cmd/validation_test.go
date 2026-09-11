package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.hcl")
	configPath := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(schemaPath, []byte("schema {}"), 0600))
	require.NoError(t, os.WriteFile(configPath, []byte("{}"), 0600))

	valid := &Config{
		DB: ConfigDB{
			Type:     "mysql",
			Addr:     "localhost",
			Port:     "3306",
			User:     "user",
			Password: "pw",
			Name:     "db_name",
		},
		Gen: ConfigGen{
			SchemaPath:  schemaPath,
			ConfigPath:  configPath,
			GenPath:     filepath.Join(tmpDir, "output_"),
			FileName:    "gen.go",
			PackageName: "gen",
			ClassName:   "Gen",
		},
	}

	t.Run("valid config", func(t *testing.T) {
		require.NoError(t, validateConfig(valid, true, true))
	})

	t.Run("missing db type", func(t *testing.T) {
		cfg := *valid
		cfg.DB.Type = ""
		err := validateConfig(&cfg, false, false)
		requireAppError(t, err, ErrorKindUser, "config-validate")
	})

	t.Run("sqlite requires path", func(t *testing.T) {
		cfg := *valid
		cfg.DB.Type = "sqlite"
		cfg.DB.Path = ""
		err := validateConfig(&cfg, false, false)
		requireAppError(t, err, ErrorKindUser, "config-validate")
	})

	t.Run("mysql requires addr", func(t *testing.T) {
		cfg := *valid
		cfg.DB.Addr = ""
		err := validateConfig(&cfg, false, false)
		requireAppError(t, err, ErrorKindUser, "config-validate")
	})

	t.Run("filename cannot contain path", func(t *testing.T) {
		cfg := *valid
		cfg.Gen.FileName = "dir/gen.go"
		err := validateConfig(&cfg, false, false)
		requireAppError(t, err, ErrorKindUser, "config-validate")
	})

	t.Run("missing schema file is user error", func(t *testing.T) {
		cfg := *valid
		cfg.Gen.SchemaPath = filepath.Join(tmpDir, "missing.hcl")
		err := validateConfig(&cfg, true, false)
		requireAppError(t, err, ErrorKindUser, "config-validate")
	})
	t.Run("package name must be a Go identifier", func(t *testing.T) {
		cfg := *valid
		cfg.Gen.PackageName = "package-name"
		err := validateConfig(&cfg, false, false)
		requireAppError(t, err, ErrorKindUser, "config-validate")
		require.Contains(t, err.Error(), "Gen.PackageName")
	})

	t.Run("class name must not be a Go keyword", func(t *testing.T) {
		cfg := *valid
		cfg.Gen.ClassName = "type"
		err := validateConfig(&cfg, false, false)
		requireAppError(t, err, ErrorKindUser, "config-validate")
		require.Contains(t, err.Error(), "Gen.ClassName")
	})
}

func requireAppError(t *testing.T, err error, kind ErrorKind, stage string) {
	t.Helper()

	var appErr *AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, kind, appErr.Kind)
	require.Equal(t, stage, appErr.Stage)
}
