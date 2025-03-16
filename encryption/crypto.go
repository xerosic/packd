package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

// EncryptWithRSA encrypts data using the provided RSA public key
func EncryptWithRSA(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return nil, errors.New("public key is nil")
	}

	// Check if data is too large
	maxSize := publicKey.Size() - 2*sha256.Size - 2
	if len(data) > maxSize {
		return nil, fmt.Errorf("data too large for RSA encryption: %d bytes (max: %d bytes)",
			len(data), maxSize)
	}

	// Use OAEP padding with SHA-256 for secure encryption
	encryptedData, err := rsa.EncryptOAEP(
		sha256.New(),
		nil, // Adding proper random source
		publicKey,
		data,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return encryptedData, nil
}

// DecryptWithRSA decrypts data using the provided RSA private key
func DecryptWithRSA(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	// Use OAEP padding with SHA-256 for secure decryption
	decryptedData, err := rsa.DecryptOAEP(
		sha256.New(),
		nil,
		privateKey,
		encryptedData,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return decryptedData, nil
}

// LoadPublicKeyFromFile loads an RSA public key from a PEM file
func LoadPublicKeyFromFile(fileName string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}

// LoadPrivateKeyFromFile loads an RSA private key from a PEM file
func LoadPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return priv, nil
}

// Encrypts data of any size using RSA for key exchange and AES for data
func Encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	if publicKey == nil {
		return nil, errors.New("public key is nil")
	}

	// 1. Generate a random AES-256 key
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// 2. Encrypt the AES key with RSA
	encryptedKey, err := EncryptWithRSA(aesKey, publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	// 3. Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// 4. Generate a random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 5. Encrypt the data with AES-GCM
	ciphertext := aesGCM.Seal(nil, nonce, data, nil)

	// 6. Format the output: [RSA key length (4 bytes)][RSA encrypted key][nonce][AES encrypted data]
	output := make([]byte, 4+len(encryptedKey)+len(nonce)+len(ciphertext))

	// Write RSA encrypted key length
	binary.LittleEndian.PutUint32(output[:4], uint32(len(encryptedKey)))

	// Copy RSA encrypted key
	copy(output[4:], encryptedKey)

	// Copy nonce
	copy(output[4+len(encryptedKey):], nonce)

	// Copy AES encrypted data
	copy(output[4+len(encryptedKey)+len(nonce):], ciphertext)

	return output, nil
}

// Decrypts data encrypted using the Encrypt function
func Decrypt(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	// Need at least 4 bytes for RSA key length
	if len(encryptedData) < 4 {
		return nil, errors.New("encrypted data too short")
	}

	// 1. Read the RSA encrypted key length
	keyLength := binary.LittleEndian.Uint32(encryptedData[:4])

	// Validate the key length
	if len(encryptedData) < int(4+keyLength) {
		return nil, errors.New("encrypted data too short for key length")
	}

	// 2. Extract the RSA encrypted key
	encryptedKey := encryptedData[4 : 4+keyLength]

	// 3. Decrypt the AES key with RSA
	aesKey, err := DecryptWithRSA(encryptedKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// 4. Create AES-GCM cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// 5. Extract nonce and ciphertext
	nonceSize := aesGCM.NonceSize()
	if len(encryptedData) < 4+int(keyLength)+nonceSize {
		return nil, errors.New("encrypted data too short for nonce")
	}

	nonce := encryptedData[4+int(keyLength) : 4+int(keyLength)+nonceSize]
	ciphertext := encryptedData[4+int(keyLength)+nonceSize:]

	// 6. Decrypt the data with AES-GCM
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("AES decryption failed: %w", err)
	}

	return plaintext, nil
}
