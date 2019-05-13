package preproc

// TODO: add minimum size variable (default ~30%?)
// TODO: have the integral image specific stuff done by interface functions

import (
	"image"
	"image/color"

	"rescribe.xyz/go.git/binarize"
)

type windowslice struct {
	topleft     uint64
	topright    uint64
	bottomleft  uint64
	bottomright uint64
}

func getwindowslice(i [][]uint64, x int, size int) windowslice {
	maxy := len(i) - 1
	maxx := x + size
	if maxx > len(i[0])-1 {
		maxx = len(i[0]) - 1
	}

	return windowslice{i[0][x], i[0][maxx], i[maxy][x], i[maxy][maxx]}
}

// checkwindow checks the window from x to see whether more than
// thresh proportion of the pixels are white, if so it returns true.
func checkwindow(integral [][]uint64, x int, size int, thresh float64) bool {
	height := len(integral)
	window := getwindowslice(integral, x, size)
	// divide by 255 as each on pixel has the value of 255
	sum := (window.bottomright + window.topleft - window.topright - window.bottomleft) / 255
	area := size * height
	proportion := float64(area)/float64(sum) - 1
	return proportion <= thresh
}

// returns the proportion of the given window that is black pixels
func proportion(integral [][]uint64, x int, size int) float64 {
	height := len(integral)
	window := getwindowslice(integral, x, size)
	// divide by 255 as each on pixel has the value of 255
	sum := (window.bottomright + window.topleft - window.topright - window.bottomleft) / 255
	area := size * height
	return float64(area)/float64(sum) - 1
}

// findbestedge goes through every vertical line from x to x+w to
// find the one with the lowest proportion of black pixels.
func findbestedge(integral [][]uint64, x int, w int) int {
	var bestx int
	var best float64

	if w == 1 {
		return x
	}

	right := x + w
	for ; x < right; x++ {
		prop := proportion(integral, x, 1)
		if prop > best {
			best = prop
			bestx = x
		}
	}

	return bestx
}

// findedges finds the edges of the main content, by moving a window of wsize
// from the middle of the image to the left and right, stopping when it reaches
// a point at which there is a lower proportion of black pixels than thresh.
func findedges(integral [][]uint64, wsize int, thresh float64) (int, int) {
	maxx := len(integral[0]) - 1
	var lowedge, highedge int = 0, maxx

	for x := maxx / 2; x < maxx-wsize; x++ {
		if checkwindow(integral, x, wsize, thresh) {
			highedge = findbestedge(integral, x, wsize)
			break
		}
	}

	for x := maxx / 2; x > 0; x-- {
		if checkwindow(integral, x, wsize, thresh) {
			lowedge = findbestedge(integral, x, wsize)
			break
		}
	}

	return lowedge, highedge
}

// wipesides fills the sections of image not within the boundaries
// of lowedge and highedge with white
func wipesides(img *image.Gray, lowedge int, highedge int) *image.Gray {
	b := img.Bounds()
	new := image.NewGray(b)

	// set left edge white
	for x := b.Min.X; x < lowedge; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(x, y, color.Gray{255})
		}
	}
	// copy middle
	for x := lowedge; x < highedge; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(x, y, img.GrayAt(x, y))
		}
	}
	// set right edge white
	for x := highedge; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			new.SetGray(x, y, color.Gray{255})
		}
	}

	return new
}

// wipe fills the sections of image which fall outside the content
// area with white
func Wipe(img *image.Gray, wsize int, thresh float64) *image.Gray {
	integral := binarize.Integralimg(img)
	lowedge, highedge := findedges(integral, wsize, thresh)
	return wipesides(img, lowedge, highedge)
}