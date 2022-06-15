package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type SecretSize int

const (
	// AES-128 = 128 bits = 16 bytes (8*2)
	SecretAES128 SecretSize = 8
	// AES-192 = 192 bits = 24 bytes (12*2)
	SecretAES192 SecretSize = 12
	// AES-256 = 256 bits = 32 bytes (16*2)
	SecretAES256 SecretSize = 16
)

// Cryptography helps encrypting and decrypting private data which is often
// required to be exposed externally.
type Cryptography struct {
	secret string
}

// New returns `Cryptography` type. It accepts an application wide `secret`
// which will be used as the "fallback" value if the methods are called with a
// nil `secret` argument.
func New(secret string) Cryptography {
	return Cryptography{
		secret: secret,
	}
}

// Secret generates a random bytes key in given size. The key size must be
// either 16, 24 or 32 bytes to select AES-128, AES-192 or AES-256 modes
// respectively. Depending on the requiremenets, this function can be used only
// once to create an application wide secret key then stored as an environment
// variable or repeatedly for each encryption/decryption. You can dump the
// generated secret with with `fmt.Printf("%x", byte)` method.
func (c Cryptography) Secret(size SecretSize) ([]byte, error) {
	key := make([]byte, size)

	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

// EncryptAsString returns an encrypted string version of the given data using a
// secret key. Decryption requires the same secret key and the `DecryptString`
// method.
func (c Cryptography) EncryptAsString(data, secret []byte) (string, error) {
	if secret == nil {
		secret = []byte(c.secret)
	}

	val, _, err := c.encrypt(data, secret, true)
	if err != nil {
		return "", err
	}

	return val, nil
}

// EncryptAsByte returns an encrypted byte version of the given data using a
// secret key. Decryption requires the same secret key and the `DecryptByte`
// method.
func (c Cryptography) EncryptAsByte(data, secret []byte) ([]byte, error) {
	if secret == nil {
		secret = []byte(c.secret)
	}

	_, val, err := c.encrypt(data, secret, false)
	if err != nil {
		return nil, err
	}

	return val, nil
}

// encrypt returns an encrypted version of the given data using a secret key.
// Decryption requires the same secret key.
func (c Cryptography) encrypt(data, secret []byte, isString bool) (string, []byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", nil, fmt.Errorf("new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil, fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", nil, fmt.Errorf("io read: %w", err)
	}

	byte := aead.Seal(nonce, nonce, data, nil)

	if isString {
		return base64.URLEncoding.EncodeToString(byte), nil, nil
	}

	return "", byte, nil
}

// DecryptString returns an decrypted byte version of the given string data
// using a secret key. It requires the same secret key as the `EncryptAsString`
// method used.
func (c Cryptography) DecryptString(data string, secret []byte) ([]byte, error) {
	if secret == nil {
		secret = []byte(c.secret)
	}

	byte, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("decode string: %w", err)
	}

	return c.decrypt(byte, secret)
}

// DecryptByte returns an decrypted byte version of the given byte data using
// a secret key. It requires the same secret key as the `EncryptAsByte` method
// used.
func (c Cryptography) DecryptByte(data, secret []byte) ([]byte, error) {
	if secret == nil {
		secret = []byte(c.secret)
	}

	return c.decrypt(data, secret)
}

// decrypt returns a decrypted version of the given data using a secret key. It
// requires the same secret key as the relevant encryption method used.
func (c Cryptography) decrypt(data, secret []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	size := aead.NonceSize()
	if len(data) < size {
		return nil, fmt.Errorf("nonce size: invalid length")
	}

	nonce, text := data[:size], data[size:]

	res, err := aead.Open(nil, nonce, text, nil)
	if err != nil {
		return nil, fmt.Errorf("aead open: %w", err)
	}

	return res, nil
}
