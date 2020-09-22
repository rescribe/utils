// Copyright 2020 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// analysestats analyses a set of 'best', 'conf', and 'hocr' files
// in a directory, outputting results to a .csv file for further
// investigation.
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
type Bookstats = map[string]*stat

type hocrPars struct {
	Par []struct {
		Lang string `xml:"lang,attr"`
	} `xml:"body>div>div>p"`
}

// getTrainingUsed parses a hOCR file to find the training
// file used to create it.
func getTrainingUsed(hocrfn string) (string, error) {
	b, err := ioutil.ReadFile(hocrfn)
	if err != nil {
		return "", err
	}

	var par hocrPars
	err = xml.Unmarshal(b, &par)
	if err != nil {
		return "", err
	}

	if len(par.Par) < 1 {
		return "", fmt.Errorf("No <p> tags found")
	}

	return par.Par[0].Lang, nil
}

// getMeanStddevOfBest calculates the mean and standard deviation
// of the confidence values of every page in bestfn, as listed in
// conffn.
func getMeanStddevOfBest(bestfn string, conffn string) (float64, float64, error) {
	f, err := os.Open(conffn)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to open %s: %v", conffn, err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)

	// create a map of confs from the conf file
	var confs map[string]int
	confs = make(map[string]int)
	for s.Scan() {
		line := s.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		c, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		fn := filepath.Base(parts[0])
		confs[fn] = c
	}

	f, err = os.Open(bestfn)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to open %s: %v", bestfn, err)
	}
	defer f.Close()
	s = bufio.NewScanner(f)

	var bestConfs []int
	for s.Scan() {
		fn := s.Text()
		c, ok := confs[fn]
		if !ok {
			continue
		}
		bestConfs = append(bestConfs, c)
	}

	var sum int
	for _, v := range bestConfs {
		sum += v
	}
	mean := float64(sum) / float64(len(bestConfs))

	var a, stddev float64
	if len(bestConfs) > 1 {
		for _, v := range bestConfs {
			a += (float64(v) - mean) * (float64(v) - mean)
		}
		variance := a / float64(len(bestConfs) - 1)
		stddev = math.Sqrt(variance)
	}

	return mean, stddev, nil
}

// walker returns a walkfunc that checks for hocr and best files,
// and uses them to fill the bookstats map & structure. Note that
// the stat file is read when the best file is read, as they need
// to be parsed together to get the statistics we're interested
// in.
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
			(*bookstats)[prefix] = &stat{year: year}
		}

		switch ext {
		case "hocr":
			training, err := getTrainingUsed(fpath)
			if err != nil {
				log.Printf("Warning: failed to get training used from %s: %v\n", fpath, err)
				return nil
			}
			(*bookstats)[prefix].training = training
		case "best":
			confpath := strings.Replace(fpath, "-best", "-conf", -1)
			mean, stddev, err := getMeanStddevOfBest(fpath, confpath)
			if err != nil {
				log.Printf("Warning: failed to get mean & standard deviation from %s and %s: %v\n", fpath, confpath, err)
				return nil
			}
			(*bookstats)[prefix].mean = mean
			(*bookstats)[prefix].stddev = stddev
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

	csvw.Write([]string{"Name", "Year", "Mean", "Standard Deviation", "Training"})
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
