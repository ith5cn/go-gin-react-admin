// Package query provides database-injected query helpers; it has no system business dependencies.
package query

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	commonResponse "server/model/common/response"

	"gorm.io/gorm"
)

// ErrNoDatabase indicates missing dependency injection.
var ErrNoDatabase = errors.New("database is not initialized")

// PageQuery contains normalized pagination.
type PageQuery struct{ Page, Size int }

// ParsePage supports page, size and the legacy limit parameter, capped at 1000 rows.
func ParsePage(query map[string]string) PageQuery {
	page, _ := strconv.Atoi(query["page"])
	size, _ := strconv.Atoi(query["size"])
	if size <= 0 {
		size, _ = strconv.Atoi(query["limit"])
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 1000 {
		size = 1000
	}
	// Avoid integer overflow in the SQL offset.
	if page > int(^uint(0)>>1)/size {
		page = 1
	}
	return PageQuery{Page: page, Size: size}
}

// QueryFilter 描述一个查询条件：Param 是前端参数名，Column 是数据库列名，
// Op 是操作符（eq/neq/gt/gte/lt/lte/like/in/notin/between）。
type QueryFilter struct {
	Param  string
	Column string
	Op     string
}

// PageListFiltered 是 codegen 生成代码的分页查询契约（PageList 的增强版）：
// 支持全量查询操作符、orderBy/orderType 排序白名单和软删除开关。
func PageListFiltered(db *gorm.DB, query map[string]string, model interface{}, dest interface{}, filters []QueryFilter, sortable map[string]string, defaultOrder string, useSoftDelete bool) (*commonResponse.PageResult, error) {
	if db == nil {
		return nil, ErrNoDatabase
	}
	page := ParsePage(query)
	base := db.Model(model)
	if useSoftDelete {
		base = base.Where("delete_time IS NULL")
	}
	base = applyQueryFilters(base, query, filters)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	order := resolveOrder(query, sortable, defaultOrder)
	if err := base.Order(order).Offset((page.Page - 1) * page.Size).Limit(page.Size).Find(dest).Error; err != nil {
		return nil, err
	}
	return &commonResponse.PageResult{List: dest, Total: total}, nil
}

// SoftDeleteRecord 通过 UPDATE delete_time 实现软删除，供 codegen 生成的软删模型使用。
func SoftDeleteRecord(db *gorm.DB, table string, id string) error {
	if db == nil {
		return ErrNoDatabase
	}
	return db.Table(table).Where("id = ?", id).
		Updates(map[string]interface{}{"delete_time": gorm.Expr("NOW()"), "update_time": gorm.Expr("NOW()")}).Error
}

func applyQueryFilters(db *gorm.DB, query map[string]string, filters []QueryFilter) *gorm.DB {
	for _, filter := range filters {
		value := query[filter.Param]
		if value == "" {
			continue
		}
		column := filter.Column
		switch filter.Op {
		case "neq":
			db = db.Where(column+" <> ?", value)
		case "gt":
			db = db.Where(column+" > ?", value)
		case "gte":
			db = db.Where(column+" >= ?", value)
		case "lt":
			db = db.Where(column+" < ?", value)
		case "lte":
			db = db.Where(column+" <= ?", value)
		case "like":
			db = db.Where(column+" LIKE ?", "%"+value+"%")
		case "in":
			db = db.Where(column+" IN ?", splitQueryValues(value))
		case "notin":
			db = db.Where(column+" NOT IN ?", splitQueryValues(value))
		case "between":
			parts := splitQueryValues(value)
			if len(parts) >= 2 {
				db = db.Where(column+" BETWEEN ? AND ?", parts[0], parts[1])
			}
		default:
			db = db.Where(column+" = ?", value)
		}
	}
	return db
}

// resolveOrder 只接受排序白名单里的列，避免 orderBy 参数注入任意 SQL。
func resolveOrder(query map[string]string, sortable map[string]string, defaultOrder string) string {
	column, ok := sortable[query["orderBy"]]
	if !ok || column == "" {
		return defaultOrder
	}
	direction := "ASC"
	if strings.EqualFold(query["orderType"], "desc") {
		direction = "DESC"
	}
	return column + " " + direction
}

func splitQueryValues(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// recordData only accepts JSON fields declared in the model; identifiers and audit fields are server-owned.
func recordData[T any](db *gorm.DB, data map[string]interface{}, create bool) (map[string]interface{}, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	for _, field := range stmt.Schema.Fields {
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		switch field.DBName {
		case "id", "create_time", "update_time", "delete_time", "created_by", "updated_by":
			continue
		}
		if key == "" || key == "-" || field.PrimaryKey || !field.Creatable || (!create && !field.Updatable) {
			continue
		}
		if value, ok := data[key]; ok {
			result[field.DBName] = value
		}
	}
	if create && stmt.Schema.LookUpField("create_time") != nil {
		result["create_time"] = gorm.Expr("NOW()")
	}
	if stmt.Schema.LookUpField("update_time") != nil {
		result["update_time"] = gorm.Expr("NOW()")
	}
	return result, nil
}

// CreateRecord inserts and reads its own generated primary key, using the supplied connection or transaction.
func CreateRecord[T any](db *gorm.DB, table string, data map[string]interface{}) (*T, error) {
	if db == nil {
		return nil, ErrNoDatabase
	}
	payload, err := recordData[T](db, data, true)
	if err != nil {
		return nil, err
	}
	if err := db.Model(new(T)).Table(table).Create(payload).Error; err != nil {
		return nil, err
	}
	id, ok := payload["id"]
	if !ok {
		return nil, fmt.Errorf("generated CRUD requires an auto-increment id")
	}
	var result T
	if err := db.Table(table).Where("id = ?", id).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRecord updates whitelisted model fields and returns the matching record.
func UpdateRecord[T any](db *gorm.DB, table, id string, data map[string]interface{}) (*T, error) {
	if db == nil {
		return nil, ErrNoDatabase
	}
	payload, err := recordData[T](db, data, false)
	if err != nil {
		return nil, err
	}
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return nil, err
	}
	base := db.Table(table).Where("id = ?", id)
	if stmt.Schema.LookUpField("delete_time") != nil {
		base = base.Where("delete_time IS NULL")
	}
	if len(payload) > 0 {
		if err := base.Updates(payload).Error; err != nil {
			return nil, err
		}
	}
	var result T
	if err := base.First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRecord hard-deletes a record in the supplied database or transaction.
func DeleteRecord(db *gorm.DB, model interface{}, id string) error {
	if db == nil {
		return ErrNoDatabase
	}
	return db.Delete(model, "id = ?", id).Error
}
