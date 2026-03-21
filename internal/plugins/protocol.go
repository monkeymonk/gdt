package plugins

import "strings"

// StatusResult represents a single parsed status line from a plugin.
type StatusResult struct {
	Status  string // "OK", "WARN", or "FAIL"
	Message string
}

// ParseStatusLine parses a single line of the plugin line protocol.
// Format: STATUS message
// Returns the status, message, and whether the line was valid.
func ParseStatusLine(line string) (status string, msg string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}

	parts := strings.SplitN(line, " ", 2)
	s := parts[0]

	switch s {
	case "OK", "WARN", "FAIL":
		m := ""
		if len(parts) > 1 {
			m = parts[1]
		}
		return s, m, true
	default:
		return "", "", false
	}
}

// ParseStatusLines parses multi-line plugin output into status results.
func ParseStatusLines(output string) []StatusResult {
	var results []StatusResult
	for _, line := range strings.Split(output, "\n") {
		status, msg, ok := ParseStatusLine(line)
		if ok {
			results = append(results, StatusResult{Status: status, Message: msg})
		}
	}
	return results
}
