package encryption

import "crypto/md5"

func EncryptChunk(chunk []byte) []byte {
	md5Hash := md5.Sum(chunk)
	return md5Hash
}

func DecryptChunk() []byte {
	decrypted := make([]byte, 4)
	return decrypted
}
