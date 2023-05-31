// Command covertool prints out the total coverage from a cover profile, excluding generated files.
// Generated files are .go files that contain a "DO NOT EDIT" comment.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

func main() {
	var fileName, baseDir string
	var trace bool
	flag.StringVar(&fileName, "profile", "", "Coverage profile produced by `go test -coverprofile`")
	flag.StringVar(&baseDir, "base", ".", "Where to find the files in the coverage profile")
	flag.BoolVar(&trace, "trace", false, "Output debug info")
	flag.Parse()

	traceFn := func(string) {}
	if trace {
		traceFn = func(s string) {
			_, _ = fmt.Fprintln(os.Stderr, s)
		}
	}

	if err := run(fileName, baseDir, traceFn); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(fileName, baseDir string, traceFn func(string)) error {
	profiles, err := cover.ParseProfiles(fileName)
	if err != nil {
		return err
	}
	var totStmt, totCount int
	for _, profile := range profiles {
		if gen, err := isGenerated(filepath.Join(baseDir, profile.FileName)); err != nil {
			return err
		} else if gen {
			traceFn(fmt.Sprintf("Skipping %s", profile.FileName))
			continue
		}
		for _, block := range profile.Blocks {
			totStmt += block.NumStmt
			totCount += block.Count
		}
		traceFn(fmt.Sprintf("Added %s", profile.FileName))
	}

	pct := "0.00%"
	if totStmt > 0 {
		pct = fmt.Sprintf("%.2f", float64(totCount)/float64(totStmt)*100)
	}

	fmt.Printf("total: (statements) %s%%\n", pct)

	return nil
}

func isGenerated(path string) (bool, error) {
	if !strings.HasSuffix(path, ".go") {
		return false, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false, nil
	}
	defer func() {
		_ = f.Close()
	}()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "//") && strings.Contains(scanner.Text(), "DO NOT EDIT") {
			return true, scanner.Err()
		}
	}
	return false, scanner.Err()
}
