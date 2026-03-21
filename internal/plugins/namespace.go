package plugins

import (
	"fmt"
	"strings"
)

// NamespacedItem represents an item that can be resolved by short or qualified name.
type NamespacedItem struct {
	ShortName     string // e.g. "fps"
	QualifiedName string // e.g. "starter:fps"
	Data          interface{}
}

// AmbiguousNameError is returned when a short name matches multiple items.
type AmbiguousNameError struct {
	Name       string
	Candidates []string
}

func (e *AmbiguousNameError) Error() string {
	msg := fmt.Sprintf("multiple matches for %q:\n", e.Name)
	for _, c := range e.Candidates {
		msg += fmt.Sprintf("  %s\n", c)
	}
	msg += fmt.Sprintf("Use the full name: <plugin>:%s", e.Name)
	return msg
}

// ResolveNamespace resolves a name (short or qualified) against a list of namespaced items.
func ResolveNamespace(name string, items []NamespacedItem) (NamespacedItem, error) {
	// Try qualified name first
	if strings.Contains(name, ":") {
		for _, item := range items {
			if item.QualifiedName == name {
				return item, nil
			}
		}
		return NamespacedItem{}, fmt.Errorf("%q not found", name)
	}

	// Try short name
	var matches []NamespacedItem
	for _, item := range items {
		if item.ShortName == name {
			matches = append(matches, item)
		}
	}

	switch len(matches) {
	case 0:
		return NamespacedItem{}, fmt.Errorf("%q not found", name)
	case 1:
		return matches[0], nil
	default:
		candidates := make([]string, len(matches))
		for i, m := range matches {
			candidates[i] = m.QualifiedName
		}
		return NamespacedItem{}, &AmbiguousNameError{Name: name, Candidates: candidates}
	}
}
