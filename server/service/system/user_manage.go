package system

import (
	"fmt"
	"strings"

	commonResponse "server/model/common/response"
	systemModel "server/model/system"
	systemRequest "server/model/system/request"

	"gorm.io/gorm"
)

// 本文件是"用户管理"的业务逻辑（后台管理员对用户的增删改查），
// 与 user.go（登录/当前用户上下文，属于"认证"）职责不同。

// UserList 分页查询用户列表。
// 查询条件分两类：likes（模糊匹配，生成 LIKE '%xx%'）和 equals（精确匹配）；
// map 的 key 是前端参数名（camelCase），value 是数据库列名（snake_case）。
// operatorID 是当前操作者，用于套数据权限（dataScope）：非全部权限的角色只能看到授权部门内的用户。
func UserList(operatorID uint, query map[string]string) (*commonResponse.PageResult, error) {
	db, err := systemDB()
	if err != nil {
		return nil, err
	}

	scope, err := UserDataScope(operatorID)
	if err != nil {
		return nil, err
	}

	page := parsePage(query)
	// softDelete 统一追加 delete_time IS NULL，过滤掉已软删除的数据。
	base := softDelete(db.Model(&systemModel.AISystemUser{}))
	base = applyUserDataScope(base, scope, operatorID)
	base = applyFilters(base, query,
		map[string]string{"username": "username", "nickname": "nickname", "phone": "phone", "email": "email"},
		map[string]string{"status": "status", "deptId": "dept_id"},
	)

	// 分页的标准套路：先 Count 拿总数，再 Offset/Limit 取当前页数据。
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []systemModel.AISystemUser
	if err := base.Order("id ASC").Offset((page.Page - 1) * page.Size).Limit(page.Size).Find(&users).Error; err != nil {
		return nil, err
	}

	// 逐个用户补查角色 ID（前端列表要显示角色）。
	// 注意：这是典型的 N+1 查询写法，数据量大时应改为一次 IN 查询后按 user_id 分组。
	for i := range users {
		users[i].Roles, _ = RoleIDsByUserID(users[i].ID)
	}

	return &commonResponse.PageResult{List: users, Total: total}, nil
}

// CreateUser atomically creates a user and role bindings within the operator's department scope.
func CreateUser(operatorID uint, payload systemRequest.UserPayload) (*systemModel.AISystemUser, error) {
	db, err := systemDB()
	if err != nil {
		return nil, err
	}
	if payload.Password == nil || *payload.Password == "" {
		return nil, ErrPasswordRequired
	}
	if payload.Username == nil || strings.TrimSpace(*payload.Username) == "" {
		return nil, NewBizError("用户名不能为空")
	}
	hash, err := hashPassword(*payload.Password)
	if err != nil {
		return nil, err
	}
	var user systemModel.AISystemUser
	err = db.Transaction(func(tx *gorm.DB) error {
		scope, err := userDataScope(tx, operatorID)
		if err != nil {
			return err
		}
		if err := authorizeDepartment(scope, payload.DeptID); err != nil {
			return err
		}
		if err := authorizeRoles(tx, operatorID, payload.Roles); err != nil {
			return err
		}
		data := userPayloadData(payload)
		data["password"] = hash
		setDefaultTimes(data, true)
		if err := tx.Model(&systemModel.AISystemUser{}).Create(data).Error; err != nil {
			return err
		}
		id, ok := data["id"]
		if !ok {
			return fmt.Errorf("user insert did not return primary key")
		}
		if err := tx.Where("id = ?", id).First(&user).Error; err != nil {
			return err
		}
		return bindUserRoles(tx, user.ID, payload.Roles)
	})
	if err != nil {
		return nil, err
	}
	user.Roles = payload.Roles
	return &user, nil
}

// UpdateUser checks the existing target and destination department before any write.
func UpdateUser(operatorID uint, id string, payload systemRequest.UserPayload) (*systemModel.AISystemUser, error) {
	db, err := systemDB()
	if err != nil {
		return nil, err
	}
	var result systemModel.AISystemUser
	err = withUserWrite(db, operatorID, id, func(tx *gorm.DB, scope *DataScope, user *systemModel.AISystemUser) error {
		if payload.DeptID != nil && (user.DeptID == nil || *payload.DeptID != *user.DeptID) {
			if err := authorizeDepartment(scope, payload.DeptID); err != nil {
				return err
			}
		}
		if payload.Roles != nil {
			if err := authorizeRoles(tx, operatorID, payload.Roles); err != nil {
				return err
			}
		}
		data := userPayloadData(payload)
		setDefaultTimes(data, false)
		if err := tx.Model(user).Updates(data).Error; err != nil {
			return err
		}
		if payload.Roles != nil {
			if err := bindUserRoles(tx, user.ID, payload.Roles); err != nil {
				return err
			}
		}
		if err := tx.First(&result, user.ID).Error; err != nil {
			return err
		}
		var roles []systemModel.AISystemUserRole
		if err := tx.Where("user_id = ?", user.ID).Find(&roles).Error; err != nil {
			return err
		}
		result.Roles = make([]uint, 0, len(roles))
		for _, role := range roles {
			result.Roles = append(result.Roles, role.RoleID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteUser deletes the scoped target and its role bindings in one transaction.
func DeleteUser(operatorID uint, id string) error {
	db, err := systemDB()
	if err != nil {
		return err
	}
	return withUserWrite(db, operatorID, id, func(tx *gorm.DB, _ *DataScope, user *systemModel.AISystemUser) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&systemModel.AISystemUserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(user).Error
	})
}

// SetUserPassword resets a password only after locking an authorized target.
func SetUserPassword(operatorID uint, id, password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	db, err := systemDB()
	if err != nil {
		return err
	}
	return withUserWrite(db, operatorID, id, func(tx *gorm.DB, _ *DataScope, user *systemModel.AISystemUser) error {
		return tx.Model(user).Updates(map[string]interface{}{"password": hash, "update_time": gorm.Expr("NOW()")}).Error
	})
}

// UserAuthList 返回启用用户的 {label, value} 下拉选项，label 优先取昵称。
func UserAuthList() ([]map[string]interface{}, error) {
	db, err := systemDB()
	if err != nil {
		return nil, err
	}
	var users []systemModel.AISystemUser
	if err := softDelete(db).Where("status = ?", 1).Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		label := user.Username
		// Nickname 是 *string（数据库可空列），解引用前必须判 nil，否则会 panic。
		if user.Nickname != nil && *user.Nickname != "" {
			label = *user.Nickname
		}
		result = append(result, map[string]interface{}{"label": label, "value": user.ID})
	}
	return result, nil
}

// BindUserRolesStringID is the authenticated administrative role assignment entry point.
func BindUserRolesStringID(operatorID uint, userID string, roleIDs []uint) error {
	db, err := systemDB()
	if err != nil {
		return err
	}
	return withUserWrite(db, operatorID, userID, func(tx *gorm.DB, _ *DataScope, user *systemModel.AISystemUser) error {
		if err := authorizeRoles(tx, operatorID, roleIDs); err != nil {
			return err
		}
		return bindUserRoles(tx, user.ID, roleIDs)
	})
}

// bindUserRoles participates in the caller's transaction; it is never an authorization entry point.
func bindUserRoles(tx *gorm.DB, userID uint, roleIDs []uint) error {
	if err := tx.Where("user_id = ?", userID).Delete(&systemModel.AISystemUserRole{}).Error; err != nil {
		return err
	}
	seen := map[uint]bool{}
	for _, roleID := range roleIDs {
		if seen[roleID] {
			continue
		}
		seen[roleID] = true
		if err := tx.Create(&systemModel.AISystemUserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// RoleIDsByUserID 查询用户当前绑定的角色 ID 列表（走 user-role 中间表）。
func RoleIDsByUserID(userID uint) ([]uint, error) {
	db, err := systemDB()
	if err != nil {
		return nil, err
	}
	var rels []systemModel.AISystemUserRole
	if err := db.Where("user_id = ?", userID).Find(&rels).Error; err != nil {
		return nil, err
	}
	result := make([]uint, 0, len(rels))
	for _, rel := range rels {
		result = append(result, rel.RoleID)
	}
	return result, nil
}

// userPayloadData 把类型化入参转成 GORM 更新 map，只收集非 nil 字段（部分更新）。
func userPayloadData(payload systemRequest.UserPayload) map[string]interface{} {
	data := map[string]interface{}{}
	setColumn(data, "username", payload.Username)
	setColumn(data, "user_type", payload.UserType)
	setColumn(data, "nickname", payload.Nickname)
	setColumn(data, "phone", payload.Phone)
	setColumn(data, "email", payload.Email)
	setColumn(data, "avatar", payload.Avatar)
	setColumn(data, "signed", payload.Signed)
	setColumn(data, "dashboard", payload.Dashboard)
	setColumn(data, "dept_id", payload.DeptID)
	setColumn(data, "status", payload.Status)
	setColumn(data, "remark", payload.Remark)
	setColumn(data, "backend_setting", payload.BackendSetting)
	return data
}
