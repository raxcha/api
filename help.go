package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func readPage(root, relpath string, depth int) (*Page, error) {

	abspath := filepath.Join(root, filepath.Clean(relpath))
	info, err := os.Stat(abspath)
	if err != nil { return nil, err }

	p := &Page{
		Name: filepath.Base(relpath),
		Path: relpath,
	}

	if info.IsDir() {
		p.Type = "deep"
		p.Content = readLines(filepath.Join(abspath, "index"))
		p.Metadata = parseMetadata(p.Content)

		if depth != 0 {
			entries, _ := os.ReadDir(abspath)
			for _, e := range entries {
				if e.Name() == "index" || strings.HasPrefix(e.Name(), ".") { continue }
				childDepth := depth - 1
				if depth == -1 { childDepth = -1 }
				child, err := readPage(root, filepath.Join(relpath, e.Name()), childDepth)
				if err != nil { continue }
				p.Children = append(p.Children, child)
			}
		}
	} else {
		p.Type = "shallow"
		p.Content = readLines(abspath)
		p.Metadata = parseMetadata(p.Content)
	}

	return p, nil
}

func writePage(root string, p *Page) error {

	abspath := filepath.Join(root, filepath.Clean(p.Path))
	content := []byte(strings.Join(p.Content, "\n"))

	if p.Type == "deep" {
		if err := os.MkdirAll(abspath, 0755); err != nil { return err }
		return os.WriteFile(filepath.Join(abspath, "index"), content, 0644)
	}

	if err := os.MkdirAll(filepath.Dir(abspath), 0755); err != nil { return err }
	return os.WriteFile(abspath, content, 0644)
}

func deletePage(root, relpath string) error {

	abspath := filepath.Join(root, filepath.Clean(relpath))
	info, err := os.Stat(abspath)
	if err != nil { return err }
	if info.IsDir() { return os.RemoveAll(abspath) }
	return os.Remove(abspath)
}

func readLines(abspath string) []string {

	data, err := os.ReadFile(abspath)
	if err != nil { return []string{} }
	return strings.Split(string(data), "\n")
}

func parseMetadata(lines []string) map[string]any {

	meta := map[string]any{}
	for _, line := range lines {
		if isDashedLine(line) { break }
		idx := strings.Index(line, ":")
		if idx < 0 { continue }
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key != "" { meta[key] = val }
	}
	return meta
}

func isDashedLine(s string) bool {

	if s == "" { return false }
	for _, r := range s {
		if r != '-' { return false }
	}
	return true
}

func encodePage(p *Page) ([]byte, error) { return json.Marshal(p) }

func decodePage(data []byte) (*Page, error) {
	p := &Page{}
	return p, json.Unmarshal(data, p)
}
