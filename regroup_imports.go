// regroup_imports reorders the import block of every .go file in the tree
// into three groups separated by blank lines:
//
//  1. standard library (first path element has no dot, and is not "tongcheng")
//  2. third-party       (first path element contains a dot)
//  3. local             (path == "tongcheng" or begins with "tongcheng/")
//
// Within each group imports are sorted alphabetically by import path.
// Only the text between `import (` and `)` is rewritten; aliases (`_`, `.`,
// named), trailing comments and the rest of the file are preserved.
// No imports are added or removed.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func classify(path string) int {
	if path == "tongcheng" || strings.HasPrefix(path, "tongcheng/") {
		return 2 // local
	}
	first := path
	if i := strings.Index(path, "/"); i >= 0 {
		first = path[:i]
	}
	if strings.Contains(first, ".") {
		return 1 // third-party
	}
	return 0 // stdlib
}

// extractPath returns the unquoted import path from a spec line like
// `repo2 "tongcheng/service/tongcheng/repo"` or `_ "go.uber.org/automaxprocs"`.
func extractPath(line string) string {
	q := strings.Index(line, "\"")
	if q < 0 {
		return ""
	}
	end := strings.LastIndex(line, "\"")
	if end <= q {
		return ""
	}
	return line[q+1 : end]
}

// regroup rewrites the text that sits between `import (` and `)` into the
// canonical three-group form. block is that text (including surrounding
// newlines). nl is the newline sequence used in the file.
func regroup(block, nl string) string {
	var groups [3][]string // spec line WITHOUT leading whitespace
	for _, line := range strings.Split(block, nl) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		pv := extractPath(trimmed)
		g := classify(pv)
		groups[g] = append(groups[g], trimmed)
	}
	for g := range groups {
		sort.SliceStable(groups[g], func(i, j int) bool {
			return extractPath(groups[g][i]) < extractPath(groups[g][j])
		})
	}
	var parts []string
	for g := 0; g < 3; g++ {
		if len(groups[g]) == 0 {
			continue
		}
		var ls []string
		for _, s := range groups[g] {
			ls = append(ls, "\t"+s)
		}
		parts = append(parts, strings.Join(ls, nl))
	}
	return nl + strings.Join(parts, nl+nl) + nl
}

func processFile(name string) (changed bool, err error) {
	src, err := os.ReadFile(name)
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	f, perr := parser.ParseFile(fset, name, src, parser.ParseComments|parser.AllErrors)
	if f == nil {
		return false, perr
	}

	nl := "\n"
	if bytes.Contains(src, []byte("\r\n")) {
		nl = "\r\n"
	}

	type mod struct {
		start, end int
		repl       string
	}
	var mods []mod
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}
		if !gen.Lparen.IsValid() {
			continue // single-line import, nothing to regroup
		}
		start := fset.Position(gen.Lparen).Offset
		end := fset.Position(gen.Rparen).Offset
		block := string(src[start+1 : end])
		newBlock := regroup(block, nl)
		if newBlock != block {
			mods = append(mods, mod{start + 1, end, newBlock})
		}
	}
	if len(mods) == 0 {
		return false, nil
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].start < mods[j].start })
	var buf bytes.Buffer
	prev := 0
	for _, m := range mods {
		buf.Write(src[prev:m.start])
		buf.WriteString(m.repl)
		prev = m.end
	}
	buf.Write(src[prev:])
	if bytes.Equal(buf.Bytes(), src) {
		return false, nil
	}
	return true, os.WriteFile(name, buf.Bytes(), 0644)
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	var total, changed, skipped int
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if path != root && (base == "vendor" || strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if filepath.Base(path) == "regroup_imports.go" {
			return nil
		}
		total++
		ch, err := processFile(path)
		if err != nil {
			skipped++
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", path, err)
			return nil
		}
		if ch {
			changed++
			fmt.Println("regrouped:", path)
		}
		return nil
	})
	fmt.Printf("\n%d files total, %d regrouped, %d skipped\n", total, changed, skipped)
}
