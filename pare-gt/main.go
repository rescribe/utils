package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const usage = `Usage: pare-gt [-n num] gtdir movedir

Moves some of the ground truth from gt-dir into movedir,
ensuring that the same proportions of each ground truth
source are represented in the moved section. Proportion of
ground truth source is calculated by taking the prefix of
the filename up to the first '-' character.
`

// Prefixes is a map of the prefix string to a list of filenames
type Prefixes = map[string][]string

// walker adds any .txt path to prefixes map, under the appropriate
// prefix (blank if no '-' separator was found)
func walker(prefixes *Prefixes) filepath.WalkFunc{
	return func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := path.Ext(fpath)
		if ext != ".txt" {
			return nil
		}
		base := path.Base(fpath)
		idx := strings.Index(base, "-")
		var prefix string
		if idx > -1 {
			prefix = base[0:idx]
		}
		noext := strings.TrimSuffix(fpath, ext)
		(*prefixes)[prefix] = append((*prefixes)[prefix], noext)
		return nil
	}
}

// inStrSlice checks whether a given string is part of a slice of
// strings
func inStrSlice(sl []string, s string) bool {
	for _, v := range sl {
		if s == v {
			return true
		}
	}
	return false
}

// samplePrefixes selects random samples for each prefix, proportional
// to the amount of that prefix there are in the whole set, so that a
// total of perctosample% are sampled.
func samplePrefixes(perctosample int, prefixes Prefixes) (filestomove []string) {
	var total, sample int
	for _, v := range prefixes {
		total += len(v)
	}

	sample = total / perctosample

	for _, prefix := range prefixes {
		len := len(prefix)
		if len == 1 {
			continue
		}
		numtoget := int(float64(sample) / float64(total) * float64(len))
		if numtoget < 1 {
			numtoget = 1
		}
		for i:=0; i<numtoget; i++ {
			var selected string
			selected = prefix[rand.Int()%len]
			// pick a different random selection if the first one is
			// already in the filestomove slice
			for inStrSlice(filestomove, selected) {
				selected = prefix[rand.Int()%len]
			}
			filestomove = append(filestomove, selected)
		}
	}

	return
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	numtopare := flag.Int("n", 10, "Percentage of the ground truth to pare away.")
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	for _, d := range flag.Args() {
		info, err := os.Stat(d)
		if err != nil || !info.IsDir() {
			log.Fatalln("Error accessing directory", flag.Arg(0), err)
		}
	}

	var prefixes Prefixes
	prefixes = make(Prefixes)
	err := filepath.Walk(flag.Arg(0), walker(&prefixes))
	if err != nil {
		log.Fatalln("Failed to walk", flag.Arg(0), err)
	}

	filestomove := samplePrefixes(*numtopare, prefixes)

	for _, f := range filestomove {
		fmt.Println("Moving ground truth", f)
		b := path.Base(f)
		for _, ext := range []string{".txt", ".png"} {
			err = os.Rename(f + ext, path.Join(flag.Arg(1), b + ext))
			if err != nil {
				log.Fatalln("Error moving file", f + ext, err)
			}
		}
	}
}
