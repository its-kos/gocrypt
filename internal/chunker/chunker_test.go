package chunker

import (
	"bytes"
	"testing"
)

func makeData(n int) []byte {
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i)
	}
	return data
}

func TestSplitAssembleRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		numChunks int
	}{
		{"even split", 12, 3},
		{"uneven split", 10, 3},
		{"one chunk", 10, 1},
		{"one byte per chunk", 10, 10},
		{"more chunks than bytes", 3, 8},
		{"zero chunks", 5, 0},
		{"negative chunks", 5, -2},
		{"larger input", 1000, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := makeData(tt.size)
			if got := Assemble(Split(data, tt.numChunks)); !bytes.Equal(got, data) {
				t.Fatal("Assemble(Split(data)) differs from data")
			}
		})
	}
}

func TestSplitChunkCount(t *testing.T) {
	tests := []struct {
		size, numChunks, want int
	}{
		{12, 3, 3},
		{10, 3, 3},
		{3, 8, 3},  // capped at one byte per chunk
		{5, 0, 1},  // fewer than 1 is treated as 1
		{5, -2, 1}, // same
	}
	for _, tt := range tests {
		if got := len(Split(makeData(tt.size), tt.numChunks)); got != tt.want {
			t.Errorf("Split(%d bytes, %d) gave %d chunks, want %d", tt.size, tt.numChunks, got, tt.want)
		}
	}
}

func TestSplitSizesAreBalanced(t *testing.T) {
	chunks := Split(makeData(10), 3)
	want := []int{4, 3, 3} // the remainder goes to the first chunks
	for i, c := range chunks {
		if len(c) != want[i] {
			t.Fatalf("chunk %d has %d bytes, want %d (all sizes: %v)", i, len(c), want[i], want)
		}
	}
}

func TestSplitEmpty(t *testing.T) {
	if got := Split(nil, 3); got != nil {
		t.Fatalf("Split(nil) = %v, want nil", got)
	}
	if got := Split([]byte{}, 3); got != nil {
		t.Fatalf("Split(empty) = %v, want nil", got)
	}
	if got := Assemble(nil); len(got) != 0 {
		t.Fatalf("Assemble(nil) has %d bytes, want 0", len(got))
	}
}

func TestSplitReturnsCopies(t *testing.T) {
	data := makeData(10)
	chunks := Split(data, 2)
	chunks[0][0] ^= 0xff
	if data[0] != 0 {
		t.Fatal("modifying a chunk changed the original data")
	}
}
