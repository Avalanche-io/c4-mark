# c4-mark

A C4 ID drawn as a small mark instead of 90 characters.

A [C4 ID](https://github.com/Avalanche-io/c4), standardized as SMPTE ST 2114,
names content by its SHA-512 digest: `c4` and 88 base58 characters. Shown
raw, it is noise. c4-mark draws it as the c4 logo in corner brackets. Hover
or focus the mark to read the ID, click it or press Enter to copy it. Where
people check that IDs match, up to three glyphs follow the logo: pictures of
the digest, so IDs that differ look different at a glance.

Zero dependencies: one ES module, one CSS file, and a Go package that draws
the same SVG byte for byte.

```html
<link rel="stylesheet" href="c4mark.css">
<script type="module" src="element.js"></script>

Output <c4-id value="c459dsjfscH38cYeXXYogktxf4Cd9ibshE3BHUo6a58hBXmRQdZrAkZzsWcbWtDg5oQstpDuni4Hirj75GEmTc1sFT"></c4-id>
matches <c4-id value="c459dsjf…" glyphs="2" tail="4" label="the delivered view"></c4-id>
```

```js
import { svg, markHTML, install, decode, isValid } from './c4mark.js';

svg(id);                               // the bare mark, an SVG string
svg(id, { glyphs: 2, size: 24, ink: '#222' });
markHTML(id, { glyphs: 2, tail: 4, label: 'output view' }); // for templates
install(document);                     // hover, focus, copy, touch, for every mark
decode(id);                            // the 64-byte digest, or null
```

```go
import "github.com/Avalanche-io/c4-mark"

s := c4mark.SVG(id, c4mark.Options{Glyphs: 2, Size: 24})
```

## The API

| JavaScript                               | Go                         | what                                                                                       |
| ---------------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------ |
| `decode(id)`                             | `Decode(id)`               | the 64-byte digest, or null (false) when the string is not a C4 ID                         |
| `isValid(id)`                            | `Valid(id)`                | whether it is one: `c4`, 88 base58 digits, a value under 2^512                             |
| `fold(digest, w)`                        | `Fold(digest, w)`          | the 512 bits XOR-folded into w                                                             |
| `glyphBits(id, k)`                       | `GlyphBits(id, k)`         | 15 bits for each of k glyphs                                                               |
| `tail(id, n)`, `abbrev(id)`              | `Tail`, `Abbrev`           | the last n characters; `c4…` and the last eight                                            |
| `svg(id, { glyphs, size, ink, gap })`    | `SVG(id, Options{...})`    | the mark: 0 to 3 glyphs, height in px (or 1em), ink color, gap `default`, `none`, `hairline` |
| `markHTML(id, { glyphs, tail, label, gap })` |                        | a focusable `<span>` for templates, with a bare fallback for small sizes                   |
| `install(root)`                          |                            | the behavior, delegated: every mark under root, now and later                              |
| `defineElement(name)`, `element.js`      |                            | `<c4-id value glyphs tail label gap>`                                                      |
| `odds(k)`                                | `Odds(k)`                  | 2^(15k): the states of k glyphs                                                            |

The element's attribute is `value`, not `id`: an element's `id` is HTML's
own, and two marks of the same C4 ID on a page would collide.

**Behavior.** Hover or focus shows the ID as `c4…` and its last eight
characters, with the full ID in the tooltip. Click, Enter or Space copies the
full ID and says "copied" ("copy failed" if the clipboard refuses). On touch
a first tap shows the ID and a Copy button. Escape hides it. A string that is
not a C4 ID draws with dashed brackets and says "not a valid C4 ID".

**Theming.** Nothing is colored by the package: the mark is drawn in the
text's color (`--c4-mark-ink`), brighter on hover and focus
(`--c4-mark-ink-hover`), and the popover takes `--c4-mark-pop-bg`,
`--c4-mark-pop-ink`, `--c4-mark-pop-line` and `--c4-mark-mono`.

## Why the glyphs read the digest

A C4 ID is the 512-bit digest written in base58, padded to 88 digits. 2^512
is about 5.07 × 58^87, so the first digit after `c4` is one of 1 to 6 (one ID
in five starts `c41`), and the next few are skewed. Reading the leading
characters, or drawing pictures from them, makes different IDs look alike.

So the glyphs decode the ID back to its 512 bits and fold them: with k
glyphs, digest bit i is XORed into bit i mod 15k, and glyph g shows bits 15g
to 15g+14 as a 5 × 5 grid, mirrored left to right (5 rows of 3 columns, the
outer two mirrored). Every digest bit counts in exactly one cell, so a
difference anywhere in an ID changes the picture. SHA-512 bits are uniform,
so the odds that two different IDs draw the same are exact:

| glyphs | bits | two different IDs look the same | any two of 10 IDs on screen |
| ------ | ---- | ------------------------------- | --------------------------- |
| 1      | 15   | 1 in 32,768                     | about 1 in 730              |
| 2      | 30   | 1 in 1.07 billion               | about 1 in 24 million       |
| 3      | 45   | 1 in 35 trillion                | about 1 in 780 billion      |

Text that tells IDs apart shows the tail (`c4…k9QxT2ab`) for the same
reason: the last digits are uniform, the first are not.

**Size.** A glyph cell is 1/7 of the mark's height, so below 16 px the grid
turns to noise. `markHTML` and `<c4-id>` draw the glyphs only at a font size
of 16 px or more (`MIN_GLYPH_PX`) and fall back to the bare mark below it.
`svg()` draws what it is asked.

## Tests

```sh
node --test test/*.test.mjs   # the c4 library's vectors, and the Go package's cases
go test ./...                 # the same, and testdata/cases.json (-update rewrites it)
```

Both implementations are checked against `testdata/cases.json`: every
combination of ID, glyph count, size, ink and gap the Go package writes,
which the JavaScript module must match byte for byte.

## License

Copyright 2026 Joshua Ian Kolden. Apache License 2.0 (see LICENSE).
