package store

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"
)

func TestFileStoreOverwriteAndReopen(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	s, err := Open(ctx, "", dir)
	if err != nil {
		t.Fatal(err)
	}
	exercise(t, s)
	s.Close()
	s, err = Open(ctx, "", dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var value map[string]int
	if err = s.Load(ctx, "test-state", &value); err != nil || value["revision"] != 2 {
		t.Fatal("reopen did not preserve latest state", err)
	}
	if err = s.Save(ctx, "../escape", value); err == nil {
		t.Fatal("path traversal accepted")
	}
}
func exercise(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()
	for _, n := range []int{1, 2} {
		if err := s.Save(ctx, "test-state", map[string]int{"revision": n}); err != nil {
			t.Fatal(err)
		}
	}
	keys, err := s.Keys(ctx)
	if err != nil || !slices.Contains(keys, "test-state") {
		t.Fatal("saved state absent", err)
	}
	var value map[string]int
	if err = s.Load(ctx, "test-state", &value); err != nil || value["revision"] != 2 {
		t.Fatal("overwrite failed", err)
	}
}
func TestPostgresStore(t *testing.T) {
	url := os.Getenv("FUNTIME_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("PostgreSQL integration runs in CI with FUNTIME_TEST_DATABASE_URL")
	}
	ctx := context.Background()
	s, err := Open(ctx, url, "")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	exercise(t, s)
	key := fmt.Sprintf("restart-%d", time.Now().UnixNano())
	if err = s.Save(ctx, key, map[string]string{"text": "Состояние игры"}); err != nil {
		t.Fatal(err)
	}
	other, err := Open(ctx, url, "")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	var value map[string]string
	if err = other.Load(ctx, key, &value); err != nil || value["text"] != "Состояние игры" {
		t.Fatal("PostgreSQL persistence failed", err)
	}
}
