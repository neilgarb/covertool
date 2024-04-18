// Command covertool prints out the total coverage from a cover profile, excluding generated files.
// Generated files are .go files that contain a "DO NOT EDIT" comment.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"

	"golang.org/x/tools/cover"
)

func main() {
	var fileName string
	var trace bool
	var grouping string
	var outPath string

	flag.StringVar(&fileName, "profile", "", "Coverage profile produced by `go test -coverprofile`")
	flag.BoolVar(&trace, "trace", false, "Output debug info")
	flag.StringVar(&grouping, "grouping", "", "If provided, results will be matched against this regex, and grouped by the 1st grouped match, e.g. 'foo/([^/]*)/' will group foo/bar/baz as bar")
	flag.StringVar(&outPath, "out", "", "If provided, results will be output to this file with this line format '[bucket] [total] [covered]'. This option requires grouping to be set.")
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

	if err := run(fileName, outPath, traceFn, groupRe); err != nil {
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

func run(fileName, outPath string, traceFn func(string), groupRe *regexp.Regexp) error {
	profiles, err := cover.ParseProfiles(fileName)
	if err != nil {
		return err
	}
	var total bucket
	buckets := make(map[string]bucket)
	for _, profile := range profiles {
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
		if outPath != "" {
			if err := writeReport(outPath, buckets); err != nil {
				return err
			}
		} else {
			sortedBuckets := make([]string, 0, len(buckets))
			for name, b := range buckets {
				sortedBuckets = append(sortedBuckets, fmt.Sprintf("%s: %s", name, b.coverPct()))
			}
			sort.Strings(sortedBuckets)
			for _, b := range sortedBuckets {
				fmt.Println(b)
			}
		}
	}

	fmt.Printf("total: (statements) %s\n", total.coverPct())

	return nil
}

func writeReport(path string, buckets map[string]bucket) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	for name, buck := range buckets {
		_, err := fmt.Fprintln(f, name, buck.statements, buck.count)
		if err != nil {
			return err
		}
	}
	return nil
}
