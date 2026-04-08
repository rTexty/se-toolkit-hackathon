package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateNonExistentDir(t *testing.T) {
	err := Migrate(context.Background(), nil, "/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestMigrateDirWithSubdirs(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	err := Migrate(context.Background(), nil, dir)
	if err != nil {
		t.Fatalf("Migrate should skip subdirs: %v", err)
	}
}
