package atlas

import (
	"context"
	"os"

	"ariga.io/atlas/schemahcl"
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlite"
	"github.com/gosuda/ornn/db"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func New(dbType DbType, conn *db.Conn) *Atlas {
	atl := &Atlas{}
	atl.Init(dbType, conn)
	return atl
}

type Atlas struct {
	DbName      string
	DbType      DbType
	marshaler   schemahcl.MarshalerFunc
	unmarshaler schemahcl.EvalFunc
	driver      migrate.Driver
}

func (t *Atlas) Init(dbType DbType, conn *db.Conn) error {
	var err error
	t.DbName = conn.DbName
	t.DbType = dbType
	switch dbType {
	case DbTypeMySQL, DbTypeMaria, DbTypeTiDB:
		t.marshaler = mysql.MarshalHCL
		t.unmarshaler = mysql.EvalHCL
		t.driver, err = mysql.Open(conn.Raw())
	case DbTypePostgre, DbTypeCockroachDB:
		t.marshaler = postgres.MarshalHCL
		t.unmarshaler = postgres.EvalHCL
		t.driver, err = postgres.Open(conn.Raw())
	case DbTypeSQLite:
		t.marshaler = sqlite.MarshalHCL
		t.unmarshaler = sqlite.EvalHCL
		t.driver, err = sqlite.Open(conn.Raw())
	}
	if err != nil {
		return err
	}
	return nil
}

func (t *Atlas) Save(path string, sch *schema.Schema) error {
	bt, err := t.MarshalHCL(sch)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bt, 0700)
}

func (t *Atlas) Load(path string) (*schema.Schema, error) {
	bt, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return t.UnmarshalHCL(bt)
}

func (t *Atlas) MarshalHCL(sch *schema.Schema) ([]byte, error) {
	bt, err := t.marshaler.MarshalSpec(sch)
	if err != nil {
		return nil, err
	}
	return bt, nil
}

func (t *Atlas) UnmarshalHCL(bt []byte) (*schema.Schema, error) {
	sch := schema.New("")
	parser := hclparse.NewParser()
	if _, diag := parser.ParseHCL(bt, ""); diag.HasErrors() {
		return nil, diag
	}
	err := t.unmarshaler.Eval(parser, sch, nil)
	if err != nil {
		return nil, err
	}
	return sch, nil
}

func (t *Atlas) InspectSchema() (*schema.Schema, error) {
	sch, err := t.driver.InspectSchema(context.Background(), "", nil)
	if err != nil {
		return nil, err
	}
	return sch, nil
}

func (t *Atlas) MigrateSchema(sch *schema.Schema) error {
	schemaCur, err := t.InspectSchema()
	if err != nil {
		return err
	}
	diffs, err := t.driver.SchemaDiff(schemaCur, sch)
	if err != nil {
		return err
	}
	return t.driver.ApplyChanges(context.Background(), diffs)
}
