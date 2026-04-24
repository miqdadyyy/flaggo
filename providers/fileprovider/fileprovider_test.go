package fileprovider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/miqdadyyy/flaggo"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "flags.json")
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNew_CreatesFile(t *testing.T) {
	path := tempPath(t)
	_, err := New(Options{Path: path})
	must(t, err)

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to be created: %v", err)
	}
}

func TestNew_LoadsExistingFile(t *testing.T) {
	path := tempPath(t)
	must(t, os.WriteFile(path, []byte(`{"flag1":{"enabled":true,"rollout":50}}`), 0o644))

	p, err := New(Options{Path: path})
	must(t, err)

	cfg, ok := p.Get(context.Background(), "flag1")
	if !ok {
		t.Fatal("expected flag1 to exist")
	}
	if !cfg.Enabled || cfg.Rollout != 50 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestSetAndGet(t *testing.T) {
	p, err := New(Options{Path: tempPath(t)})
	must(t, err)

	ctx := context.Background()
	want := flaggo.FlagConfig{
		Enabled: true,
		Rollout: 75,
		Actors:  map[string][]string{"user_id": {"1", "2"}},
	}
	must(t, p.Set(ctx, "flag1", want))

	got, ok := p.Get(ctx, "flag1")
	if !ok {
		t.Fatal("expected flag1 to exist")
	}
	if got.Enabled != want.Enabled || got.Rollout != want.Rollout {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if len(got.Actors["user_id"]) != 2 {
		t.Fatalf("expected 2 actors, got %d", len(got.Actors["user_id"]))
	}
}

func TestGet_NotFound(t *testing.T) {
	p, err := New(Options{Path: tempPath(t)})
	must(t, err)

	_, ok := p.Get(context.Background(), "nope")
	if ok {
		t.Fatal("expected flag not to exist")
	}
}

func TestAll(t *testing.T) {
	p, err := New(Options{Path: tempPath(t)})
	must(t, err)

	ctx := context.Background()
	must(t, p.Set(ctx, "a", flaggo.FlagConfig{Enabled: true}))
	must(t, p.Set(ctx, "b", flaggo.FlagConfig{Enabled: false}))

	all, err := p.All(ctx)
	must(t, err)
	if len(all) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(all))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	p, err := New(Options{Path: tempPath(t)})
	must(t, err)

	ctx := context.Background()
	must(t, p.Set(ctx, "a", flaggo.FlagConfig{Enabled: true}))

	all, err := p.All(ctx)
	must(t, err)
	all["b"] = flaggo.FlagConfig{Enabled: true}

	all2, err := p.All(ctx)
	must(t, err)
	if len(all2) != 1 {
		t.Fatal("modifying returned map should not affect provider")
	}
}

func TestPersistence(t *testing.T) {
	path := tempPath(t)
	ctx := context.Background()

	p1, err := New(Options{Path: path})
	must(t, err)
	must(t, p1.Set(ctx, "flag1", flaggo.FlagConfig{Enabled: true, Rollout: 60}))

	p2, err := New(Options{Path: path})
	must(t, err)

	cfg, ok := p2.Get(ctx, "flag1")
	if !ok {
		t.Fatal("expected flag1 to persist")
	}
	if !cfg.Enabled || cfg.Rollout != 60 {
		t.Fatalf("unexpected persisted config: %+v", cfg)
	}
}

func TestNew_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	path := filepath.Join(dir, "flags.json")

	_, err := New(Options{Path: path})
	must(t, err)

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to be created in nested dir: %v", err)
	}
}
