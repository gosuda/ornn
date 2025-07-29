package atlas

import (
	"fmt"
	"os"
	"testing"

	"github.com/gosuda/ornn/db/db_mysql"
	"github.com/gosuda/ornn/db/db_postgres"
	"github.com/stretchr/testify/require"
)

func TestMysql(t *testing.T) {
	db, err := db_mysql.New("127.0.0.1", "3306", "root", "951753ck", "test")
	require.NoError(t, err)

	atl := &Atlas{}
	err = atl.Init(DbTypeMySQL, db)
	require.NoError(t, err)

	sch, err := atl.InspectSchema()
	require.NoError(t, err)

	bt, err := atl.MarshalHCL(sch)
	require.NoError(t, err)

	err = os.WriteFile("./schema_mysql.hcl", bt, 0700)
	require.NoError(t, err)

	schNew, err := atl.UnmarshalHCL(bt)
	require.NoError(t, err)
	fmt.Println("db name :", schNew.Name)
}

func TestPostgres(t *testing.T) {
	db, err := db_postgres.New("127.0.0.1", "5432", "postgres", "", "postgres")
	require.NoError(t, err)

	atl := &Atlas{}
	err = atl.Init(DbTypePostgre, db)
	require.NoError(t, err)

	sch, err := atl.InspectSchema()
	require.NoError(t, err)

	bt, err := atl.MarshalHCL(sch)
	require.NoError(t, err)

	err = os.WriteFile("./schema_postgresql.hcl", bt, 0700)
	require.NoError(t, err)

	schNew, err := atl.UnmarshalHCL(bt)
	require.NoError(t, err)
	fmt.Println("db name :", schNew.Name)
}
