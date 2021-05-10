package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"unicode"
)

const usage = `Usage: dlgbook bookid [-a author] [-y year] [-t title] [savedir]

Downloads all pages from a Google Book, using the getgbook
tool, extracting the date, author name and title from
Google Books (unless given as arguments to dlgbook), and
saves them into a directory named YEAR_AUTHORSURNAME_Title 
`

const maxPartLength = 48

// formatAuthors formats a list of authors by just selecting
// the first one listed, and returning the uppercased final
// name.
func formatAuthors(authors []string) string {
	if len(authors) == 0 {
		return ""
	}

	s := authors[0]

	parts := strings.Fields(s)
	if len(parts) > 1 {
		s = parts[len(parts)-1]
	}

	s = strings.ToUpper(s)

	if len(s) > maxPartLength {
		s = s[:maxPartLength]
	}

	s = strings.Map(stripNonLetters, s)

	return s
}

// mapTitle is a function for strings.Map to strip out
// unwanted characters from the title.
func stripNonLetters(r rune) rune {
	if !unicode.IsLetter(r) {
		return -1
	}
	return r
}

// formatTitle formats a title to our preferences, notably
// by stripping spaces and punctuation characters.
func formatTitle(title string) string {
	s := strings.Map(stripNonLetters, title)
	if len(s) > maxPartLength {
		s = s[:maxPartLength]
	}
	return s
}

// getMetadata queries Google Books for metadata we care about
// and returns it formatted as we need it.
func getMetadata(id string) (string, string, string, error) {
	var author, title, year string
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes/%s", id)

	// designed to be unmarshalled by encoding/json's Unmarshal()
	type bookInfo struct {
		VolumeInfo struct {
			Title string
			Authors []string
			PublishedDate string
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return author, title, year, fmt.Errorf("Error downloading metadata %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return author, title, year, fmt.Errorf("Error downloading metadata %s: %v", url, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return author, title, year, fmt.Errorf("Error reading metadata %s: %v", url, err)
	}

	v := bookInfo{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return author, title, year, fmt.Errorf("Error parsing metadata %s: %v", url, err)
	}

	author = formatAuthors(v.VolumeInfo.Authors)
	title = formatTitle(v.VolumeInfo.Title)
	year = v.VolumeInfo.PublishedDate

	return author, title, year, nil
}

func main() {
	author := flag.String("author", "", "Set author, rather than autodetecting")
	title := flag.String("title", "", "Set title, rather than autodetecting")
	year := flag.String("year", "", "Set year, rather than autodetecting")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}
	bookid := flag.Arg(0)

	if *author == "" || *title == "" || *year == "" {
		a, t, y, err := getMetadata(bookid)
		if err != nil {
			log.Fatal(err)
		}
		if *author == "" {
			*author = a
		}
		if *title == "" {
			*title = t
		}
		if *year == "" {
			*year = y
		}
	}

	dir := fmt.Sprintf("%s_%s_%s_%s", *year, *author, *title, bookid)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("Couldn't create directory %s: %v", dir, err)
	}
	fmt.Printf("Saving book to %s\n", dir)

	cmd := exec.Command("getgbook", bookid)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error running getgbook %s: %v", bookid, err)
	}

	// getgbook downloads into bookid directory, so move files out of
	// there directly into dir
	tmpdir := path.Join(dir, bookid)
	f, err := os.Open(tmpdir)
	if err != nil {
		log.Fatalf("Failed to open %s to move files: %v", tmpdir, err)
	}
	files, err := f.Readdir(0)
	if err != nil {
		log.Fatalf("Failed to readdir %s to move files: %v", tmpdir, err)
	}
	for _, v := range files {
		orig := path.Join(tmpdir, v.Name())
		new := path.Join(dir, v.Name())
		err = os.Rename(orig, new)
		if err != nil {
			log.Fatalf("Failed to move %s to %s: %v", orig, new, err)
		}
	}

	err = os.Remove(tmpdir)
	if err != nil {
		log.Fatalf("Failed to remove temporary directory %s: %v", tmpdir, err)
	}

	fmt.Printf("Successfully download to %s\n", dir)
}
