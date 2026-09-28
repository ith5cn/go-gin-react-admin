package system

import (
	"errors"
	"testing"

	"server/internal/testutil"
	systemModel "server/model/system"
	systemRequest "server/model/system/request"
	gormInit "server/setup/gorm"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func userTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock := testutil.DB(t)
	previous := gormInit.Gorm.Databases
	gormInit.Gorm.Databases = &gormInit.Databases{AISystem: db}
	t.Cleanup(func() { gormInit.Gorm.Databases = previous })
	return db, mock
}

func expectSelfScope(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT r.id, r.code, r.data_scope").WithArgs(uint(7), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "data_scope"}).AddRow(2, "member", 5))
}

func TestUserWritesDenyOutOfScopeBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"update", func() error { _, err := UpdateUser(7, "8", systemRequest.UserPayload{}); return err }},
		{"delete", func() error { return DeleteUser(7, "8") }},
		{"password", func() error { return SetUserPassword(7, "8", "new-password") }},
		{"roles", func() error { return BindUserRolesStringID(7, "8", []uint{1}) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, mock := userTestDB(t)
			mock.ExpectBegin()
			expectSelfScope(mock)
			mock.ExpectQuery("SELECT .*ai_system_user.*delete_time IS NULL.*id = \\?.*id = \\?.*FOR UPDATE").WithArgs(uint(7), "8", 1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "dept_id"}))
			mock.ExpectRollback()
			if err := tc.run(); !errors.Is(err, ErrDataScopeDenied) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestDepartmentAuthorization(t *testing.T) {
	allowed, other, zero := uint(10), uint(20), uint(0)
	for _, tc := range []struct {
		name  string
		scope *DataScope
		dept  *uint
		want  bool
	}{
		{"nil scope", nil, &allowed, false}, {"all", &DataScope{All: true}, nil, true},
		{"owned", &DataScope{DeptIDs: []uint{10}}, &allowed, true},
		{"other", &DataScope{DeptIDs: []uint{10}}, &other, false},
		{"self", &DataScope{SelfOnly: true}, &allowed, false},
		{"unassigned", &DataScope{DeptIDs: []uint{10}}, nil, false},
		{"zero", &DataScope{DeptIDs: []uint{10}}, &zero, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := authorizeDepartment(tc.scope, tc.dept) == nil; got != tc.want {
				t.Fatalf("allowed = %v", got)
			}
		})
	}
}

func TestRoleBindingFailureRollsBackUserUpdate(t *testing.T) {
	_, mock := userTestDB(t)
	mock.ExpectBegin()
	expectSelfScope(mock)
	mock.ExpectQuery("SELECT .*ai_system_user.*FOR UPDATE").WithArgs(uint(7), "7", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dept_id"}).AddRow(7, 10))
	mock.ExpectQuery("SELECT r.id, r.code").WithArgs(uint(7), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(2, "member"))
	mock.ExpectExec("UPDATE `ai_system_user`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM `ai_system_user_role`").WithArgs(uint(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	failure := errors.New("role insert failed")
	mock.ExpectExec("INSERT INTO `ai_system_user_role`").WillReturnError(failure)
	mock.ExpectRollback()
	nickname := "changed"
	_, err := UpdateUser(7, "7", systemRequest.UserPayload{Nickname: &nickname, Roles: []uint{2}})
	if !errors.Is(err, failure) {
		t.Fatalf("error = %v", err)
	}
}

func TestScopePredicateIncludesSelfAndGroupsDepartments(t *testing.T) {
	db, _ := testutil.DB(t)
	stmt := applyUserDataScope(db.Session(&gorm.Session{DryRun: true}).Model(&systemModel.AISystemUser{}), &DataScope{DeptIDs: []uint{10}}, 7).
		Where("id = ?", 8).Find(&[]systemModel.AISystemUser{}).Statement
	want := "SELECT * FROM `ai_system_user` WHERE (dept_id IN (?) OR id = ?) AND id = ?"
	if stmt.SQL.String() != want {
		t.Fatalf("unsafe predicate: %s", stmt.SQL.String())
	}
}

func TestUpdateCannotMoveSelfOutsideAuthorizedDepartments(t *testing.T) {
	_, mock := userTestDB(t)
	mock.ExpectBegin()
	expectSelfScope(mock)
	mock.ExpectQuery("SELECT .*ai_system_user.*FOR UPDATE").WithArgs(uint(7), "7", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dept_id"}).AddRow(7, 10))
	mock.ExpectRollback()
	department := uint(20)
	_, err := UpdateUser(7, "7", systemRequest.UserPayload{DeptID: &department})
	if !errors.Is(err, ErrDataScopeDenied) {
		t.Fatalf("error=%v", err)
	}
}

func TestRoleAssignmentRejectsUnownedRoles(t *testing.T) {
	db, mock := testutil.DB(t)
	mock.ExpectQuery("SELECT r.id, r.code").WithArgs(uint(7), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(2, "member"))
	if err := authorizeRoles(db, 7, []uint{1}); !errors.Is(err, ErrRoleAssignmentDenied) {
		t.Fatalf("error=%v", err)
	}
}

func TestScopedTargetCanBeDeletedAtomically(t *testing.T) {
	_, mock := userTestDB(t)
	mock.ExpectBegin()
	expectSelfScope(mock)
	mock.ExpectQuery("SELECT .*ai_system_user.*FOR UPDATE").WithArgs(uint(7), "7", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "dept_id"}).AddRow(7, 10))
	mock.ExpectExec("DELETE FROM `ai_system_user_role`").WithArgs(uint(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM `ai_system_user`").WithArgs(uint(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := DeleteUser(7, "7"); err != nil {
		t.Fatal(err)
	}
}
