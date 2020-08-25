// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// dehyphenate does basic dehyphenation on a hocr file
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"rescribe.xyz/utils/pkg/hocr"
)

// BUGS:
// - loses all elements not captured in hocr structure such as html headings
//   might be best to copy the header and footer separately and put the hocr in between, but would still need to ensure all elements are captured
// - loses any formatting; doesn't need to be identical, but e.g. linebreaks after elements would be handy
// - need to handle OcrChar

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dehyphenate [-hocr] in out\n")
		fmt.Fprintf(os.Stderr, "Dehyphenates a file.\n")
		flag.PrintDefaults()
	}
	usehocr := flag.Bool("hocr", false, "process hocr files, rather than plain text")
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	in, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error reading %s: %v", flag.Arg(1), err)
	}

	var finaltxt string
	var h hocr.Hocr

	if *usehocr {
		h, err = hocr.Parse(in)
		if err != nil {
			log.Fatal(err)
		}

		for i, l := range h.Lines {
			w := l.Words[len(l.Words)-1]
			if len(w.Chars) == 0 {
				if len(w.Text) > 0 && w.Text[len(w.Text) - 1] == '-' {
					h.Lines[i].Words[len(l.Words)-1].Text = w.Text[0:len(w.Text)-1] + h.Lines[i+1].Words[0].Text
					h.Lines[i+1].Words[0].Text = ""
				}
			} else {
				log.Printf("TODO: handle OcrChar")
			}
		}
	} else {
		var newlines []string
		lines := strings.Split(string(in), "\n")
		for i, line := range lines {
			words := strings.Split(line, " ")
			last := words[len(words)-1]
			// the - 2 here is to account for a trailing newline and counting from zero
			if len(last) > 0 && last[len(last) - 1] == '-' && i < len(lines) - 2 {
				nextwords := strings.Split(lines[i+1], " ")
				if len(nextwords) > 0 {
					line = line[0:len(line)-1] + nextwords[0]
				}
				if len(nextwords) > 1 {
					lines[i+1] = strings.Join(nextwords[1:], " ")
				} else {
					lines[i+1] = ""
				}
			}
			newlines = append(newlines, line)
		}
		finaltxt = strings.Join(newlines, "\n")
	}

	f, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatalf("Error creating file %s: %v", flag.Arg(1), err)
	}
	defer f.Close()

	if *usehocr {
		e := xml.NewEncoder(f)
		err = e.Encode(h)
		if err != nil {
			log.Fatalf("Error encoding XML: %v", err)
		}
	} else {
		_, err := io.WriteString(f, finaltxt)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}
}
