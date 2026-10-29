package chunker

// Split cuts data into numChunks pieces of nearly equal size.
func Split(data []byte, numChunks int) [][]byte {
	if numChunks < 1 {
		numChunks = 1
	}
	if len(data) == 0 {
		return nil
	}
	if numChunks > len(data) {
		numChunks = len(data)
	}

	base := len(data) / numChunks
	extra := len(data) % numChunks

	chunks := make([][]byte, 0, numChunks)
	start := 0

	for i := 0; i < numChunks; i++ {
		size := base
		if i < extra {
			size++
		}
		end := start + size

		chunk := make([]byte, size)
		copy(chunk, data[start:end])
		chunks = append(chunks, chunk)

		start = end
	}
	return chunks
}

// Assemble glues the chunks back together, in the order given.
func Assemble(chunks [][]byte) []byte {
	total := 0
	for _, c := range chunks {
		total += len(c)
	}

	result := make([]byte, 0, total)
	for _, c := range chunks {
		result = append(result, c...)
	}
	return result
}
