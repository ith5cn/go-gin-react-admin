package system

import (
	"errors"

	systemModel "server/model/system"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// withUserWrite scopes and locks the target before any mutation, using the same transaction.
func withUserWrite(db *gorm.DB, operatorID uint, id string, mutate func(*gorm.DB, *DataScope, *systemModel.AISystemUser) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		scope, err := userDataScope(tx, operatorID)
		if err != nil {
			return err
		}
		var target systemModel.AISystemUser
		err = applyUserDataScope(softDelete(tx.Model(&target)), scope, operatorID).
			Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&target).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDataScopeDenied
		}
		if err != nil {
			return err
		}
		return mutate(tx, scope, &target)
	})
}

func authorizeDepartment(scope *DataScope, deptID *uint) error {
	if scope != nil && scope.All {
		return nil
	}
	if scope != nil && deptID != nil && *deptID != 0 {
		for _, allowed := range scope.DeptIDs {
			if allowed == *deptID {
				return nil
			}
		}
	}
	return ErrDataScopeDenied
}

// Non-super administrators may delegate only their own enabled roles.
// Having an "all departments" data scope does not imply role administration authority.
func authorizeRoles(tx *gorm.DB, operatorID uint, roleIDs []uint) error {
	if len(roleIDs) == 0 {
		return nil
	}
	var owned []systemModel.AISystemRole
	if err := tx.Table("ai_system_role AS r").Select("r.id, r.code").
		Joins("JOIN ai_system_user_role AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = ? AND r.delete_time IS NULL", operatorID, 1).Scan(&owned).Error; err != nil {
		return err
	}
	allowed := map[uint]bool{}
	for _, role := range owned {
		if isSuperAdminRole(role) {
			return nil
		}
		allowed[role.ID] = true
	}
	for _, id := range roleIDs {
		if !allowed[id] {
			return ErrRoleAssignmentDenied
		}
	}
	return nil
}
