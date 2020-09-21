// Copyright 2020 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// analysestats analyses a set of 'best', 'conf', and 'hocr' files
// in a directory, outputting results to a .csv file for further
// investigation.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"strconv"
)

const usage = `Usage: analysestats statsdir csvfile

analysestats analyses a set of 'best', 'conf', and 'hocr' files
in the 'statsdir' directory, outputting results to the 'csvfile'
file in CSV format for further investigation.
`

// stat represents key stats / metadata for a book
type stat struct {
	mean     float64
	stddev   float64
	training string
	year     int
}

// Bookstats is a map of the stats attached to each book (key is book name)
type Bookstats = map[string]stat

func walker(bookstats *Bookstats) filepath.WalkFunc {
	return func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b := filepath.Base(fpath)
		parts := strings.Split(b, "-")
		// if no - or name is too short to have a useful prefix, bail
		if len(parts) < 2 || len(b) < 6 {
			return nil
		}
		prefix := b[0:len(b)-6] // 6 is length of '-hocr' + 1
		ext := parts[len(parts)-1]

		if ext != "hocr" && ext != "best" {
			return nil
		}

		var year int
		parts2 := strings.Split(b, "_")
		if len(parts2) > 2 {
			// we can ignore an error as a zero year is correct in that case anyway
			year, _ = strconv.Atoi(parts2[0])
		}

		_, ok := (*bookstats)[prefix]
		if !ok {
			(*bookstats)[prefix] = stat{year: year}
		}

		switch ext {
		case "hocr":
			// TODO: parse hocr enough to get training used
		case "best":
			// TODO: read conf also and fill in mean and stddev
		}

		return nil
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	info, err := os.Stat(flag.Arg(0))
	if err != nil || !info.IsDir() {
		log.Fatalln("Error accessing directory", flag.Arg(0), err)
	}

	var bookstats Bookstats
	bookstats = make(Bookstats)

	err = filepath.Walk(flag.Arg(0), walker(&bookstats))
	if err != nil {
		log.Fatalln("Failed to walk", flag.Arg(0), err)
	}

	f, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatalf("Failed to create file %s: %v\n", flag.Arg(1), err)
	}
	defer f.Close()
	csvw := csv.NewWriter(f)

	csvw.Write([]string{"Year", "Name", "Mean", "Standard Deviation", "Training"})
	for name, stats := range bookstats {
		year := fmt.Sprintf("%d", stats.year)
		mean := fmt.Sprintf("%0.1f", stats.mean)
		stddev := fmt.Sprintf("%0.1f", stats.stddev)
		err = csvw.Write([]string{name, year, mean, stddev, stats.training})
		if err != nil {
			log.Fatalf("Failed to write record %s to csv: %v\n", name, err)
		}
	}
	csvw.Flush()
}
