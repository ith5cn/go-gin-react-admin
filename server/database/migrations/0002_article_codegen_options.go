package migrations

import (
	"encoding/json"

	"gorm.io/gorm"
)

// repairArticleCodegenOptions fills only the incomplete article example shipped in the baseline.
// Category options are a snapshot; customized sources and dictionaries remain untouched.
func repairArticleCodegenOptions(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		base := func(column, view string) *gorm.DB {
			return tx.Table("nest_tool_generate_columns").
				Where("table_id IN (?)", tx.Table("nest_tool_generate_tables").Select("id").
					Where("table_name = ? AND business_name = ? AND package_name IN ? AND source = ? AND delete_time IS NULL", "ai_article", "ai-article", []string{"system", "generated"}, "ai_system")).
				Where("column_name = ? AND view_type = ? AND delete_time IS NULL", column, view).
				Where("COALESCE(TRIM(dict_type), '') = '' AND COALESCE(TRIM(option_source), '') = '' AND COALESCE(TRIM(option_config), '') = ''")
		}
		var missing int64
		if err := base("category_id", "select").Count(&missing).Error; err != nil {
			return err
		}
		if missing > 0 && tx.Migrator().HasTable("ai_article_category") {
			var options []struct {
				Label string `json:"label" gorm:"column:label"`
				Value uint   `json:"value" gorm:"column:value"`
			}
			if err := tx.Table("ai_article_category").Select("category_name AS label, id AS value").
				Where("delete_time IS NULL AND status = ?", 1).Order("sort ASC, id ASC").Scan(&options).Error; err != nil {
				return err
			}
			if len(options) > 0 {
				config, err := json.Marshal(map[string]interface{}{"options": options})
				if err != nil {
					return err
				}
				if err := base("category_id", "select").Updates(map[string]interface{}{"option_source": "static", "option_config": string(config)}).Error; err != nil {
					return err
				}
			}
		}
		for _, field := range []struct{ name, config string }{
			{"status", `{"options":[{"label":"正常","value":1},{"label":"停用","value":2}]}`},
			{"is_link", `{"options":[{"label":"是","value":1},{"label":"否","value":2}]}`},
			{"is_hot", `{"options":[{"label":"是","value":1},{"label":"否","value":2}]}`},
		} {
			if err := base(field.name, "saSelect").Updates(map[string]interface{}{
				"view_type": "select", "option_source": "static", "option_config": field.config,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
