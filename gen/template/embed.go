package template

import _ "embed"

//go:embed func_insert.template
var InsertTmpl string

//go:embed func_select.template
var SelectTmpl string

//go:embed func_update.template
var UpdateTmpl string

//go:embed func_delete.template
var DeleteTmpl string
