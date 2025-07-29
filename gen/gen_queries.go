package gen

import (
	"fmt"

	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/parser"
)

type GenQueries struct {
	conf *config.Config
	psr  parser.Parser

	class map[string]map[string]*parser.ParsedQuery
}

func (t *GenQueries) Init(conf *config.Config, psr parser.Parser) {
	t.conf = conf
	t.psr = psr
	t.class = make(map[string]map[string]*parser.ParsedQuery)
}

func (t *GenQueries) SetData() (err error) {
	// schema
	for _, group := range t.conf.Schema.Tables {
		Queries, ok := t.conf.Queries.Class[group.Name]
		if ok != true {
			continue
		}
		err := t.SetDataGroup(group.Name, Queries)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *GenQueries) SetDataGroup(groupName string, queries []*config.Query) (err error) {
	if t.class == nil {
		t.class = make(map[string]map[string]*parser.ParsedQuery)
	} else if t.class[groupName] == nil {
		t.class[groupName] = make(map[string]*parser.ParsedQuery)
	}

	for _, query := range queries {
		parseQuery, err := t.SetDataQuery(groupName, query)
		if err != nil {
			return err
		}
		t.class[groupName][query.Name] = parseQuery
	}
	return nil
}

func (t *GenQueries) SetDataQuery(groupName string, query *config.Query) (parseQuery *parser.ParsedQuery, err error) {
	parseQuery, err = t.psr.Parse(query.Sql)
	if err != nil {
		query.ErrParser = fmt.Sprintf("%v", err)
		return nil, nil
	}
	return parseQuery, nil
}
