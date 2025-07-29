package config

import (
	"encoding/json"
	"io/ioutil"

	"ariga.io/atlas/sql/schema"
	"github.com/gosuda/ornn/config/atlas"
)

type Config struct {
	Global  Global  `json:"global"`
	Queries Queries `json:"queries"`

	Schema Schema `json:"-"`
}

// TODO - 추후 config 형식 변경 예정
func (t *Config) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	return nil
}

func (t *Config) Init(dbType atlas.DbType, schema *schema.Schema, filePath, fileName, packageName, className, doNotEdit string) error {
	// init global config
	t.Global.FilePath = filePath
	t.Global.FileName = fileName
	t.Global.PackageName = packageName
	t.Global.ClassName = className
	t.Global.DoNotEdit = doNotEdit

	t.Global.Import = []*Import{ // TODO
		{Alias: "", Path: "fmt"},
		{Alias: ".", Path: "github.com/gosuda/ornn/db"},
	}

	// init schema
	t.Schema.Init(dbType, schema)

	// init queries by schema
	t.Queries.init(&t.Schema)

	return nil
}

func (t *Config) Save(path string) error {
	data, err := json.MarshalIndent(&t, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, 0700)
	if err != nil {
		return err
	}
	return nil
}
