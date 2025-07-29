package template

import _ "embed"

//go:embed insert.template
var InsertTmpl string

//go:embed select.template
var SelectTmpl string

//go:embed update.template
var UpdateTmpl string

//go:embed delete.template
var DeleteTmpl string

//go:embed use_case.template
var UseCaseTmpl string
