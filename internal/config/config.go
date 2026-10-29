package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	InputPath  string
	OutputPath string
	MinChunks  int
}

func LoadConfig() (*Config, error) {
	// Load whichever env files exist. Missing ones are skipped.
	for _, f := range []string{".env.local", ".env"} {
		_ = godotenv.Load(f)
	}

	inputPath := os.Getenv("GOCRYPT_INPUT")
	if inputPath == "" {
		inputPath = "input.txt"
	}

	outputPath := os.Getenv("GOCRYPT_OUTPUT")
	if outputPath == "" {
		outputPath = "output.txt"
	}

	minChunks := os.Getenv("GOCRYPT_MIN_CHUNKS")
	if minChunks == "" {
		minChunks = "4"
	}

	v, err := strconv.Atoi(minChunks)
	if err != nil {
		return &Config{}, fmt.Errorf("GOCRYPT_MIN_CHUNKS must be a whole number, got %q", minChunks)
	} else if v < 2 {
		return &Config{}, fmt.Errorf("GOCRYPT_MIN_CHUNKS must be at least 2, got %d", v)
	}

	cfg := &Config{
		InputPath:  inputPath,
		OutputPath: outputPath,
		MinChunks:  v,
	}

	return cfg, nil
}
