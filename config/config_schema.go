package config

import (
	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/atlas"
)

type Schema struct {
	DbType atlas.DbType `json:"-"`

	*schema.Schema `json:"-"`
}

func (t *Schema) Init(dbType atlas.DbType, sch *schema.Schema) {
	t.DbType = dbType
	t.Schema = sch
}

func (t *Schema) GetFieldTypeAll(fieldName string) (fieldTypeByTable map[string]string, exist bool) {
	fieldTypeByTable = make(map[string]string)
	for _, tbl := range t.Tables {
		fld, exist := tbl.Column(fieldName)
		if exist {
			fieldTypeByTable[tbl.Name] = fld.Type.Raw
			exist = true
		}
	}
	return fieldTypeByTable, exist
}

func (t *Schema) GetFieldType(tableName, fieldName string) (fieldType string, exist bool) {
	if tbl, exist := t.Table(tableName); exist {
		if fld, exist := tbl.Column(fieldName); exist {
			return fld.Type.Raw, true
		}
	}
	return "", false
}
