package gormInit

import (
	"testing"

	"gorm.io/gorm"
)

func TestGetDoesNotFallBackToSystemDatabase(t *testing.T) {
	systemDB, legacyDB := &gorm.DB{}, &gorm.DB{}
	registry := &_gorm{Databases: &Databases{AISystem: systemDB, Named: map[string]*gorm.DB{"legacy": legacyDB}}}
	if db, err := registry.Get("legacy"); err != nil || db != legacyDB {
		t.Fatalf("db=%v err=%v", db, err)
	}
	if _, err := registry.Get("typo"); err == nil {
		t.Fatal("unknown connection silently accepted")
	}
	if _, err := new(_gorm).Get("ai_system"); err == nil {
		t.Fatal("uninitialized database accepted")
	}
}
