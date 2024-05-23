package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const secretkey = "neooshka"

func generateRandom(size int) (string, error) {

	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func getAesCipher() (cipher.AEAD, []byte, error) {

	key := sha256.Sum256([]byte(secretkey))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, nil, err
	}

	return aesgcm, key[7 : 7+aesgcm.NonceSize()], err
}

func GenerateToken() (string, error) {

	randID, err := generateRandom(16)
	if err != nil {
		return "", err
	}

	aesgcm, nonce, err := getAesCipher()
	if err != nil {
		return "", err
	}

	token := aesgcm.Seal(nil, nonce, []byte(randID), nil)

	return hex.EncodeToString(token), nil
}

func GetUserID(token string) (string, error) {

	aesgcm, nonce, err := getAesCipher()
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(token)
	if err != nil {
		return "", err
	}

	userID, err := aesgcm.Open(nil, nonce, ciphertext, nil)

	return string(userID), err
}
