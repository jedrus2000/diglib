package tools

import "strings"

func min(is ...int) int {
	min := is[0]
	for _, i := range is[1:] {
		if i < min {
			min = i
		}
	}
	return min
}

func ClearPathStringForWindows(oldString string) string {
	r := strings.NewReplacer("?", "_", "*", "_", ":", "_", "|", "_", "\\", "_", "\"", "_")
	s := r.Replace(oldString)
	return s[:min(len(s), 200)]
}
