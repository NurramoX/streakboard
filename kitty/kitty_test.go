package kitty

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"strings"
	"testing"
)

func TestDiacriticsTable(t *testing.T) {
	// The protocol defines exactly 297 row/column diacritics, in
	// strictly ascending codepoint order.
	if len(diacritics) != 297 {
		t.Fatalf("len = %d, want 297", len(diacritics))
	}
	for i := 1; i < len(diacritics); i++ {
		if diacritics[i] <= diacritics[i-1] {
			t.Fatalf("not ascending at index %d: %U <= %U", i, diacritics[i], diacritics[i-1])
		}
	}
}

func TestPlaceholder(t *testing.T) {
	got, err := Placeholder(1, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b[38;2;0;0;1m\U0010EEEE̅̅\U0010EEEE̅̍\x1b[39m\n" +
		"\x1b[38;2;0;0;1m\U0010EEEE̍̅\U0010EEEE̍̍\x1b[39m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPlaceholderIDByteOrder(t *testing.T) {
	got, err := Placeholder(0x123456, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "\x1b[38;2;18;52;86m") {
		t.Errorf("got %q, want truecolor foreground 18;52;86", got)
	}
}

func TestPlaceholderRejectsBadInput(t *testing.T) {
	cases := []struct {
		id         uint32
		rows, cols int
	}{
		{0, 1, 1},
		{MaxID + 1, 1, 1},
		{1, 0, 1},
		{1, 1, MaxSpan + 1},
	}
	for _, c := range cases {
		if _, err := Placeholder(c.id, c.rows, c.cols); err == nil {
			t.Errorf("Placeholder(%d, %d, %d): want error", c.id, c.rows, c.cols)
		}
	}
}

// chunks splits concatenated APC sequences and returns each sequence's
// control keys and payload.
func chunks(t *testing.T, b []byte) (keys, payloads []string) {
	t.Helper()
	for _, seq := range strings.SplitAfter(string(b), "\x1b\\") {
		if seq == "" {
			continue
		}
		body, ok := strings.CutPrefix(seq, "\x1b_G")
		if !ok {
			t.Fatalf("sequence %q does not start with APC G", seq)
		}
		body = strings.TrimSuffix(body, "\x1b\\")
		k, p, ok := strings.Cut(body, ";")
		if !ok {
			t.Fatalf("sequence %q has no payload separator", seq)
		}
		keys = append(keys, k)
		payloads = append(payloads, p)
	}
	return keys, payloads
}

func TestTransmitPNG(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	out, err := TransmitPNG(7, img)
	if err != nil {
		t.Fatal(err)
	}
	keys, payloads := chunks(t, out)
	if len(keys) != 1 {
		t.Fatalf("got %d chunks, want 1", len(keys))
	}
	if keys[0] != "a=t,f=100,q=2,i=7" {
		t.Errorf("keys = %q", keys[0])
	}
	raw, err := base64.StdEncoding.DecodeString(payloads[0])
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds() != img.Bounds() {
		t.Errorf("bounds = %v, want %v", decoded.Bounds(), img.Bounds())
	}
}

func TestTransmitPNGChunked(t *testing.T) {
	// Incompressible noise so the PNG payload spans several chunks.
	img := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	x := uint32(1)
	for i := range img.Pix {
		x = x*1664525 + 1013904223
		img.Pix[i] = uint8(x >> 24)
	}
	out, err := TransmitPNG(1, img)
	if err != nil {
		t.Fatal(err)
	}
	keys, payloads := chunks(t, out)
	if len(keys) < 2 {
		t.Fatalf("got %d chunks, want several", len(keys))
	}
	if keys[0] != "a=t,f=100,q=2,i=1,m=1" {
		t.Errorf("first chunk keys = %q", keys[0])
	}
	for i, k := range keys[1 : len(keys)-1] {
		if k != "m=1" {
			t.Errorf("middle chunk %d keys = %q", i+1, k)
		}
	}
	if last := keys[len(keys)-1]; last != "m=0" {
		t.Errorf("last chunk keys = %q", last)
	}
	var all strings.Builder
	for _, p := range payloads {
		if len(p) > chunkSize {
			t.Errorf("chunk payload %d bytes exceeds %d", len(p), chunkSize)
		}
		all.WriteString(p)
	}
	raw, err := base64.StdEncoding.DecodeString(all.String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(bytes.NewReader(raw)); err != nil {
		t.Fatalf("reassembled payload is not a valid png: %v", err)
	}
}

func TestPlaceAndDelete(t *testing.T) {
	place, err := Place(9, 7, 106)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(place), "\x1b_Ga=p,U=1,q=2,i=9,r=7,c=106\x1b\\"; got != want {
		t.Errorf("Place = %q, want %q", got, want)
	}
	if got, want := string(Delete(9)), "\x1b_Ga=d,d=I,q=2,i=9\x1b\\"; got != want {
		t.Errorf("Delete = %q, want %q", got, want)
	}
}
