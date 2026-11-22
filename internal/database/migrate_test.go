package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
	"testing"
)

func legacyDB(t *testing.T, orphan bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	statements := []string{
		`CREATE TABLE categories (id integer PRIMARY KEY, name text, description text, created_at datetime, updated_at datetime, deleted_at datetime)`,
		`CREATE TABLE questions (id integer PRIMARY KEY, category_id integer, title text, content text, image text, explanation text, source_url text, published numeric NOT NULL DEFAULT false, is_critical numeric, position integer NOT NULL, license_type text, created_at datetime, updated_at datetime, deleted_at datetime, CONSTRAINT fk_questions_category FOREIGN KEY(category_id) REFERENCES categories(id))`,
		`CREATE TABLE answers (id integer PRIMARY KEY, question_id integer, content text, label text, "order" integer, is_correct numeric, created_at datetime, updated_at datetime, deleted_at datetime, CONSTRAINT fk_questions_answers FOREIGN KEY(question_id) REFERENCES questions(id))`,
		`INSERT INTO questions(id,category_id,title,content,position,license_type) VALUES(10,1,'Old question','Old content',1,'A1')`,
		`INSERT INTO answers(id,question_id,content,label,"order",is_correct) VALUES(20,10,'Old answer','A',1,1)`,
	}
	for _, sql := range statements {
		if err = db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	if !orphan {
		if err = db.Exec(`INSERT INTO categories(id,name) VALUES(1,'Legacy category')`).Error; err != nil {
			t.Fatal(err)
		}
	}
	conn, _ := db.DB()
	if err = conn.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestMigrationPreservesReferencedQuestions(t *testing.T) {
	path := legacyDB(t, false)
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	defer conn.Close()
	var typeCode string
	if err = db.Raw("SELECT type FROM categories WHERE id=1").Scan(&typeCode).Error; err != nil {
		t.Fatal(err)
	}
	if typeCode != "UNCATEGORIZED" {
		t.Fatal("legacy type was guessed")
	}
	var content string
	if err = db.Raw("SELECT content FROM answers WHERE id=20 AND question_id=10").Scan(&content).Error; err != nil {
		t.Fatal(err)
	}
	if content != "Old answer" {
		t.Fatal("answer ID/link/content changed")
	}
	var fk int
	if err = db.Raw("PRAGMA foreign_keys").Scan(&fk).Error; err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatal("foreign keys not restored")
	}
	if !db.Migrator().HasColumn("questions", "road_sign_id") {
		t.Fatal("missing new column")
	}
	if err = autoMigrate(db); err != nil {
		t.Fatalf("second migration not idempotent: %v", err)
	}
}
func TestMigrationRollsBackInvalidReferences(t *testing.T) {
	path := legacyDB(t, true)
	if _, err := InitDB(path); err == nil {
		t.Fatal("migration accepted invalid foreign keys")
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	defer conn.Close()
	if db.Migrator().HasColumn("questions", "road_sign_id") {
		t.Fatal("failed migration left schema changes")
	}
	var count int64
	if err = db.Table("answers").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("failed migration lost old data")
	}
}
