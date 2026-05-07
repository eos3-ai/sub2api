package handler

import "strings"

// sanitizeCSVCell mitigates CSV formula injection when opened in spreadsheet apps.
func sanitizeCSVCell(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed == "" {
		return value
	}
	switch trimmed[0] {
	case '=', '+', '-', '@':
		if strings.HasPrefix(value, "'") {
			return value
		}
		return "'" + value
	default:
		return value
	}
}
