package util

import (
	"fmt"
	"strings"
)

const (
	PrepareStatementDelimeter = "%"
	PrepareStatementAfter     = "?"

	TplDelimiter = "#"
	TplAfter     = "%s"
	TplSplit     = "/"
)

// ClearInQuot replaces all characters inside quotes with spaces.
func ClearInQuot(s string) string {
	var inSingle, inDouble, inBacktick bool
	var out strings.Builder
	out.Grow(len(s))

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '\'':
			inSingle = !inSingle
			out.WriteByte(' ')
		case '"':
			inDouble = !inDouble
			out.WriteByte(' ')
		case '`':
			inBacktick = !inBacktick
			out.WriteByte(' ')
		default:
			if inSingle || inDouble || inBacktick {
				out.WriteByte(' ')
			} else {
				out.WriteByte(ch)
			}
		}
	}
	return out.String()
}

// SplitByDelimiter splits SQL into two parts at the first occurrence of delimiter.
func SplitByDelimiter(sql, delimiter string) (before, after string) {
	cleaned := ClearInQuot(strings.ToLower(sql))
	if pos := strings.Index(cleaned, delimiter); pos != -1 {
		return sql[:pos], sql[pos:]
	}
	return sql, ""
}

// ExportBetweenDelimiter extracts all values between delimiters and checks for duplicates.
func ExportBetweenDelimiter(input, delimiter string) ([]string, error) {
	cleaned := ClearInQuot(input)
	var outputs []string
	var buf strings.Builder
	in := false

	for i := 0; i < len(input); i++ {
		ch, qch := input[i], cleaned[i]
		if string(qch) == delimiter {
			if in {
				val := buf.String()
				for _, existing := range outputs {
					if existing == val {
						return nil, fmt.Errorf("duplicate field name: %s", val)
					}
				}
				outputs = append(outputs, val)
				buf.Reset()
			}
			in = !in
			continue
		}
		if in {
			buf.WriteByte(ch)
		}
	}
	return outputs, nil
}

// ReplaceBetweenDelimiter replaces content inside delimiters with a fixed string.
func ReplaceBetweenDelimiter(input, delimiter, replace string) string {
	cleaned := ClearInQuot(input)
	var out strings.Builder
	in := false

	for i := 0; i < len(input); i++ {
		ch, qch := input[i], cleaned[i]
		if string(qch) == delimiter {
			if in {
				out.WriteString(replace)
			}
			in = !in
			continue
		}
		if !in {
			out.WriteByte(ch)
		}
	}
	return out.String()
}

// ClearDelimiter removes only the delimiter characters, preserving the original string.
func ClearDelimiter(input, delimiter string) string {
	cleaned := ClearInQuot(input)
	var out strings.Builder
	for i := 0; i < len(input); i++ {
		if string(cleaned[i]) != delimiter {
			out.WriteByte(input[i])
		}
	}
	return out.String()
}

// ReplaceInDelimiter keeps only the last segment after a splitter inside the delimiter.
func ReplaceInDelimiter(input, delimiter, splitter string) string {
	// e.g., xxxx#AAAA#xxxx -> xxxxAAAAxxxx
	//       xxxx#AAAA/BBBB#xxxx -> xxxxBBBBxxxx
	cleaned := ClearInQuot(input)
	var out, buf strings.Builder
	in := false

	for i := 0; i < len(input); i++ {
		ch, qch := input[i], cleaned[i]
		if string(qch) == delimiter {
			in = !in
			if in {
				buf.Reset() // enter
			} else {
				out.WriteString(buf.String()) // exit
			}
			continue
		}

		if !in {
			out.WriteByte(ch)
		} else {
			if string(qch) == splitter {
				buf.Reset() // keep only after splitter
			} else {
				buf.WriteByte(ch)
			}
		}
	}
	return out.String()
}

// ExportInsertQueryValues extracts the part inside the first VALUES (...) in an INSERT statement.
func ExportInsertQueryValues(sqlInsert string) string {
	lower := strings.ToLower(sqlInsert)
	idx := strings.Index(lower, "values")
	if idx == -1 {
		return ""
	}

	var start, end int
	for i := idx; i < len(sqlInsert); i++ {
		switch sqlInsert[i] {
		case '(':
			if start == 0 {
				start = i
			}
		case ')':
			end = i
		}
	}
	if start < end {
		return sqlInsert[start+1 : end]
	}
	return ""
}

// ConvFirstToUpper makes the first character upper case.
func ConvFirstToUpper(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// IsParserValArg checks if a value looks like a parser variable e.g. :v0, :v1
func IsParserValArg(val []byte) bool {
	s := string(val)
	return len(s) >= 3 && strings.HasPrefix(s, ":v")
}
