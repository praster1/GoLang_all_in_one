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
	var paragraph []string
	paragraphLine := 0
	checkParagraph := func() {
		// A code span may continue across a soft line break. Mask it before
		// checking individual lines, preserving diagnostic line numbers.
		text := maskCodeSpans(strings.Join(paragraph, "\n"))
		for offset, line := range strings.Split(text, "\n") {
			for _, match := range inlineLink.FindAllStringSubmatch(line, -1) {
				local, err := checkTarget(path, match[1])
				if local {
					links++
				}
				if err != nil {
					problems = append(problems, fmt.Sprintf("%s:%d: %s: %v", path, paragraphLine+offset, match[1], err))
				}
			}
		}
		paragraph = nil
	}
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		marker, length := fenceMarker(trimmed)
		if fence != 0 {
			if marker == fence && length >= fenceLength && strings.TrimSpace(trimmed[length:]) == "" {
				fence, fenceLength = 0, 0
			}
			continue
		}
		// Backtick fence info strings cannot contain backticks. In particular,
		// a line beginning with a complete triple-backtick span is not a fence.
		if length >= 3 && (marker != '`' || !strings.ContainsRune(trimmed[length:], '`')) {
			checkParagraph()
			fence, fenceLength = marker, length
			continue
		}
		if trimmed == "" {
			checkParagraph()
			continue
		}
		if len(paragraph) == 0 {
			paragraphLine = lineNumber
		}
		paragraph = append(paragraph, line)
	}
	checkParagraph()
	if err := scanner.Err(); err != nil {
		problems = append(problems, fmt.Sprintf("%s: %v", path, err))
	}
	return links, problems
}

// maskCodeSpans hides matched CommonMark backtick spans while retaining newlines.
// Unequal runs cannot close a span, and an unmatched opener remains literal.
func maskCodeSpans(text string) string {
	masked := []byte(text)
	for start := 0; start < len(text); {
		if text[start] == '\\' && start+1 < len(text) {
			start += 2 // An escaped backtick cannot open a span.
			continue
		}
		if text[start] != '`' {
			start++
			continue
		}
		_, openingLength := fenceMarker(text[start:])
		openingEnd := start + openingLength
		closingEnd := -1
		for search := openingEnd; search < len(text); {
			next := strings.IndexByte(text[search:], '`')
			if next < 0 {
				break
			}
			closingStart := search + next
			_, length := fenceMarker(text[closingStart:])
			search = closingStart + length
			// Inside a span, even a backslash before the closing run is literal.
			if length == openingLength {
				closingEnd = search
				break
			}
		}
		if closingEnd < 0 {
			start = openingEnd
			continue
		}
		for i := start; i < closingEnd; i++ {
			if masked[i] != '\n' {
				masked[i] = ' '
			}
		}
		start = closingEnd
	}
	return string(masked)
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
