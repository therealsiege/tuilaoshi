package ui

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

// Every dashboard key must have a row in the feature map.
func TestFeatureMapListsDashboardKeys(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join("..", "..", "docs", "features.md"))
	if err != nil {
		t.Fatal(err)
	}
	keys := reflect.ValueOf(DefaultDashboardKeys())
	for i := 0; i < keys.NumField(); i++ {
		b, ok := keys.Field(i).Interface().(key.Binding)
		if !ok {
			continue
		}
		row := "| `" + b.Help().Key + "` |"
		if !strings.Contains(string(doc), row) {
			t.Errorf("docs/features.md has no row for dashboard key %q (%s)", b.Help().Key, b.Help().Desc)
		}
	}
}
