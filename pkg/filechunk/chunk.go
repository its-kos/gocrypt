package filechunk

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func Chunk(filePath string, chunkSize int) ([][]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(fs)
	if err != nil && err != io.EOF {
		return nil, err
	}

	fmt.Printf("Original file size: %d bytes\n", stat.Size())
	fmt.Printf("Chunk size: %d bytes\n", chunkSize)
	fmt.Printf("File split into %d chunks\n", stat.Size()/int64(chunkSize))

	// Cant use array cause we dont know the size ahead of time.
	// Would be nice, more efficient since the size is constant.
	// var bArray [stat.Size()/int64(chunkSize)]byte
	bSlice := make([][]byte, 0)

	if stat.Size() < int64(chunkSize) {
		return append(bSlice, fs), nil
	} else {
		for i := 0; i < int(stat.Size()); i += chunkSize {
			end := i + chunkSize
			if end > len(fs) {
				end = len(fs)
			}
			bSlice = append(bSlice, fs[i:end])
		}
		return bSlice, nil
	}
}

// func Restore(chunks []byte) os.File {
// 	return *os.NewFile(nil, "whatever")
// }
