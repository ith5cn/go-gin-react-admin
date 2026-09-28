package system

import (
	"os"
	"regexp"
	"strings"
	"testing"

	systemModel "server/model/system"
)

func TestArticleSeedOptionsCanGenerate(t *testing.T) {
	baseline, err := os.ReadFile("../../database/ai_system.sql")
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`'(category_id|status|is_link|is_hot)'.*'select', NULL, 'static', '([^']+)'`)
	var columns []systemModel.ToolGenerateColumn
	for _, line := range strings.Split(string(baseline), "\n") {
		if !strings.HasPrefix(line, "INSERT INTO `nest_tool_generate_columns`") {
			continue
		}
		match := pattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		columns = append(columns, systemModel.ToolGenerateColumn{
			ColumnName: match[1], ViewType: "select", IsInsert: 2, IsEdit: 2,
			OptionSource: ptrStr("static"), OptionConfig: ptrStr(match[2]),
		})
	}
	if len(columns) != 4 {
		t.Fatalf("article seed has %d configured option fields, want 4", len(columns))
	}
	if err := validateCodegenColumns(columns); err != nil {
		t.Fatalf("seed cannot generate: %v", err)
	}
	config, err := parseCodegenOptionConfig(columns[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Options) != 10 || config.Options[0].Value != float64(1) {
		t.Fatalf("invalid category options: %#v", config.Options)
	}
}
