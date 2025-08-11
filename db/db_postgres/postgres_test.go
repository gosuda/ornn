package db_postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConn(t *testing.T) {
	conn, err := New(Dsn("localhost", "5432", "postgres", "", "postgres"), "postgres")
	require.NoError(t, err)

	err = conn.Raw().Ping()
	require.NoError(t, err)
	// conn.Db.Query()
}
