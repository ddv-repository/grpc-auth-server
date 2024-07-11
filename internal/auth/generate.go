package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Base62Encode(input []byte) string {
	var result []byte
	bi := new(big.Int).SetBytes(input)
	base := big.NewInt(int64(len(base62Chars)))
	mod := new(big.Int)
	for bi.Cmp(big.NewInt(0)) != 0 {
		bi.DivMod(bi, base, mod)
		result = append(result, base62Chars[mod.Int64()])
	}
	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

func Base62Decode(input string) ([]byte, error) {
	bi := big.NewInt(0)
	base := big.NewInt(int64(len(base62Chars)))
	for _, c := range input {
		index := int64(strings.IndexRune(base62Chars, c))
		if index == -1 {
			return nil, errors.New("invalid character in base62 string")
		}
		bi.Mul(bi, base)
		bi.Add(bi, big.NewInt(index))
	}
	return bi.Bytes(), nil
}

func Serialize(token AccessToken) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(token)
	return buf.Bytes(), err
}

func Deserialize(data []byte) (AccessToken, error) {
	var token AccessToken
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&token)
	return token, err
}

func SerializeJSON(token AccessToken) ([]byte, error) {
	return json.Marshal(token)
}

func DeserializeJSON(data []byte) (AccessToken, error) {
	var token AccessToken
	err := json.Unmarshal(data, &token)
	return token, err
}

func Encrypt(data []byte, passphrase string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func Decrypt(data []byte, passphrase string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(passphrase))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
