package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckMarkdown(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "target.md", "# Target\n")
	writeFixture(t, root, "image.svg", "<svg/>\n")
	writeFixture(t, root, "with space.md", "# Target\n")
	content := strings.Join([]string{
		"[existing](target.md#not-checked)",
		"![image](image.svg)",
		"[space](with%20space.md)",
		"[web](https://example.com/missing.md)",
		"[anchor](#not-checked)",
		"````markdown",
		"[sample](ignored-in-fence.md)",
		"```",
		"[sample](still-ignored.md)",
		"````",
		"~~~text",
		"[sample](ignored-in-tilde.md)",
		"~~~",
		"[broken](missing.md)",
	}, "\n")
	path := writeFixture(t, root, "README.md", content)
	links, problems := checkMarkdown(path)
	if links != 4 {
		t.Fatalf("checked %d local links, want 4", links)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], ":14: missing.md:") {
		t.Fatalf("problems = %v, want the missing link on line 14", problems)
	}
}

func TestCheckTree(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "README.md", "[chapter](book/chapter.md)\n")
	writeFixture(t, root, "book/chapter.md", "[home](../README.md)\n")
	writeFixture(t, root, "_practice/notes.md", "[broken](missing.md)\n")
	writeFixture(t, root, ".git/notes.md", "[broken](missing.md)\n")
	writeFixture(t, root, ".github/pull_request_template.md", "[home](../README.md)\n")
	documents, links, problems := checkTree(root)
	if documents != 3 || links != 3 || len(problems) != 0 {
		t.Fatalf("got documents=%d links=%d problems=%v; want 3, 3, none", documents, links, problems)
	}
}

func TestCheckTarget(t *testing.T) {
	root := t.TempDir()
	document := writeFixture(t, root, "README.md", "# Test\n")
	tests := []struct {
		target  string
		local   bool
		wantErr bool
	}{
		{target: "README.md", local: true},
		{target: "missing.md", local: true, wantErr: true},
		{target: "/README.md", local: true, wantErr: true},
		{target: "https://go.dev/doc/"},
		{target: "//example.com/doc"},
		{target: "mailto:reader@example.com"},
		{target: "#heading"},
		{target: "%ZZ.md", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			local, err := checkTarget(document, tt.target)
			if local != tt.local || (err != nil) != tt.wantErr {
				t.Errorf("checkTarget(%q) = (%v, %v); want local=%v, error=%v", tt.target, local, err, tt.local, tt.wantErr)
			}
		})
	}
}

func writeFixture(t *testing.T, root, name, contents string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
