package tools

import "strings"

func ClearPathStringForWindows(oldString string) string {
	r := strings.NewReplacer("?", "_", "*", "_", ":", "_", "|", "_", "\\", "_", "\"", "_")
	return r.Replace(oldString)
}
