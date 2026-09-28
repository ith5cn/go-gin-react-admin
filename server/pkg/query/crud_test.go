package query

import (
	"testing"
	"time"

	"server/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
)

type testArticle struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	Title      string     `json:"title"`
	CreateTime *time.Time `json:"createTime"`
	DeleteTime *time.Time `json:"-"`
}

func TestCreateUsesInsertedIDAndWhitelistedFields(t *testing.T) {
	db, mock := testutil.DB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `articles` .*`create_time`,`title`").WithArgs("ours").WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT .*articles.*WHERE id = \\?").WithArgs(int64(42), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(42, "ours"))
	result, err := CreateRecord[testArticle](db, "articles", map[string]interface{}{"title": "ours", "id": 999, "deleteTime": "yesterday", "unknown": "discard"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 42 {
		t.Fatalf("wrong record: %#v", result)
	}
}

func TestPaginationAndSortWhitelist(t *testing.T) {
	page := ParsePage(map[string]string{"page": "-1", "size": "99999999"})
	if page.Page != 1 || page.Size != 1000 {
		t.Fatalf("page=%#v", page)
	}
	if got := resolveOrder(map[string]string{"orderBy": "id; DROP TABLE users", "orderType": "desc"}, map[string]string{"title": "title"}, "id DESC"); got != "id DESC" {
		t.Fatal(got)
	}
}
