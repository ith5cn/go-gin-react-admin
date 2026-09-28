package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"server/internal/testutil"
	gormInit "server/setup/gorm"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestPermissionRejectsMissingIdentityAndUnauthorizedRole(t *testing.T) {
	for _, identified := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing identity", true: "no roles"}[identified], func(t *testing.T) {
			db, mock := testutil.DB(t)
			previous := gormInit.Gorm.Databases
			gormInit.Gorm.Databases = &gormInit.Databases{AISystem: db}
			t.Cleanup(func() { gormInit.Gorm.Databases = previous })
			r := gin.New()
			if identified {
				r.Use(func(c *gin.Context) { c.Set("user_id", uint(7)); c.Next() })
				mock.ExpectQuery("SELECT r.id, r.code").WithArgs(uint(7), 1).WillReturnRows(sqlmock.NewRows([]string{"id", "code"}))
			}
			r.GET("/protected", Perm("system/user/update"), func(c *gin.Context) { t.Error("unauthorized handler ran") })
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/protected", nil))
			want := http.StatusUnauthorized
			if identified {
				want = http.StatusForbidden
			}
			if recorder.Code != want {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
