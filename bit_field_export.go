package puyo2

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

func (bf *BitField) drawHand(hand *Hand, puyo *image.Image, out *image.NRGBA) {
	ix := bf.puyoImagePosX(hand.PuyoSet.Axis)
	rect := image.Rectangle{
		Min: image.Pt((hand.Position[0]+1)*32, 32),
		Max: image.Pt((hand.Position[0]+2)*32, 64),
	}
	draw.Draw(out, rect, *puyo, image.Pt(ix, 0), draw.Over)

	xoffset := 0
	yoffset := -32
	switch hand.Position[1] {
	case 0:
	case 1:
		yoffset = 0
		xoffset = 32
	case 2:
		yoffset = 32
	case 3:
		yoffset = 0
		xoffset = -32
	}
	rect.Min.X += xoffset
	rect.Min.Y += yoffset
	rect.Max.X = rect.Min.X + 32
	rect.Max.Y = rect.Min.Y + 32
	ix = bf.puyoImagePosX(hand.PuyoSet.Child)
	draw.Draw(out, rect, *puyo, image.Pt(ix, 0), draw.Over)
}

func (bf *BitField) drawField(puyo *image.Image, out *image.NRGBA) {
	for y := 13; y >= 0; y-- {
		draw.Draw(out, image.Rectangle{image.Pt(0, (13-y)*32), image.Pt(32, (14-y)*32)}, *puyo, image.Pt(5*32, 0), draw.Src)
		draw.Draw(out, image.Rectangle{image.Pt(32*7, (13-y)*32), image.Pt(32*8, (14-y)*32)}, *puyo, image.Pt(5*32, 0), draw.Src)
		for x := 0; x < 8; x++ {
			// block
			if (x == 0 || x == 7 || y == 0) && y != 13 {
				draw.Draw(out, image.Rectangle{image.Pt(x*32, 13*32), image.Pt((x+1)*32, 14*32)}, *puyo, image.Pt(5*32, 0), draw.Src)
			} else { // background
				draw.Draw(out, image.Rectangle{image.Pt(x*32, (13-y)*32), image.Pt((x+1)*32, (14-y)*32)}, *puyo, image.Pt(5*32, 32), draw.Src)
			}
		}
	}
}

func (bf *BitField) drawFieldAndPuyo(puyo *image.Image) *image.NRGBA {
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*14)})
	bf.drawField(puyo, out)

	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			c := bf.Color(x, y)
			if c == Empty {
				continue
			}
			if c == Ojama {
				point := image.Pt((x+2)*32, (14-y)*32)
				draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, *puyo, image.Pt(5*32, 2*32), draw.Over)
			} else {
				bf.drawPuyo(c, x, y, puyo, out)
			}
		}
	}

	return out
}

func (bf *BitField) drawFieldAndPuyoWithVanish(puyo *image.Image, vanish *FieldBits) *image.NRGBA {
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*14)})
	bf.drawField(puyo, out)

	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			c := bf.Color(x, y)
			if c == Empty {
				continue
			}
			if vanish.Onebit(x, y) == 0 {
				if c == Ojama {
					point := image.Pt((x+2)*32, (14-y)*32)
					draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, *puyo, image.Pt(5*32, 2*32), draw.Over)
				} else {
					bf.drawPuyo(c, x, y, puyo, out)
				}
			} else { // vanishing puyo
				point := image.Pt((x+2)*32, (14-y)*32)
				pos := 10
				switch c {
				case Red:
					pos = 5
				case Green:
					pos = 6
				case Blue:
					pos = 7
				case Yellow:
					pos = 8
				case Purple:
					pos = 9
				}
				draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, *puyo, image.Pt(5*32, pos*32), draw.Over)
			}
		}
	}
	return out
}

func (bf *BitField) puyoImagePosX(c Color) int {
	switch c {
	case Red:
		return 0
	case Green:
		return 32
	case Blue:
		return 64
	case Yellow:
		return 96
	case Purple:
		return 128
	default:
		return 160
	}
}

func (bf *BitField) drawPuyo(c Color, x int, y int, puyo *image.Image, out *image.NRGBA) {
	ix := 0
	iy := 0
	u, d, r, l := false, false, false, false
	if bf.Color(x, y+1) == c {
		u = true
	}
	if bf.Color(x, y-1) == c {
		d = true
	}
	if x > 0 && bf.Color(x-1, y) == c {
		l = true
	}
	if x < 5 && bf.Color(x+1, y) == c {
		r = true
	}
	if u == false && d == false && l == false && r == false {
		iy = 0
	} else if u && d == false && l == false && r == false {
		iy = 32
	} else if u == false && d && l == false && r == false {
		iy = 64
	} else if u && d && l == false && r == false {
		iy = 96
	} else if u == false && d == false && l && r == false {
		iy = 128
	} else if u && d == false && l && r == false {
		iy = 160
	} else if u == false && d && l && r == false {
		iy = 192
	} else if u && d && l && r == false {
		iy = 224
	} else if u == false && d == false && l == false && r {
		iy = 256
	} else if u && d == false && l == false && r {
		iy = 288
	} else if u == false && d && l == false && r {
		iy = 320
	} else if u && d && l == false && r {
		iy = 352
	} else if u == false && d == false && l && r {
		iy = 384
	} else if u && d == false && l && r {
		iy = 416
	} else if u == false && d && l && r {
		iy = 448
	} else {
		iy = 480
	}
	switch c {
	case Red:
		ix = 0
	case Blue:
		ix = 64
	case Green:
		ix = 32
	case Yellow:
		ix = 96
	case Purple:
		ix = 128
	case Empty:
		return
	}
	point := image.Pt((x+2)*32, (14-y)*32)
	draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, *puyo, image.Pt(ix, iy), draw.Over)
}

func (bf *BitField) loadPuyoImage() image.Image {
	puyo2Config := os.Getenv("PUYO2_CONFIG")
	path := "images/puyos.png"
	if puyo2Config != "" {
		path = fmt.Sprintf("%s/puyos.png", puyo2Config)
	}
	fpuyo, _ := os.Open(path)
	defer fpuyo.Close()
	puyo, _, _ := image.Decode(fpuyo)
	return puyo
}

func (bf *BitField) loadTranparentPuyoImage() image.Image {
	puyo2Config := os.Getenv("PUYO2_CONFIG")
	path := "images/puyos_transparent.png"
	if puyo2Config != "" {
		path = fmt.Sprintf("%s/puyos_transparent.png", puyo2Config)
	}
	fpuyo, _ := os.Open(path)
	defer fpuyo.Close()
	puyo, _, _ := image.Decode(fpuyo)
	return puyo
}

func (bf *BitField) ExportImage(name string) {
	puyo := bf.loadPuyoImage()
	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, bf.drawFieldAndPuyo(&puyo))
}

func (bf *BitField) ExportHandImage(name string, hand *Hand) {
	puyo := bf.loadPuyoImage()
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*17)})
	fieldAndPuyo := bf.drawFieldAndPuyo(&puyo)
	if hand != nil {
		bf.drawHand(hand, &puyo, out)
	}

	rect := fieldAndPuyo.Rect
	rect.Min.Y += 32 * 3
	rect.Max.Y += 32 * 3
	draw.Draw(out, rect, fieldAndPuyo, image.Pt(0, 0), draw.Over)

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}

func (obf *BitField) ExportSimulateImage(path string) {
	bf := obf.Clone()
	idx := 1
	bf.ExportImage(fmt.Sprintf("%s/%d.png", path, idx))
	for {
		v := bf.FindVanishingBits()
		if v.IsEmpty() {
			return
		}
		idx++
		bf.ExportImageWithVanish(fmt.Sprintf("%s/%d.png", path, idx), v)
		bf.Drop(v)
		idx++
		bf.ExportImage(fmt.Sprintf("%s/%d.png", path, idx))
	}
}

func (obf *BitField) ExportHandsSimulateImage(hands []Hand, path string) {
	os.Mkdir(path, 0755)
	bf := obf.Clone()
	idx := 1
	bf.ExportHandImage(fmt.Sprintf("%s/%d.png", path, idx), nil)
	beforeDrop := bf.MattulwanEditorParam()
	bf.Drop(bf.Bits(Empty).MaskField12())
	if beforeDrop != bf.MattulwanEditorParam() {
		idx++
		bf.ExportHandImage(fmt.Sprintf("%s/%d.png", path, idx), nil)
	}
	for _, hand := range hands {
		idx++
		bf.ExportHandImage(fmt.Sprintf("%s/%d.png", path, idx), &hand)
		placed, _ := bf.placePuyo(hand.PuyoSet, hand.Position)
		if placed == false {
			bf.ShowDebug()
			fmt.Printf("hand %v\n", hand)
			panic("can not place puyo.")
		}
		for {
			v := bf.FindVanishingBits()
			if v.IsEmpty() {
				break
			}
			idx++
			bf.ExportHandImageWithVanish(fmt.Sprintf("%s/%d.png", path, idx), v, nil)
			bf.Drop(v)
			idx++
			bf.ExportHandImage(fmt.Sprintf("%s/%d.png", path, idx), nil)
		}
	}
}

func (bf *BitField) ExportOnlyPuyoImage(name string) {
	puyo := bf.loadPuyoImage()
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*14)})

	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			c := bf.Color(x, y)
			if c == Empty {
				continue
			}
			if c == Ojama {
				point := image.Pt((x+2)*32, (14-y)*32)
				draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, puyo, image.Pt(5*32, 2*32), draw.Over)
			} else {
				bf.drawPuyo(c, x, y, &puyo, out)
			}
		}
	}

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}

func (bf *BitField) ExportImageWithVanish(name string, vanish *FieldBits) {
	puyo := bf.loadPuyoImage()
	out := bf.drawFieldAndPuyoWithVanish(&puyo, vanish)

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}

func (bf *BitField) ExportHandImageWithVanish(name string, vanish *FieldBits, hand *Hand) {
	puyo := bf.loadPuyoImage()
	fieldAndPuyo := bf.drawFieldAndPuyoWithVanish(&puyo, vanish)
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*17)})
	if hand != nil {
		bf.drawHand(hand, &puyo, out)
	}

	rect := fieldAndPuyo.Rect
	rect.Min.Y += 32 * 3
	rect.Max.Y += 32 * 3
	draw.Draw(out, rect, fieldAndPuyo, image.Pt(0, 0), draw.Over)

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}

func (bf *BitField) ExportImageWithTransparent(name string, trans *FieldBits) {
	puyo := bf.loadPuyoImage()
	puyot := bf.loadTranparentPuyoImage()
	var dpuyo image.Image
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*14)})
	bf.drawField(&puyo, out)

	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			c := bf.Color(x, y)
			if c == Empty {
				continue
			}
			if trans.Onebit(x, y) == 0 {
				dpuyo = puyo
			} else {
				dpuyo = puyot
			}
			if c == Ojama {
				point := image.Pt((x+2)*32, (14-y)*32)
				draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, dpuyo, image.Pt(5*32, 2*32), draw.Over)
			} else {
				bf.drawPuyo(c, x, y, &dpuyo, out)
			}
		}
	}

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}
