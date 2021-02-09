// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package hocr

// TODO: Parse line name to zero pad line numbers, so they can
//       be sorted easily

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"rescribe.xyz/utils/pkg/line"
)

// LineText extracts the text from an OcrLine
func LineText(l OcrLine) string {
	linetext := ""

	linetext = l.Text
	if noText(linetext) {
		linetext = ""
		for _, w := range l.Words {
			if w.Class != "ocrx_word" {
				continue
			}
			linetext += w.Text + " "
		}
	}
	if noText(linetext) {
		linetext = ""
		for _, w := range l.Words {
			if w.Class != "ocrx_word" {
				continue
			}
			for _, c := range w.Chars {
				if c.Class != "ocrx_cinfo" {
					continue
				}
				linetext += c.Text
			}
			linetext += " "
		}
	}
	linetext = strings.TrimRight(linetext, " ")
	return linetext
}

func parseLineDetails(h Hocr, dir string, name string) (line.Details, error) {
	lines := make(line.Details, 0)

	for _, p := range h.Pages {
		imgpath, err := imagePath(p.Title)
		if err != nil {
			return lines, err
		}
		imgpath = filepath.Join(dir, filepath.Base(imgpath))

		var img image.Image
		var gray *image.Gray
		pngf, err := os.Open(imgpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error opening image %s: %v", imgpath, err)
		}
		defer pngf.Close()
		img, _, err = image.Decode(pngf)
		if err == nil {
			b := img.Bounds()
			gray = image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
			draw.Draw(gray, b, img, b.Min, draw.Src)
		}

		for _, l := range p.Lines {
			totalconf := float64(0)
			num := 0
			for _, w := range l.Words {
				c, err := wordConf(w.Title)
				if err != nil {
					return lines, err
				}
				num++
				totalconf += c
			}

			coords, err := BoxCoords(l.Title)
			if err != nil {
				return lines, err
			}

			var ln line.Detail
			ln.Name = l.Id
			ln.Avgconf = (totalconf / float64(num)) / 100
			ln.Text = LineText(l)
			ln.OcrName = name
			if gray != nil {
				var imgd line.ImgDirect
				imgd.Img = gray.SubImage(image.Rect(coords[0], coords[1], coords[2], coords[3]))
				ln.Img = imgd
			}
			lines = append(lines, ln)
		}
		pngf.Close()
	}
	return lines, nil
}

// GetLineDetails parses a hocr file and returns a corresponding
// line.Details, including image extracts for each line
func GetLineDetails(hocrfn string) (line.Details, error) {
	var newlines line.Details

	file, err := ioutil.ReadFile(hocrfn)
	if err != nil {
		return newlines, err
	}

	h, err := Parse(file)
	if err != nil {
		return newlines, err
	}

	n := strings.Replace(filepath.Base(hocrfn), ".hocr", "", 1)
	return parseLineDetails(h, filepath.Dir(hocrfn), n)
}

// GetLineBasics parses a hocr file and returns a corresponding
// line.Details, without any image extracts
func GetLineBasics(hocrfn string) (line.Details, error) {
	var newlines line.Details

	file, err := ioutil.ReadFile(hocrfn)
	if err != nil {
		return newlines, err
	}

	h, err := Parse(file)
	if err != nil {
		return newlines, err
	}

	n := strings.Replace(filepath.Base(hocrfn), ".hocr", "", 1)
	return parseLineDetails(h, filepath.Dir(hocrfn), n)
}
