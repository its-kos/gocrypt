package filechunk

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func Chunk(filePath string, chunkSize int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return make([]byte, 0), err
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return make([]byte, 0), err
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return make([]byte, 0), err
	}

	fmt.Printf("Original file size: %d bytes\n", stat.Size())
	fmt.Printf("Chunk size: %d bytes\n", chunkSize)
	fmt.Printf("File split into %d chunks\n", stat.Size()/int64(chunkSize))

	return bs, nil
}

func Restore() {

}
