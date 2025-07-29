package template

import (
	"fmt"
	"strings"
	"text/template"
)

// parseTemplate parses and executes a template string with given args.
func parseTemplate(tmplStr string, args map[string]any) string {
	tmpl, err := template.New(tmplStr).Parse(tmplStr)
	if err != nil {
		return ""
	}
	builder := &strings.Builder{}
	if err := tmpl.Execute(builder, args); err != nil {
		return ""
	}
	return builder.String()
}

// genQuery_body_setArgs generates Go code for args := []any{...}
func genQuery_body_setArgs(args []string) string {
	items := genQuery_body_arg(args)
	if items != "" {
		items += "\n"
	}
	return fmt.Sprintf("args := []any{%s}\n", items)
}

// genQuery_body_arg formats args as a comma-separated list with indentation
func genQuery_body_arg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	var b strings.Builder
	for _, arg := range args {
		fmt.Fprintf(&b, "\n\t%s,", arg)
	}
	return b.String()
}

// genQuery_body_multiInsertProc generates code for multi-insert procedure
func genQuery_body_multiInsertProc(args []string) string {
	if len(args) == 0 {
		return ""
	}

	// checkLen: argLen != len(arg1) || argLen != len(arg2) ...
	checkConds := make([]string, len(args))
	for i, arg := range args {
		checkConds[i] = fmt.Sprintf("argLen != len(%s)", arg)
	}
	checkLen := strings.Join(checkConds, " || ")

	// append part: arg1[i], arg2[i], ...
	appendArgs := make([]string, len(args))
	for i, arg := range args {
		appendArgs[i] = fmt.Sprintf("%s[i]", arg)
	}

	return fmt.Sprintf(`argLen := len(%s)
if argLen == 0 {
	return 0, fmt.Errorf("arg len is zero")
}
if %s {
	return 0, fmt.Errorf("arg len is not same")
}

args := make([]any, 0, argLen*%d)
for i := 0; i < argLen; i++ {
	args = append(args, I_to_arri(
		%s,
	)...)
}
`,
		args[0],
		checkLen,
		len(args),
		strings.Join(appendArgs, ",\n\t\t"),
	)
}

// genQuery_body_multiInsert generates code snippet for multi insert SQL
func genQuery_body_multiInsert(query string) string {
	return fmt.Sprintf("\n\tstrings.Repeat(\", (%s)\", argLen-1),", query)
}
