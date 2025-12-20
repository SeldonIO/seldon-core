package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	root := flag.String("root", ".", "root folder to scan")
	out := flag.String("out", "steps.txt", "output file")
	method := flag.String("method", "Step", "method name to match (e.g. Step)")
	receiver := flag.String("receiver", "", "optional receiver identifier to match (e.g. scenario). Empty matches any receiver.")
	flag.Parse()

	steps, err := extractSteps(*root, *method, *receiver)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	sort.Strings(steps)

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	for _, s := range steps {
		// Write exactly as requested: step: `...`
		// Use backticks in output for readability, even if original used quotes.
		fmt.Fprintf(f, "step: `%s`\n", s)
	}

	fmt.Printf("Wrote %d step(s) to %s\n", len(steps), *out)
}

func extractSteps(root, method, receiver string) ([]string, error) {
	seen := map[string]struct{}{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Skip vendor and hidden dirs by default
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		// Skip generated files (common heuristic)
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_generated.go") || strings.HasSuffix(base, ".gen.go") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			// If a file doesn't parse, skip it but report which file
			// (you can change this to return err if you want hard-fail)
			fmt.Fprintf(os.Stderr, "warn: could not parse %s: %v\n", path, err)
			return nil
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Match: <receiver>.<method>(...)
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if sel.Sel == nil || sel.Sel.Name != method {
				return true
			}

			// Optional receiver filter (e.g. require "scenario.Step")
			if receiver != "" {
				if id, ok := sel.X.(*ast.Ident); !ok || id.Name != receiver {
					return true
				}
			}

			if len(call.Args) < 1 {
				return true
			}

			// We want the first argument string literal
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			// Unquote handles both raw (`...`) and interpreted ("...")
			unq, err := strconv.Unquote(lit.Value)
			if err != nil {
				// If unquote fails, skip
				return true
			}

			unq = strings.TrimSpace(unq)
			if unq == "" {
				return true
			}

			seen[unq] = struct{}{}
			return true
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	steps := make([]string, 0, len(seen))
	for s := range seen {
		steps = append(steps, s)
	}
	return steps, nil
}
