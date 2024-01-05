package key

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// Encode RSA private key to PEM
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey, passphrase string) ([]byte, error) {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	} else {
		return pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyBytes,
		}), nil
	}
}

// Encode EC private key to PEM
func EncodECPrivateKeyToPEM(privateKey *ecdsa.PrivateKey, passphrase string) ([]byte, error) {
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	} else {
		return pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: keyBytes,
		}), nil
	}
	/*
		x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
		pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

		x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
		pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	*/
}

// Encode RSA private key to PEM
func EncodeRSAPrivateKeyToPEM(privateKey *rsa.PrivateKey, passphrase string) ([]byte, error) {
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if passphrase != "" {
		var err error
		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(passphrase), x509.PEMCipherAES256)
		if err != nil {
			return nil, err
		}
	}
	return pem.EncodeToMemory(block), nil
}

// Decode RSA private key from PEM
func DecodeRSAPrivateKeyFromPEM(pemBytes []byte, passphrase string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("Not a private key")
	}
	var privateKeyBytes []byte
	var err error
	if x509.IsEncryptedPEMBlock(block) {
		privateKeyBytes, err = x509.DecryptPEMBlock(block, []byte(passphrase))
		if err != nil {
			return nil, err
		}
	} else {
		privateKeyBytes = block.Bytes
	}
	return x509.ParsePKCS1PrivateKey(privateKeyBytes)
}
