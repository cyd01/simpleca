package key

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
)

// Encrypted PKCS#8 (RFC 5958) using PBES2 (RFC 8018) with PBKDF2-HMAC-SHA256
// and AES-256-CBC. This is the format OpenSSL writes by default, as
// "ENCRYPTED PRIVATE KEY".

const (
	encryptedPKCS8Type = "ENCRYPTED PRIVATE KEY"

	pbkdf2Iterations = 600000
	// Upper bound accepted when decrypting: the iteration count comes from the
	// key being loaded, which the web API takes from the request body.
	maxPBKDF2Iterations = 10000000
)

var (
	oidPBES2      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 13}
	oidPBKDF2     = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 12}
	oidHMACSHA256 = asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 9}
	oidAES256CBC  = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 42}

	errDecrypt = errors.New("Can not decrypt private key with passphrase")
)

type encryptedPrivateKeyInfo struct {
	Algorithm pkix.AlgorithmIdentifier
	Data      []byte
}

type pbes2Params struct {
	KDF    pkix.AlgorithmIdentifier
	Scheme pkix.AlgorithmIdentifier
}

type pbkdf2Params struct {
	Salt       []byte
	Iterations int
	KeyLength  int `asn1:"optional"`
	PRF        pkix.AlgorithmIdentifier
}

// encryptPKCS8 encrypts a DER encoded PKCS#8 private key into a PEM block.
func encryptPKCS8(der []byte, passphrase string) (*pem.Block, error) {
	salt := make([]byte, 16)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	dk, err := pbkdf2.Key(sha256.New, passphrase, salt, pbkdf2Iterations, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}
	padding := aes.BlockSize - len(der)%aes.BlockSize
	data := make([]byte, len(der)+padding)
	copy(data, der)
	for i := len(der); i < len(data); i++ {
		data[i] = byte(padding)
	}
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(data, data)

	kdfBytes, err := asn1.Marshal(pbkdf2Params{
		Salt:       salt,
		Iterations: pbkdf2Iterations,
		PRF:        pkix.AlgorithmIdentifier{Algorithm: oidHMACSHA256, Parameters: asn1.NullRawValue},
	})
	if err != nil {
		return nil, err
	}
	ivBytes, err := asn1.Marshal(iv)
	if err != nil {
		return nil, err
	}
	paramBytes, err := asn1.Marshal(pbes2Params{
		KDF:    pkix.AlgorithmIdentifier{Algorithm: oidPBKDF2, Parameters: asn1.RawValue{FullBytes: kdfBytes}},
		Scheme: pkix.AlgorithmIdentifier{Algorithm: oidAES256CBC, Parameters: asn1.RawValue{FullBytes: ivBytes}},
	})
	if err != nil {
		return nil, err
	}
	out, err := asn1.Marshal(encryptedPrivateKeyInfo{
		Algorithm: pkix.AlgorithmIdentifier{Algorithm: oidPBES2, Parameters: asn1.RawValue{FullBytes: paramBytes}},
		Data:      data,
	})
	if err != nil {
		return nil, err
	}
	return &pem.Block{Type: encryptedPKCS8Type, Bytes: out}, nil
}

// decryptPKCS8 returns the DER encoded PKCS#8 private key held in an
// "ENCRYPTED PRIVATE KEY" PEM block.
func decryptPKCS8(block *pem.Block, passphrase string) ([]byte, error) {
	var info encryptedPrivateKeyInfo
	if rest, err := asn1.Unmarshal(block.Bytes, &info); err != nil || len(rest) != 0 {
		return nil, errors.New("Can not decode encrypted private key")
	}
	if !info.Algorithm.Algorithm.Equal(oidPBES2) {
		return nil, errors.New("Unsupported private key encryption (only PBES2 is supported)")
	}
	var params pbes2Params
	if _, err := asn1.Unmarshal(info.Algorithm.Parameters.FullBytes, &params); err != nil {
		return nil, errors.New("Can not decode encrypted private key")
	}
	if !params.KDF.Algorithm.Equal(oidPBKDF2) || !params.Scheme.Algorithm.Equal(oidAES256CBC) {
		return nil, errors.New("Unsupported private key encryption (only PBKDF2 with AES-256-CBC is supported)")
	}
	var kdf pbkdf2Params
	if _, err := asn1.Unmarshal(params.KDF.Parameters.FullBytes, &kdf); err != nil {
		return nil, errors.New("Unsupported private key encryption (only HMAC-SHA256 is supported)")
	}
	if !kdf.PRF.Algorithm.Equal(oidHMACSHA256) {
		return nil, errors.New("Unsupported private key encryption (only HMAC-SHA256 is supported)")
	}
	if kdf.Iterations < 1 || kdf.Iterations > maxPBKDF2Iterations {
		return nil, errors.New("Encrypted private key has an unacceptable iteration count")
	}
	var iv []byte
	if _, err := asn1.Unmarshal(params.Scheme.Parameters.FullBytes, &iv); err != nil || len(iv) != aes.BlockSize {
		return nil, errors.New("Can not decode encrypted private key")
	}
	if len(info.Data) == 0 || len(info.Data)%aes.BlockSize != 0 {
		return nil, errors.New("Can not decode encrypted private key")
	}

	dk, err := pbkdf2.Key(sha256.New, passphrase, kdf.Salt, kdf.Iterations, 32)
	if err != nil {
		return nil, err
	}
	aesBlock, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}
	data := make([]byte, len(info.Data))
	cipher.NewCBCDecrypter(aesBlock, iv).CryptBlocks(data, info.Data)

	padding := int(data[len(data)-1])
	if padding < 1 || padding > aes.BlockSize {
		return nil, errDecrypt
	}
	for _, b := range data[len(data)-padding:] {
		if int(b) != padding {
			return nil, errDecrypt
		}
	}
	return data[:len(data)-padding], nil
}
