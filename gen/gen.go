package gen

import (
	"fmt"

	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/parser"
	"github.com/rs/zerolog/log"
)

type Gen struct {
	data *GenQueries
	code *GenCode
}

func (t *Gen) Gen(conf *config.Config, psr parser.Parser) (code string, err error) {
	// set query data for generate code
	t.data = &GenQueries{}
	t.data.Init(conf, psr)
	err = t.data.SetData()
	if err != nil {
		return "", err
	}

	// check query error
	for tableName, def := range conf.Queries.Class {
		for _, query := range def {
			if query.ErrParser != "" {
				log.Error().Str("table name", tableName).Str("query name", query.Name).Str("err", query.ErrParser).Msg("parser err")
				err = fmt.Errorf("query error")
			}
			if query.ErrQuery != "" {
				log.Error().Str("table name", tableName).Str("query name", query.Name).Str("err", query.ErrParser).Msg("parser err")
				err = fmt.Errorf("query error")
			}
		}
	}

	if err != nil {
		return "", err
	}

	// gen code
	t.code = &GenCode{}
	code, err = t.code.code(conf, t.data)
	if err != nil {
		return "", err
	}

	return code, nil
}
