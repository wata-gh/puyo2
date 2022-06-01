package main

import (
	"bufio"
	"flag"
	"image"
	"image/draw"
	"image/png"
	"os"
	"strings"
)

type Canvas struct {
	Width     int
	Height    int
	Image     *image.NRGBA
	PuyoImage image.Image
}

func NewCanvas(width int, height int) *Canvas {
	fpuyo, err := os.Open("images/puyos.png")
	if err != nil {
		panic(err)
	}
	defer fpuyo.Close()

	canvas := new(Canvas)
	canvas.PuyoImage, _, err = image.Decode(fpuyo)
	if err != nil {
		panic(err)
	}
	canvas.Width = width
	canvas.Height = height
	canvas.Image = image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*width, 32*height)})
	return canvas
}

func (c *Canvas) PlacePuyo(puyo rune, x int, y int) {
	ix := 0
	iy := 0
	switch puyo {
	case 'r':
		ix = 0
	case 'g':
		ix = 32
	case 'b':
		ix = 64
	case 'y':
		ix = 96
	case 'p':
		ix = 128
	case '.':
		return
	}

	point := image.Pt((x+1)*32, (y+1)*32)
	draw.Draw(c.Image, image.Rectangle{image.Pt(x*32, y*32), point}, c.PuyoImage, image.Pt(ix, iy), draw.Over)
}

func (c *Canvas) Export(name string) {
	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, c.Image)
}

func main() {
	width := flag.Int("width", 10, "width of canvas")
	height := flag.Int("height", 10, "height of canvas")
	flag.Parse()
	c := NewCanvas(*width, *height)
	var sc = bufio.NewScanner(os.Stdin)
	row := 0
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\n")
		for i, puyo := range line {
			c.PlacePuyo(puyo, i, row)
		}
		row++
	}
	c.Export(flag.Arg(0))
}
