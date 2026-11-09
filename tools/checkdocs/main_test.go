package main

import (
	"fmt"
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

func TestCheckMarkdownCodeSpans(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "target.md", "# Target\n")
	tests := []struct {
		name        string
		content     string
		links       int
		problemLine int
	}{
		{"generic calls", "Use `first[int](nil)` and `number.Map[string](...)`. [real](target.md)", 1, 0},
		{"multiple backticks", "Use ``[sample](ignored.md) with ` inside`` [real](target.md)", 1, 0},
		{"different inner runs", "Use `text `` [sample](ignored.md) `` end` [real](target.md)", 1, 0},
		{"unmatched opener", "Use `literal [broken](missing.md)", 1, 1},
		{"unequal unmatched runs", "Use `literal `` [broken](missing.md)", 1, 1},
		{"matched after unmatched", "Use `literal ``[sample](ignored.md)`` [real](target.md)", 1, 0},
		{"escaped openers", "Use \\`[broken](missing.md)\\`", 1, 1},
		{"escaped backslash", "Use \\\\`[sample](ignored.md)` [real](target.md)", 1, 0},
		{"backslash inside code", "Use `text\\` [broken](missing.md)", 1, 1},
		{"code in link label", "[`api[int]`](target.md)", 1, 0},
		{"multiline span", "Start `first[int]\n(nil) [sample](ignored.md)\nend` [broken](missing.md)", 1, 3},
		{"blank line boundary", "Start `unclosed\n\n[broken](missing.md) `later", 1, 3},
		{"fence boundary", "Start `unclosed\n```\n[sample](ignored.md)\n```\n[broken](missing.md) `later", 1, 5},
		{"triple inline span", "```[sample](ignored.md)``` [broken](missing.md)", 1, 1},
		{"tilde is not inline code", "~~[broken](missing.md)~~", 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixture(t, root, "README.md", tt.content)
			links, problems := checkMarkdown(path)
			if links != tt.links {
				t.Fatalf("links=%d want=%d, problems=%v", links, tt.links, problems)
			}
			if tt.problemLine == 0 {
				if len(problems) != 0 {
					t.Fatalf("unexpected problems=%v", problems)
				}
			} else if len(problems) != 1 || !strings.Contains(problems[0], fmt.Sprintf(":%d: missing.md:", tt.problemLine)) {
				t.Fatalf("problems=%v, want missing.md on line %d", problems, tt.problemLine)
			}
		})
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
