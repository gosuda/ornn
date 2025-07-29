## ornn: Object Relation No Need

Ornn is a tool that generates database-using code from schema and sql

built in atlas ( https://github.com/ariga/atlas )

1. built your db schema

2. run ornn

3. code generated, all done!

4. If additional queries are required, add in the config.

오른이 되세요

## Quick Start

```go
package main

import (
	"fmt"
	"os"

	"github.com/gosuda/ornn/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Errorf("error : %v", err)
		os.Exit(1)
	}
}
```

## Build

    go get github.com/gosuda/ornn
    cd $GOPATH/src/github.com/gosuda/ornn
    go build .


## Run

```
Usage:
  ornn [flags]

Flags:
  -A, --db_addr string            database server address (default "127.0.0.1")
  -i, --db_id string              database server id
  -n, --db_name string            database name
      --db_path string            path for save db files. sqlite only (default "./output/sqlite-database.db")
  -P, --db_port string            database server port (default "3306")
  -p, --db_pw string              database server password
  -D, --db_type string            database type ( mysql, mariadb, postgres, sqlite, tidb, cockroachdb ) (default "mysql")
      --file_config_load bool     load config from existing file ( default false )
      --file_config_path string   config json file path (default "./output/config.json")
      --file_gen_path string      generate golang file path (default "./output/gen.go")
      --file_schema_load bool     load schema from existing file and migrate database ( default false )
      --file_schema_path string   schema hcl file path (default "./output/schema.hcl")
      --gen_class_name string     class name (default "Gen")
      --gen_do_not_edit string    do not edit comment (default "// Code generated - DO NOT EDIT.\n// This file is a generated and any changes will be lost.\n")
      --gen_package_name string   package name (default "gen")
  -h, --help                      help for ornn
```

Example

    ./ornn -D mysql -A 127.0.0.1 -P 3306 -i root -p 1234 -n db_name
