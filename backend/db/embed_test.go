package dbmigrations

import (
	"io/fs"
	"sort"
	"testing"
)

func TestEmbed_contains_system_config_migration(t *testing.T) {
	entries, err := fs.Glob(FS, "migrations/0012_system_config.up.sql")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("0012_system_config.up.sql not found in embedded FS")
	}
}

func TestEmbed_0012_sorts_after_0011(t *testing.T) {
	entries, err := fs.Glob(FS, "migrations/*.up.sql")
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	sort.Strings(entries)

	var found bool
	for _, e := range entries {
		if e == "migrations/0011_profile_photo.up.sql" {
			found = true
		}
		if e == "migrations/0012_system_config.up.sql" {
			if !found {
				t.Fatal("0012_system_config.up.sql appears before 0011_profile_photo.up.sql in sorted order")
			}
			return
		}
	}
	t.Fatal("0012_system_config.up.sql not found in sorted migration list after 0011")
}
