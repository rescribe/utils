// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: fonttobytes font.ttf")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		return
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalln("Failed to open file", flag.Arg(0), err)
	}
	fontbytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln("Failed to read file", flag.Arg(0), err)
	}

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write(fontbytes)
	w.Close()

	// This could be done with %+v in printf, but using the decimal rather than
	// hex output saves quite a few bytes, so we do that instead.
	fmt.Printf("[]byte{")
	for i, b := range compressed.Bytes() {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%d", b)
	}
	fmt.Printf("}\n")
}
