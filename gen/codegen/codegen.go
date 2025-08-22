package codegen

import (
	"go/format"

	"golang.org/x/tools/imports"
)

type CodeGen struct {
	Global
}

func (t *CodeGen) Code() (code string) {
	writer := &Writer{}
	writer.Init()
	t.Global.Code(writer)

	return FormatAndFixImports(writer.Bytes())
}

func FormatAndFixImports(src []byte) (dst string) {
	out, err := format.Source(src)
	if err != nil {
		return ""
	}
	out, err = imports.Process("", out, nil)
	if err != nil {
		return ""
	}
	return string(out)
}
