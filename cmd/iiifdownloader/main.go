package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

const usage = `Usage: iiifdownloader url

Downloads all pages from a IIIF server.

Currently supports the following IIIF using services:
- BNF's Gallica   example url: https://gallica.bnf.fr/ark:/12148/bpt6k6468158v
- BSB / MDZ       example url: https://reader.digitale-sammlungen.de//de/fs1/object/display/bsb10132387_00005.html
- DFG Viewer      example url: http://dfg-viewer.de/show?set%5Bmets%5D=http%3A%2F%2Fdaten.digitale-sammlungen.de%2F~db%2Fmets%2Fbsb11274872_mets.xml&cHash=fd18451ee968c125ab2bdbfd3717eae6
`

const bnfPrefix = `https://gallica.bnf.fr/ark:/`
const bsbPrefix = `https://reader.digitale-sammlungen.de/de/fs1/object/display/`
const dfgPrefix = `http://dfg-viewer.de/`

// filesAreIdentical checks whether two files are identical.
func filesAreIdentical(fn1, fn2 string) (bool, error) {
	f1, err := os.Open(fn1)
	defer f1.Close()
	if err != nil {
		return false, fmt.Errorf("Error opening file %s: %v\n", fn1, err)
	}
	b1, err := ioutil.ReadAll(f1)
	if err != nil {
		return false, fmt.Errorf("Error reading file %s: %v\n", fn1, err)
	}
	f2, err := os.Open(fn2)
	defer f2.Close()
	if err != nil {
		return false, fmt.Errorf("Error opening file %s: %v\n", fn2, err)
	}
	b2, err := ioutil.ReadAll(f2)
	if err != nil {
		return false, fmt.Errorf("Error reading file %s: %v\n", fn2, err)
	}

	for i, _ := range b1 {
		if b1[i] != b2[i] {
			return false, nil
		}
	}
	return true, nil
}

// parseMets downloads and parses an XML encoded METS file,
// returning a list of image URLs.
// Example URL: https://daten.digitale-sammlungen.de/~db/mets/bsb10132387_mets.xml
func parseMets(u string) ([]string, error) {
	var urls []string

	// designed to be unmarshalled by encoding/xml's Unmarshal()
	type metsXML struct {
		FileGrps []struct {
			Attr string `xml:"USE,attr"`
			Files []struct {
				Url string `xml:"href,attr"`
			} `xml:"file>FLocat"`
		} `xml:"fileSec>fileGrp"`
	}

	resp, err := http.Get(u)
	if err != nil {
		return urls, fmt.Errorf("Error downloading mets %s: %v", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return urls, fmt.Errorf("Error downloading mets %s: %v", u, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return urls, fmt.Errorf("Error reading mets XML %s: %v", u, err)
	}

	v := metsXML{}
	err = xml.Unmarshal(b, &v)
	if err != nil {
		return urls, fmt.Errorf("Error parsing mets XML %s: %v", u, err)
	}

	for _, grp := range v.FileGrps {
		if grp.Attr == "MAX" {
			for _, f := range grp.Files {
				urls = append(urls, f.Url)
			}
		}
	}

	return urls, nil
}

// parseIIIFManifest downloads and parses a JSON encoded IIIF
// Manifest file, returning a list of image URLs.
// Example URL: https://api.digitale-sammlungen.de/iiif/presentation/v2/bsb10132387/manifest
func parseIIIFManifest(u string) ([]string, error) {
	var urls []string

	// designed to be unmarshalled by encoding/json's Unmarshal()
	type iiifManifest struct {
		Sequences []struct {
			Canvases []struct {
				Images []struct {
					Resource struct {
						Id string `json:"@id"`
					}
				}
			}
		}
	}

	resp, err := http.Get(u)
	if err != nil {
		return urls, fmt.Errorf("Error downloading IIIF manifest %s: %v", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return urls, fmt.Errorf("Error downloading IIIF manifest %s: %v", u, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return urls, fmt.Errorf("Error reading IIIF manifest %s: %v", u, err)
	}

	v := iiifManifest{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return urls, fmt.Errorf("Error parsing IIIF manifest %s: %v", u, err)
	}

	for _, canvas := range v.Sequences[0].Canvases {
		for _, image := range canvas.Images {
			urls = append(urls, image.Resource.Id)
		}
	}

	return urls, nil
}

// urlToPgName returns an appropriate filename for a page, given
// a url. This is currently optimised for BSB URLs, but will be
// made more generic when necessary.
func urlToPgName(u string) string {
	safe := strings.Replace(u, "/", "_", -1)

	b := path.Base(u)
	if b != "default.jpg" {
		return b
	}

	f := strings.Split(u, "/")
	if len(f) < 5 {
		return safe
	}
	name := f[len(f) - 5]

	f2 := strings.Split(name, "_")
	var numpart, pgnum string
	if len(f2) < 2 {
		numpart = name
	} else {
		numpart = f2[len(f2)-1]
	}
	if len(numpart) < 4 {
		pgnum = numpart
	} else {
		pgnum = numpart[len(numpart)-4:]
	}

	return pgnum + ".jpg"
}

// dlPage downloads a page url to bookdir.
func dlPage(bookdir, u string) error {
	name := urlToPgName(u)
	fn := path.Join(bookdir, name)

	fmt.Printf("Downloading page %s to %s\n", u, fn)

	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("Error downloading page %s: %v", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error downloading page %s: %v", u, err)
	}

	f, err := os.Create(fn)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("Error creating file %s: %v\n", fn, err)
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("Error writing file %s: %v\n", fn, err)
	}

	resp.Body.Close()
	f.Close()

	return nil
}

// dlNoPgNums downloads all pages, starting from zero, until either
// a 404 is returned, or identical files are returned for two subsequent
// pages (the latter being the behaviour of BNF's server).
func dlNoPgNums(bookdir, pgurlStart, pgurlEnd, pgurlAltStart, pgurlAltEnd string) error {
	pgnum := 0
	for {
		pgnum++

		fmt.Printf("Downloading page %d\n", pgnum)

		fn := path.Join(bookdir, fmt.Sprintf("%04d.jpg", pgnum))
		_, err := os.Stat(fn)
		if err == nil || os.IsExist(err) {
			fmt.Printf("Skipping already present page %d\n", pgnum)
			continue
		}

		u := fmt.Sprintf("%s%d%s", pgurlStart, pgnum, pgurlEnd)
		resp, err := http.Get(u)
		if err != nil {
			return fmt.Errorf("Error downloading page %d, %s: %v\n", pgnum, u, err)
		}
		defer resp.Body.Close()
		switch {
		case resp.StatusCode == http.StatusNotFound:
			fmt.Printf("Got 404, assuming end of pages, for page %d, %s\n", pgnum, u)
			return nil
		case resp.StatusCode != http.StatusOK:
			fmt.Printf("Error downloading page %d, %s: HTTP Code %s\n", pgnum, u, resp.Status)

			if pgurlAltStart == "" && pgurlAltEnd == "" {
				return fmt.Errorf("No alternative URL to try, book failed (or ended, hopefully)")
			}

			fmt.Printf("Trying to redownload page %d at lower quality\n", pgnum)
			u = fmt.Sprintf("%s%d%s", pgurlAltStart, pgnum, pgurlAltEnd)
			resp, err = http.Get(u)
			if err != nil {
				return fmt.Errorf("Error downloading page %d, %s: %v\n", pgnum, u, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("Error downloading page %d, %s: HTTP Code %s\n", pgnum, u, resp.Status)
			}
		}

		f, err := os.Create(fn)
		defer f.Close()
		if err != nil {
			return fmt.Errorf("Error creating file %s: %v\n", fn, err)
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return fmt.Errorf("Error writing file %s: %v\n", fn, err)
		}

		// Close once finished with, as defer won't trigger until the end of the function
		resp.Body.Close()
		f.Close()

		// Check that the last two downloaded files aren't identical, as this
		// can happen when there are no more pages to download.
		if pgnum == 1 {
			continue
		}
		fn2 := path.Join(bookdir, fmt.Sprintf("%04d.jpg", pgnum-1))
		identical, err := filesAreIdentical(fn, fn2)
		if err != nil {
			return fmt.Errorf("Error checking for files being identical: %v\n", err)
		}
		if identical {
			fmt.Println("Last 2 pages were identical, looks like it's the end of the book")
			err = os.Remove(fn)
			if err != nil {
				return fmt.Errorf("Error removing dupilicate page %d: %v", fn, err)
			}
			return nil
		}
	}
}

// sanitiseUrl partly sanitises a url. This is very basic,
// but enough for us for now.
func sanitiseUrl(u string) string {
	var s string
	s = strings.Replace(u, "//", "/", -1)
	s = strings.Replace(s, "https:/", "https://", -1)
	s = strings.Replace(s, "http:/", "http://", -1)
	return s
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	u := sanitiseUrl(flag.Arg(0))

	var bookdir string
	var pgurlStart, pgurlEnd string
	var pgurlAltStart, pgurlAltEnd string
	var pgUrls []string
	var noPgNums bool
	var err error

	switch {
	case strings.HasPrefix(u, bnfPrefix):
		f := strings.Split(u[len(bnfPrefix):], "/")
		if len(f) < 2 {
			log.Fatalln("Failed to extract BNF book ID from URL")
		}
		var lastpart string
		dot := strings.Index(f[1], ".")
		if dot == -1 {
			lastpart = f[1]
		} else {
			lastpart = f[1][0:dot]
		}
		bookid := f[0] + "/" + lastpart
		bookdir = f[0] + "-" + lastpart

		pgurlStart = "https://gallica.bnf.fr/iiif/ark:/" + bookid + "/f"
		pgurlEnd = "/full/full/0/native.jpg"
		noPgNums = true

		// BNF don't have all pages available from IIIF, but they do have
		// the missing ones in less good quality from an alternative URL.
		pgurlAltStart = "https://gallica.bnf.fr/ark:/" + bookid + "/f"
		pgurlAltEnd = ".highres"
	case strings.HasPrefix(u, bsbPrefix):
		f := strings.Split(u[len(bsbPrefix):], "_")
		if len(f) < 2 {
			log.Fatalln("Failed to extract BNF book ID from URL")
		}
		bookid := f[0]
		bookdir = bookid
		iiifurl := "https://api.digitale-sammlungen.de/iiif/presentation/v2/" + bookid + "/manifest"

		pgUrls, err = parseIIIFManifest(iiifurl)
		if err != nil {
			log.Fatalf("Error parsing IIIF manifest url %s: %v\n", iiifurl, err)
		}
	case strings.HasPrefix(u, dfgPrefix):
		// dfg can have a url encoded mets url in several parts of the viewer url
		metsNames := []string{"set[mets]", "tx_dlf[id]"}
		var metsurl string
		escurl, err := url.QueryUnescape(u)
		if err != nil {
			log.Fatalf("Error unescaping url %s: %v\n", u, err)
		}
		for _, v := range metsNames {
			i := strings.Index(escurl, v)
			if i != -1 {
				start := i + len(v) + 1 // +1 is to pass the equals sign
				end := strings.Index(escurl[start:], "&")
				if end == -1 {
					end = len(escurl)
				} else {
					end += start
				}
				metsurl = escurl[start:end]
			}
		}
		if len(metsurl) == 0 {
			log.Fatalf("No mets url found in %s\n", u)
		}

		b := path.Base(metsurl)
		f := strings.Split(b, "_")
		bookdir = f[0]

		pgUrls, err = parseMets(metsurl)
		if err != nil {
			log.Fatalf("Error parsing mets url %s: %v\n", metsurl, err)
		}
	default:
		log.Fatalln("Error: generic IIIF downloading not supported yet")
	}

	err = os.MkdirAll(bookdir, 0777)
	if err != nil {
		log.Fatalf("Error creating book dir: %v\n", err)
	}

	if len(pgUrls) > 0 {
		for _, v := range pgUrls {
			dlPage(bookdir, v)
		}
	} else if noPgNums {
		err = dlNoPgNums(bookdir, pgurlStart, pgurlEnd, pgurlAltStart, pgurlAltEnd)
		if err != nil {
			log.Fatalf("Error downloading pages: %v\n", err)
		}
	} else {
		log.Fatalf("Failed to find any pages\n")
	}
}
