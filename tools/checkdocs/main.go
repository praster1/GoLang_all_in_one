// Command checkdocs checks local inline Markdown links from the repository root.
package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var inlineLink = regexp.MustCompile(`\[[^\]\n]*\]\(([^\s)]+)(?:\s+"[^"]*")?\)`)

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		fmt.Fprintln(os.Stderr, "Run go run ./tools/checkdocs from the repository root.")
		os.Exit(1)
	}
	documents, links, problems := checkTree(root)
	for _, problem := range problems {
		fmt.Fprintln(os.Stderr, problem)
	}
	if len(problems) > 0 {
		os.Exit(1)
	}
	fmt.Printf("Checked %d Markdown files and %d local links.\n", documents, links)
}

func checkTree(root string) (documents, links int, problems []string) {
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "_practice", "bin", "dist", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		documents++
		count, errors := checkMarkdown(path)
		links += count
		problems = append(problems, errors...)
		return nil
	})
	if err != nil {
		problems = append(problems, err.Error())
	}
	return documents, links, problems
}

func checkMarkdown(path string) (links int, problems []string) {
	file, err := os.Open(path)
	if err != nil {
		return 0, []string{err.Error()}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	var fence byte
	fenceLength := 0
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if marker, length := fenceMarker(trimmed); length >= 3 {
			if fence == 0 {
				fence, fenceLength = marker, length
			} else if marker == fence && length >= fenceLength && strings.TrimSpace(trimmed[length:]) == "" {
				fence, fenceLength = 0, 0
			}
			continue
		}
		if fence != 0 {
			continue
		}
		for _, match := range inlineLink.FindAllStringSubmatch(line, -1) {
			local, err := checkTarget(path, match[1])
			if local {
				links++
			}
			if err != nil {
				problems = append(problems, fmt.Sprintf("%s:%d: %s: %v", path, lineNumber, match[1], err))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		problems = append(problems, fmt.Sprintf("%s: %v", path, err))
	}
	return links, problems
}

func fenceMarker(line string) (byte, int) {
	if len(line) == 0 || (line[0] != '`' && line[0] != '~') {
		return 0, 0
	}
	length := 0
	for length < len(line) && line[length] == line[0] {
		length++
	}
	return line[0], length
}

func checkTarget(markdownPath, target string) (local bool, err error) {
	u, err := url.Parse(strings.Trim(target, "<>"))
	if err != nil {
		return false, fmt.Errorf("invalid link: %w", err)
	}
	if u.IsAbs() || u.Host != "" || u.Path == "" {
		return false, nil
	}
	if strings.HasPrefix(u.Path, "/") {
		return true, fmt.Errorf("use a path relative to the Markdown file")
	}
	path := filepath.Join(filepath.Dir(markdownPath), filepath.FromSlash(u.Path))
	if _, err := os.Stat(path); err != nil {
		return true, err
	}
	return true, nil
}
