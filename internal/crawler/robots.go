package crawler

import (
	"regexp"
	"strings"
)

type robotRule struct {
	path  string
	allow bool
}
type robotGroup struct {
	agents []string
	rules  []robotRule
}

// Áp dụng nhóm User-agent cụ thể trước nhóm *, rồi chọn quy tắc đường dẫn dài
// nhất; Allow thắng nếu cùng độ dài. Chỉ đọc các đường dẫn public /hang/*.
func robotsAllowed(body, path string) bool {
	var groups []robotGroup
	g := robotGroup{}
	hasRules := false
	flush := func() {
		if len(g.agents) > 0 {
			groups = append(groups, g)
		}
		g = robotGroup{}
		hasRules = false
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if key == "user-agent" {
			if hasRules {
				flush()
			}
			g.agents = append(g.agents, strings.ToLower(value))
			continue
		}
		if key == "allow" || key == "disallow" {
			hasRules = true
			if value != "" {
				g.rules = append(g.rules, robotRule{value, key == "allow"})
			}
		}
	}
	flush()
	specific := false
	for _, g := range groups {
		for _, a := range g.agents {
			if a != "*" && a != "" && strings.Contains(strings.ToLower(userAgent), a) {
				specific = true
			}
		}
	}
	longest := -1
	allowed := true
	for _, g := range groups {
		matches := false
		for _, a := range g.agents {
			if specific {
				matches = matches || (a != "*" && a != "" && strings.Contains(strings.ToLower(userAgent), a))
			} else {
				matches = matches || a == "*"
			}
		}
		if !matches {
			continue
		}
		for _, rule := range g.rules {
			pattern := rule.path
			end := strings.HasSuffix(pattern, "$")
			pattern = strings.TrimSuffix(pattern, "$")
			pattern = "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, ".*")
			if end {
				pattern += "$"
			}
			if regexp.MustCompile(pattern).MatchString(path) {
				n := len(strings.ReplaceAll(rule.path, "*", ""))
				if n > longest || n == longest && rule.allow {
					longest = n
					allowed = rule.allow
				}
			}
		}
	}
	return allowed
}
