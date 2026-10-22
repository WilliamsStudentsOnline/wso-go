package mobile_fetcher

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
interval: 45s
output_dir: /tmp/mobile
jobs:
  - name: example
    url: https://example.com/data
    output: data.json
    headers:
      Authorization: "Bearer ${TEST_MOBILE_TOKEN}"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_MOBILE_TOKEN", "secret-token")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Interval != 45*time.Second {
		t.Fatalf("interval = %v, want 45s", cfg.Interval)
	}
	if cfg.OutputDir != "/tmp/mobile" {
		t.Fatalf("output_dir = %q", cfg.OutputDir)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("jobs len = %d", len(cfg.Jobs))
	}
	job := cfg.Jobs[0]
	if job.Name != "example" || job.URL != "https://example.com/data" || job.Output != "data.json" {
		t.Fatalf("unexpected job: %+v", job)
	}
	if got := job.Headers["Authorization"]; got != "Bearer secret-token" {
		t.Fatalf("Authorization = %q, want expanded token", got)
	}
}

func TestLoadConfigInvalidInterval(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
interval: not-a-duration
output_dir: /tmp/mobile
jobs:
  - name: example
    url: https://example.com
    output: out.json
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("expected error for invalid interval")
	}
}

func TestApplyOverrides(t *testing.T) {
	cfg := &Config{
		Interval:  30 * time.Second,
		IntervalS: "30s",
		OutputDir: "/old",
		Jobs:      []Job{{Name: "a", URL: "http://x", Output: "a.json"}},
	}
	cfg.ApplyOverrides(Overrides{
		OutputDir: "/new",
		Interval:  10 * time.Second,
	})
	if cfg.OutputDir != "/new" {
		t.Fatalf("output_dir = %q", cfg.OutputDir)
	}
	if cfg.Interval != 10*time.Second {
		t.Fatalf("interval = %v", cfg.Interval)
	}

	cfg.ApplyOverrides(Overrides{})
	if cfg.OutputDir != "/new" || cfg.Interval != 10*time.Second {
		t.Fatal("empty overrides should leave config unchanged")
	}
}

func TestLoadBundledConfig(t *testing.T) {
	t.Setenv("SPINITRON_TOKEN", "test-token")

	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig bundled: %v", err)
	}
	if cfg.Interval != 30*time.Second {
		t.Fatalf("interval = %v", cfg.Interval)
	}
	if cfg.Owner != "wso" || cfg.Group != "deploy" {
		t.Fatalf("owner/group = %q/%q, want wso/deploy", cfg.Owner, cfg.Group)
	}
	if len(cfg.Jobs) != 6 {
		t.Fatalf("expected 6 jobs, got %d", len(cfg.Jobs))
	}
	for _, job := range cfg.Jobs {
		if auth, ok := job.Headers["Authorization"]; ok {
			if auth != "Bearer test-token" {
				t.Fatalf("job %s Authorization = %q", job.Name, auth)
			}
		}
	}
}

func TestResolveOwnershipEmpty(t *testing.T) {
	cfg := &Config{}
	own, err := cfg.ResolveOwnership()
	if err != nil {
		t.Fatal(err)
	}
	if own.UID != -1 || own.GID != -1 {
		t.Fatalf("empty ownership = %+v, want uid/gid -1", own)
	}
}

func TestResolveOwnershipCurrentUser(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	g, err := user.LookupGroupId(u.Gid)
	if err != nil {
		t.Fatal(err)
	}
	wantUID, err := strconv.Atoi(u.Uid)
	if err != nil {
		t.Fatal(err)
	}
	wantGID, err := strconv.Atoi(g.Gid)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Owner: u.Username, Group: g.Name}
	own, err := cfg.ResolveOwnership()
	if err != nil {
		t.Fatal(err)
	}
	if own.UID != wantUID || own.GID != wantGID {
		t.Fatalf("ownership = %+v, want uid=%d gid=%d", own, wantUID, wantGID)
	}
}

func TestResolveOwnershipUnknown(t *testing.T) {
	cfg := &Config{Owner: "definitely-not-a-real-user-zz"}
	if _, err := cfg.ResolveOwnership(); err == nil {
		t.Fatal("expected error for unknown owner")
	}
	cfg = &Config{Group: "definitely-not-a-real-group-zz"}
	if _, err := cfg.ResolveOwnership(); err == nil {
		t.Fatal("expected error for unknown group")
	}
}
