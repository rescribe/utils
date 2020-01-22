package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: boxtotxt in.box\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(flag.Arg(0))
	defer f.Close()
	if err != nil {
		log.Fatalf("Could not open file %s: %v\n", flag.Arg(0), err)
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		t := scanner.Text()
		s := strings.Split(t, "")
		if len(s) < 1 {
			continue
		}
		fmt.Printf("%s", s[0])
	}

	fmt.Printf("\n")
}
