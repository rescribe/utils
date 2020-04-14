// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// hocrtotxt prints the text from a hocr file
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"rescribe.xyz/utils/pkg/hocr"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hocrtotxt hocrfile\n")
		fmt.Fprintf(os.Stderr, "Prints the text from a hocr file.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	text, err := hocr.GetText(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", text)
}
