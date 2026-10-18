// Package c4mark draws a C4 ID as a small mark: the c4 logo in corner
// brackets, optionally followed by up to three glyphs drawn from the ID's
// SHA-512 digest. It writes the same SVG, byte for byte, as the JavaScript
// module in this repository (c4mark.js), so server-side tools and browsers
// draw identical marks.
//
// Copyright 2026 Joshua Ian Kolden. Apache License 2.0.
package c4mark

import (
	"math/big"
	"strconv"
	"strings"
)

// Version is the package version, the same as the JavaScript module's.
const Version = "0.1.0"

const b58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var limit = new(big.Int).Lsh(big.NewInt(1), 512)

// Decode returns the 64-byte digest a C4 ID names, and false when the
// string is not a C4 ID ("c4" and 88 base58 digits of a value under 2^512).
func Decode(id string) ([]byte, bool) {
	if len(id) != 90 || !strings.HasPrefix(id, "c4") {
		return nil, false
	}
	n := new(big.Int)
	b := big.NewInt(58)
	for i := 2; i < 90; i++ {
		v := strings.IndexByte(b58, id[i])
		if v < 0 {
			return nil, false
		}
		n.Mul(n, b)
		n.Add(n, big.NewInt(int64(v)))
	}
	if n.Cmp(limit) >= 0 {
		return nil, false
	}
	out := make([]byte, 64)
	n.FillBytes(out)
	return out, true
}

// Valid reports whether id is a C4 ID.
func Valid(id string) bool {
	_, ok := Decode(id)
	return ok
}

// Fold XOR-folds a digest's 512 bits into w bits: digest bit i, counting
// from the most significant bit of the first byte, lands in bit i mod w.
func Fold(digest []byte, w int) []byte {
	out := make([]byte, w)
	for i := 0; i < 512; i++ {
		if (digest[i>>3]>>(7-uint(i&7)))&1 == 1 {
			out[i%w] ^= 1
		}
	}
	return out
}

// GlyphBits returns the bits of each of k glyphs (15 each) for an ID.
func GlyphBits(id string, k int) ([][]byte, bool) {
	d, ok := Decode(id)
	if !ok {
		return nil, false
	}
	f := Fold(d, 15*k)
	out := make([][]byte, k)
	for g := range out {
		out[g] = f[15*g : 15*g+15]
	}
	return out, true
}

// Tail is the ID's last n characters: text that tells IDs apart shows the
// tail, never the characters after "c4", which are skewed.
func Tail(id string, n int) string {
	if n >= len(id) {
		return id
	}
	return id[len(id)-n:]
}

// Abbrev is an ID as a person reads it: c4… and the last eight characters.
func Abbrev(id string) string {
	if len(id) > 16 {
		return "c4…" + id[len(id)-8:]
	}
	return id
}

// Odds is the number of states k glyphs have: two different IDs draw the
// same with probability 1/Odds(k).
func Odds(k int) float64 {
	o := 1.0
	for i := 0; i < 15*k; i++ {
		o *= 2
	}
	return o
}

// Gaps between glyphs, in grid units: default tiles, none one continuous
// field, hairline a thin seam.
var Gaps = map[string]int{"default": 160, "none": 0, "hairline": 48}

func gapOf(g string) int {
	if g == "0" {
		return 0
	}
	if v, ok := Gaps[g]; ok {
		return v
	}
	return Gaps["default"]
}

const (
	h     = 1600
	x0    = 1520
	gw    = 1120
	y0    = 240
	cell  = 224
	right = 240
	logo  = `<path d="M736 567A275 275 0 1 0 736 1033" fill="none" stroke="INK" stroke-width="150" stroke-linecap="round"/>` +
		`<path d="M1200 540V1060M1200 540L950 930H1300" fill="none" stroke="INK" stroke-width="150" stroke-linejoin="round" stroke-linecap="round"/>`
)

// Width is the viewBox width of a mark with k glyphs.
func Width(k int, gap string) int {
	if k <= 0 {
		return h
	}
	return x0 + k*gw + (k-1)*gapOf(gap) + right
}

func hundredths(v int) string {
	return strconv.Itoa(v/100) + "." + pad2(v%100)
}

func pad2(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}
	return strconv.Itoa(v)
}

func esc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// Options for SVG.
type Options struct {
	Glyphs int    // 0 (the bare mark) to 3
	Size   int    // the height in px; 0 for 1em
	Ink    string // the color of every stroke and cell; "" for currentColor
	Gap    string // between glyphs: "default", "none" or "hairline"
}

// SVG draws the mark. A string that is not a C4 ID draws the bare mark with
// dashed brackets.
func SVG(id string, o Options) string {
	valid := Valid(id)
	k := 0
	if valid {
		k = max(0, min(3, o.Glyphs))
	}
	ink := o.Ink
	if ink == "" {
		ink = "currentColor"
	}
	gap := o.Gap
	if gap == "" {
		gap = "default"
	}
	sp := gapOf(gap)
	w := Width(k, gap)
	hw := w / 16
	c := esc(ink)
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 `)
	b.WriteString(strconv.Itoa(w))
	b.WriteString(` 1600" `)
	if o.Size > 0 {
		b.WriteString(`width="` + hundredths(o.Size*hw) + `" height="` + strconv.Itoa(o.Size) + `"`)
	} else {
		b.WriteString(`width="` + hundredths(hw) + `em" height="1em"`)
	}
	b.WriteString(` aria-hidden="true" focusable="false">`)
	b.WriteString(`<path d="M100 500V100h400M` + strconv.Itoa(w-500) + ` 100h400v400M` + strconv.Itoa(w-100) +
		` 1100v400h-400M500 1500H100v-400" fill="none" stroke="` + c + `" stroke-width="140"`)
	if !valid {
		b.WriteString(` stroke-dasharray="160 140"`)
	}
	b.WriteString(`/>`)
	b.WriteString(strings.ReplaceAll(logo, "INK", c))
	if k > 0 {
		bits, _ := GlyphBits(id, k)
		b.WriteString(`<g fill="` + c + `">`)
		for g := 0; g < k; g++ {
			gx := x0 + g*(gw+sp)
			for r := 0; r < 5; r++ {
				for col := 0; col < 3; col++ {
					if bits[g][r*3+col] == 0 {
						continue
					}
					xs := []int{col, 4 - col}
					if col == 2 {
						xs = []int{2}
					}
					for _, x := range xs {
						b.WriteString(`<rect x="` + strconv.Itoa(gx+x*cell) + `" y="` + strconv.Itoa(y0+r*cell) +
							`" width="224" height="224"/>`)
					}
				}
			}
		}
		b.WriteString(`</g>`)
	}
	b.WriteString(`</svg>`)
	return b.String()
}
