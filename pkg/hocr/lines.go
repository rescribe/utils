package hocr

// TODO: Parse line name to zero pad line numbers, so they can
//       be sorted easily

import (
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

func getLineText(l OcrLine) (string) {
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
	linetext += "\n"
	return linetext
}

func parseLineDetails(h Hocr, i *image.Gray, name string) (line.Details, error) {
	lines := make(line.Details, 0)

	for _, l := range h.Lines {
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

		coords, err := boxCoords(l.Title)
		if err != nil {
			return lines, err
		}

		var ln line.Detail
		ln.Name = l.Id
		ln.Avgconf = (totalconf / float64(num)) / 100
		ln.Text = getLineText(l)
		ln.OcrName = name
		if i != nil {
			var imgd line.ImgDirect
			imgd.Img = i.SubImage(image.Rect(coords[0], coords[1], coords[2], coords[3]))
			ln.Img = imgd
		}
		lines = append(lines, ln)
	}
	return lines, nil
}

// GetLineDetails parses a hocr file and returns a corresponding
// line.Details, including image extracts for each line.
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

	var img image.Image
	var gray *image.Gray
	pngfn := strings.Replace(hocrfn, ".hocr", ".png", 1)
	pngf, err := os.Open(pngfn)
	if err == nil {
		defer pngf.Close()
		img, _, err = image.Decode(pngf)
		if err == nil {
			b := img.Bounds()
			gray = image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
			draw.Draw(gray, b, img, b.Min, draw.Src)
		}
	}

	n := strings.Replace(filepath.Base(hocrfn), ".hocr", "", 1)
	return parseLineDetails(h, gray, n)
}

// GetLineBasics parses a hocr file and returns a corresponding
// line.Details, without any image extracts.
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
	return parseLineDetails(h, nil, n)
}
