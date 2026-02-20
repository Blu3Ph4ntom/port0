package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func FromCwd(cwd string) string {
	base := filepath.Base(cwd)
	name := strings.ToLower(base)
	name = nonAlphaNum.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		return "project"
	}
	return name
}

func Deconflict(base string, taken map[string]bool) string {
	if !taken[base] {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !taken[candidate] {
			return candidate
		}
	}
}
