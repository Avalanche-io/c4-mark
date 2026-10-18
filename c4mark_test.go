package c4mark

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"math/big"
	"os"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/cases.json")

// The c4 library's own test vectors (github.com/Avalanche-io/c4, id_test.go).
var vectors = []struct{ id, digest string }{
	{"c459dsjfscH38cYeXXYogktxf4Cd9ibshE3BHUo6a58hBXmRQdZrAkZzsWcbWtDg5oQstpDuni4Hirj75GEmTc1sFT",
		"cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
	{"c467rpwLCuS5DGA8KGZXKsVQ7dnPb9goRLoKfgGbLfQg9WoLUgNY77E2jT11fem3coV9nAkguBACzrU1iyZM4B8roQ", strings.Repeat("ff", 64)},
	{"c41111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", strings.Repeat("00", 64)},
}

func TestVectors(t *testing.T) {
	for _, v := range vectors {
		d, ok := Decode(v.id)
		if !ok || hex.EncodeToString(d) != v.digest {
			t.Errorf("%s: got %x, %v", v.id[:12], d, ok)
		}
	}
}

func TestNotIDs(t *testing.T) {
	for _, s := range []string{"", "hello", "c47" + strings.Repeat("1", 87), vectors[0].id[:60], vectors[0].id[:50] + "0" + vectors[0].id[51:]} {
		if Valid(s) {
			t.Errorf("%q is not a C4 ID", s)
		}
	}
}

// encode writes a digest as a C4 ID, for the test cases.
func encode(d []byte) string {
	n := new(big.Int).SetBytes(d)
	b := big.NewInt(58)
	m := new(big.Int)
	s := ""
	for n.Sign() > 0 {
		n.DivMod(n, b, m)
		s = string(b58[m.Int64()]) + s
	}
	return "c4" + strings.Repeat("1", 88-len(s)) + s
}

func digest(seed uint32) []byte {
	x := seed * 2654435761
	out := make([]byte, 64)
	for i := range out {
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		out[i] = byte(x)
	}
	return out
}

type Case struct {
	ID     string `json:"id"`
	Digest string `json:"digest,omitempty"`
	Glyphs int    `json:"glyphs"`
	Size   int    `json:"size,omitempty"`
	Ink    string `json:"ink,omitempty"`
	Gap    string `json:"gap,omitempty"`
	SVG    string `json:"svg"`
}

// cases is every input the two implementations must agree on.
func cases() []Case {
	ids := []string{vectors[0].id, vectors[1].id, vectors[2].id}
	for s := uint32(1); s <= 6; s++ {
		ids = append(ids, encode(digest(s*101)))
	}
	ids = append(ids, "not a C4 ID", "c47"+strings.Repeat("1", 87))
	var out []Case
	for i, id := range ids {
		for k := 0; k <= 3; k++ {
			c := Case{ID: id, Glyphs: k}
			switch i % 4 {
			case 1:
				c.Size = 16
			case 2:
				c.Size, c.Ink = 24, "#1a1a1a"
			case 3:
				c.Ink = `var(--ink, "x" & <y>)`
			}
			gaps := []string{""}
			if k >= 2 {
				gaps = []string{"", "none", "hairline"}
			}
			for _, g := range gaps {
				cc := c
				cc.Gap = g
				if d, ok := Decode(id); ok {
					cc.Digest = hex.EncodeToString(d)
				}
				cc.SVG = SVG(id, Options{Glyphs: cc.Glyphs, Size: cc.Size, Ink: cc.Ink, Gap: cc.Gap})
				out = append(out, cc)
			}
		}
	}
	return out
}

// TestCases checks SVG against testdata/cases.json, which the JavaScript
// tests check too: the two implementations write the same bytes.
func TestCases(t *testing.T) {
	got := cases()
	if *update {
		b, _ := json.MarshalIndent(got, "", " ")
		if err := os.WriteFile("testdata/cases.json", append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	b, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var want []Case
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != len(got) {
		t.Fatalf("%d cases, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("case %d (%s, %d glyphs, gap %q) differs", i, want[i].ID[:min(12, len(want[i].ID))], want[i].Glyphs, want[i].Gap)
		}
	}
}

func TestFoldCountsEveryBit(t *testing.T) {
	d := digest(7)
	for _, bit := range []int{0, 7, 255, 511} {
		e := append([]byte(nil), d...)
		e[bit>>3] ^= 1 << (7 - uint(bit&7))
		a, b := Fold(d, 30), Fold(e, 30)
		n := 0
		for i := range a {
			if a[i] != b[i] {
				n++
			}
		}
		if n != 1 {
			t.Errorf("bit %d changed %d folded bits", bit, n)
		}
	}
}
