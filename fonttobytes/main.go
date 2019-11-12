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

	fmt.Printf("[]byte{")
	first := true
	for _, b := range compressed.Bytes() {
		if first {
			fmt.Printf("%d", b)
			first = false
			continue
		}
		fmt.Printf(", %d", b)
	}
	fmt.Printf("}\n")
}
