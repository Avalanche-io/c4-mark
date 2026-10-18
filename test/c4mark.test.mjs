// node --test: the JavaScript module against the c4 library's vectors and
// against testdata/cases.json, which the Go package writes and checks: the
// two implementations draw the same bytes.

import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { decode, isValid, fold, glyphBits, tail, abbrev, odds, svg, markHTML, width, MIN_GLYPH_PX } from '../c4mark.js';

const hex = (b) => Buffer.from(b).toString('hex');
const VECTORS = [
  ['c459dsjfscH38cYeXXYogktxf4Cd9ibshE3BHUo6a58hBXmRQdZrAkZzsWcbWtDg5oQstpDuni4Hirj75GEmTc1sFT',
    'cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e'],
  ['c467rpwLCuS5DGA8KGZXKsVQ7dnPb9goRLoKfgGbLfQg9WoLUgNY77E2jT11fem3coV9nAkguBACzrU1iyZM4B8roQ', 'ff'.repeat(64)],
  ['c41111111111111111111111111111111111111111111111111111111111111111111111111111111111111111', '00'.repeat(64)],
];
const [A] = VECTORS[0];

test('decode gives the digests the c4 library gives', () => {
  for (const [id, d] of VECTORS) assert.equal(hex(decode(id)), d);
});

test('strings that are not C4 IDs', () => {
  for (const s of ['', 'hello', null, 'c47' + '1'.repeat(87), A.slice(0, 60), A.slice(0, 50) + '0' + A.slice(51)]) {
    assert.equal(decode(s), null);
    assert.equal(isValid(s), false);
  }
});

test('the same bytes as the Go package, for every case', () => {
  const cases = JSON.parse(readFileSync(new URL('../testdata/cases.json', import.meta.url), 'utf8'));
  assert.ok(cases.length > 50);
  for (const c of cases) {
    if (c.digest) assert.equal(hex(decode(c.id)), c.digest);
    const opts = { glyphs: c.glyphs };
    if (c.size) opts.size = c.size;
    if (c.ink) opts.ink = c.ink;
    if (c.gap) opts.gap = c.gap;
    assert.equal(svg(c.id, opts), c.svg, `${c.id.slice(0, 12)} glyphs ${c.glyphs} gap ${c.gap || 'default'}`);
  }
});

test('the fold counts every digest bit exactly once', () => {
  const d = decode(A);
  for (const bit of [0, 7, 255, 511]) {
    const e = Uint8Array.from(d); e[bit >> 3] ^= 1 << (7 - (bit & 7));
    const a = fold(d, 45), b = fold(e, 45);
    assert.equal(a.filter((v, i) => v !== b[i]).length, 1);
  }
  assert.equal(glyphBits(A, 3).length, 3);
  assert.equal(glyphBits('nope', 2), null);
});

test('a change anywhere in an ID changes its glyphs', () => {
  const B58 = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';
  for (const pos of [3, 30, 60, 89]) {
    const t = A.split(''); t[pos] = B58[(B58.indexOf(t[pos]) + 1) % 58];
    const id = t.join('');
    if (!isValid(id)) continue;
    assert.notEqual(svg(id, { glyphs: 2 }), svg(A, { glyphs: 2 }), `position ${pos}`);
  }
});

test('text shows the tail, never the characters after c4', () => {
  assert.equal(tail(A, 4), A.slice(-4));
  assert.equal(abbrev(A), `c4…${A.slice(-8)}`);
  assert.match(markHTML(A, { tail: 4 }), new RegExp(`<span class="c4-mark-tail">…${A.slice(-4)}</span>`));
});

test('the mark in HTML: focusable, labelled, glyphs with a bare fallback', () => {
  const h = markHTML(A, { glyphs: 2, label: 'output view' });
  assert.match(h, /tabindex="0"/);
  assert.match(h, /role="button"/);
  assert.ok(h.includes(`aria-label="output view ${abbrev(A)}. Press Enter to copy."`));
  assert.match(h, /c4-mark-full/);
  assert.match(h, /c4-mark-bare/);
  assert.equal(MIN_GLYPH_PX, 16);
  const bad = markHTML('c4nope', { glyphs: 2 });
  assert.match(bad, /c4-mark-invalid/);
  assert.match(bad, /not a valid C4 ID/);
  assert.match(bad, /stroke-dasharray/);
  assert.doesNotMatch(bad, /c4-mark-full/);
});

test('ink only, and the ink is a parameter', () => {
  for (const k of [0, 1, 2, 3]) assert.doesNotMatch(svg(A, { glyphs: k }), /#[0-9a-f]{3,6}\b|rgb\(/i);
  assert.match(svg(A, { glyphs: 1, ink: '#222' }), /stroke="#222"/);
  assert.match(svg(A, { glyphs: 1, ink: '#222' }), /fill="#222"/);
});

test('sizes and gaps', () => {
  assert.equal(width(0), 1600);
  assert.equal(width(2, 'none'), 4000);
  assert.equal(width(2), 4160);
  assert.match(svg(A, { glyphs: 3, size: 16 }), /width="54.40" height="16"/);
  assert.equal(odds(2), 2 ** 30);
});
