// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

// pgconf prints the total confidence for a page of hOCR
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
		fmt.Fprintf(os.Stderr, "Usage: pgconf hocr\n")
		fmt.Fprintf(os.Stderr, "Prints the total confidence for a page, as an average of the confidence of each word.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	avg, err := hocr.GetAvgConf(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error retreiving confidence for %s: %v\n", flag.Arg(0), err)
	}

	fmt.Printf("%0.0f\n", avg)
}
