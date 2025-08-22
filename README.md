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
```
go get github.com/gosuda/ornn
cd $GOPATH/src/github.com/gosuda/ornn/cmd
go build -o ornn .
```

## Run

### 1. Create Config file (config.toml)

```
[DB]
  Type     = "mysql"
  Addr     = "localhost"
  Port     = "3306"
  User     = "user"
  Password = "1234"
  Name     = "db_name"
  Path     = "./"  # for sqlite only

[Gen]
  SchemaPath   = "../output/schema.hcl"
  ConfigPath   = "../output/config.json"
  GenPath      = "../output/output_"
  FileName     = "gen.go"
  PackageName  = "gen"
  ClassName    = "Gen"
```

### 2. Run ornn with config
```
./ornn --load_schema=true --load_config=false
```