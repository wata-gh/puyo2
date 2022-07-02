package puyo2

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/bits"
	"os"
)

type ShapeBitField struct {
	Shapes []*FieldBits
}

func NewShapeBitField() *ShapeBitField {
	shapeBitField := new(ShapeBitField)
	return shapeBitField
}

func (sbf *ShapeBitField) drawField(puyo *image.Image, out *image.NRGBA) {
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

func (sbf *ShapeBitField) drawPuyo(fb *FieldBits, n int, x int, y int, puyo *image.Image, out *image.NRGBA) {
	ix := 0
	iy := 0
	u, d, r, l := false, false, false, false
	if fb.Onebit(x, y+1) > 0 {
		u = true
	}
	if fb.Onebit(x, y-1) > 0 {
		d = true
	}
	if x > 0 && fb.Onebit(x-1, y) > 0 {
		l = true
	}
	if x < 5 && fb.Onebit(x+1, y) > 0 {
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
	ix = n * 32
	point := image.Pt((x+2)*32, (14-y)*32)
	draw.Draw(out, image.Rectangle{image.Pt((x+1)*32, (13-y)*32), point}, *puyo, image.Pt(ix, iy), draw.Over)
}

func (sbf *ShapeBitField) AddShape(shape *FieldBits) {
	sbf.Shapes = append(sbf.Shapes, shape)
}

func (sbf *ShapeBitField) Drop(fb *FieldBits) {
	for _, shape := range sbf.Shapes {
		r0 := Extract(shape.m[0], ^fb.m[0])
		r1 := Extract(shape.m[1], ^fb.m[1])
		var dropmask1 [2]uint64
		for x := 0; x < 6; x++ {
			idx := x >> 2
			vc := bits.OnesCount64(fb.ColBits(x))
			dropmask1[idx] |= bits.RotateLeft64((1<<vc)-1, 14-vc) << (x & 3 * 16)
		}
		shape.m[0] = Deposit(r0, ^dropmask1[0])
		shape.m[1] = Deposit(r1, ^dropmask1[1])
	}
}

func (sbf *ShapeBitField) ExportImage(name string) {
	ffield, _ := os.Open("images/puyos.png")
	fpuyo, _ := os.Open("images/puyos_gray.png")
	defer ffield.Close()
	defer fpuyo.Close()

	field, _, _ := image.Decode(ffield)
	puyo, _, _ := image.Decode(fpuyo)
	out := image.NewNRGBA(image.Rectangle{image.Pt(0, 0), image.Pt(32*8, 32*14)})
	sbf.drawField(&field, out)

	for y := 13; y > 0; y-- {
		for x := 0; x < 6; x++ {
			for n, shape := range sbf.Shapes {
				if shape.Onebit(x, y) > 0 {
					sbf.drawPuyo(shape, n, x, y, &puyo, out)
				}
			}
		}
	}

	outfile, _ := os.Create(name)
	defer outfile.Close()
	png.Encode(outfile, out)
}

func (sbf *ShapeBitField) FindVanishingBits() *FieldBits {
	v := NewFieldBits()
	for _, shape := range sbf.Shapes {
		v = v.Or(shape.MaskField12().FindVanishingBits())
	}
	return v
}

func (sbf *ShapeBitField) ShapeCount() int {
	return len(sbf.Shapes)
}

func (sbf *ShapeBitField) Simulate1() bool {
	v := sbf.FindVanishingBits()
	if v.IsEmpty() {
		return false
	}
	sbf.Drop(v)
	return true
}

func (sbf *ShapeBitField) Simulate() *ShapeRensaResult {
	result := NewShapeRensaResult()
	for sbf.Simulate1() {
		sbf.ShowDebug()
		result.AddChain()
	}
	result.SetShapeBitField(sbf)
	return result
}

func (sbf *ShapeBitField) ShowDebug() {
	var s string
	for y := 14; y >= 0; y-- {
		s += fmt.Sprintf("%02d: ", y)
		for x := 0; x < 6; x++ {
			e := true
			for i, shape := range sbf.Shapes {
				if shape.Onebit(x, y) > 0 {
					s += fmt.Sprint(i)
					e = false
					break
				}
			}
			if e {
				s += "."
			}
		}
		s += fmt.Sprintln()
	}
	fmt.Println(s)
}
