package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"io"
	"log"
)

func EncryptChunk(chunk []byte, key []byte) ([]byte, error) {
	hasher := md5.New()
	hasher.Write(key)

	aesBlock, err := aes.NewCipher(hasher.Sum(nil))
	if err != nil {
		return nil, err
	}

	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcmInstance.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)

	cipheredChunk := gcmInstance.Seal(nonce, nonce, chunk, nil)

	return cipheredChunk, nil
}

func DecryptChunk(chunk []byte, key []byte) ([]byte, error) {
	hasher := md5.New()
	hasher.Write(key)

	aesBlock, err := aes.NewCipher(hasher.Sum(nil))
	if err != nil {
		return nil, err
	}

	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		log.Fatalln(err)
	}

	nonceSize := gcmInstance.NonceSize()
	nonce, cipheredText := chunk[:nonceSize], chunk[nonceSize:]

	originalText, err := gcmInstance.Open(nil, nonce, cipheredText, nil)
	if err != nil {
		log.Fatalln(err)
	}
	return originalText, nil
}
