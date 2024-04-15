// Command covertool prints out the total coverage from a cover profile, excluding generated files.
// Generated files are .go files that contain a "DO NOT EDIT" comment.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/cover"
)

func main() {
	var fileName, baseDir string
	var trace bool
	var grouping string
	flag.StringVar(&fileName, "profile", "", "Coverage profile produced by `go test -coverprofile`")
	flag.StringVar(&baseDir, "base", ".", "Where to find the files in the coverage profile")
	flag.BoolVar(&trace, "trace", false, "Output debug info")
	flag.StringVar(&grouping, "grouping", "", "If provided, results will be matched against this regex, and grouped by the 1st grouped match, e.g. 'foo/([^/]*)/' will group foo/bar/baz as bar")
	flag.Parse()

	traceFn := func(string) {}
	if trace {
		traceFn = func(s string) {
			_, _ = fmt.Fprintln(os.Stderr, s)
		}
	}

	var groupRe *regexp.Regexp
	if grouping != "" {
		r, err := regexp.Compile(grouping)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		groupRe = r
	}

	if err := run(fileName, baseDir, traceFn, groupRe); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type bucket struct {
	statements int
	count      int
}

func (b bucket) add(bl cover.ProfileBlock) bucket {
	b.statements += bl.NumStmt
	if bl.Count > 0 {
		b.count += bl.NumStmt
	}
	return b
}

func (b bucket) coverPct() string {
	if b.statements == 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", float64(b.count)/float64(b.statements)*100)
}

func run(fileName, baseDir string, traceFn func(string), groupRe *regexp.Regexp) error {
	profiles, err := cover.ParseProfiles(fileName)
	if err != nil {
		return err
	}
	var total bucket
	buckets := make(map[string]bucket)
	for _, profile := range profiles {
		if gen, err := isGenerated(filepath.Join(baseDir, profile.FileName)); err != nil {
			return err
		} else if gen {
			traceFn(fmt.Sprintf("Skipping %s", profile.FileName))
			continue
		}
		for _, block := range profile.Blocks {
			total = total.add(block)
			if groupRe != nil {
				bucketName := "other"
				m := groupRe.FindStringSubmatch(profile.FileName)
				if len(m) > 1 {
					bucketName = m[1]
				}
				buckets[bucketName] = buckets[bucketName].add(block)
			}
		}
		traceFn(fmt.Sprintf("Added %s", profile.FileName))
	}

	if len(buckets) > 0 {
		sortedBuckets := make([]string, 0, len(buckets))
		for name, b := range buckets {
			sortedBuckets = append(sortedBuckets, fmt.Sprintf("%s: %s", name, b.coverPct()))
		}
		sort.Strings(sortedBuckets)
		for _, b := range sortedBuckets {
			fmt.Println(b)
		}
	}

	fmt.Printf("total: (statements) %s\n", total.coverPct())

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
