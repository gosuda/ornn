package gen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gosuda/ornn/config"
	"github.com/gosuda/ornn/gen/template"
	"github.com/gosuda/ornn/parser"
)

type ORNN struct {
	conf *config.Config
	psr  parser.Parser
}

func (t *ORNN) Init(conf *config.Config, psr parser.Parser) {
	t.conf = conf
	t.psr = psr
}

func (t *ORNN) GenCode() (err error) {
	if t.conf == nil {
		return fmt.Errorf("config is emtpy")
	}

	// gen code
	gen := &Gen{}
	code, err := gen.Gen(t.conf, t.psr)
	if err != nil {
		return err
	}

	// write code to file
	genFile := t.conf.Global.FilePath + t.conf.Global.FileName
	if err := os.MkdirAll(filepath.Dir(genFile), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(genFile, []byte(code), 0600); err != nil {
		return err
	}
	if err := os.Chmod(genFile, 0600); err != nil {
		return err
	}

	// write use case
	useCase := template.UseCase(t.conf.Global.PackageName, t.conf.Global.ClassName)

	genUseCase := t.conf.Global.FilePath + "use_case.go"
	if err := os.MkdirAll(filepath.Dir(genUseCase), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(genUseCase, []byte(useCase), 0600); err != nil {
		return err
	}
	if err := os.Chmod(genUseCase, 0600); err != nil {
		return err
	}

	return nil
}
