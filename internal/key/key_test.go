package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptedRSAKeyRoundTrip(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ConvertRSAKeyToBlock(priv, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if block.Type != encryptedPKCS8Type {
		t.Fatalf("block type = %q", block.Type)
	}
	data := pem.EncodeToMemory(block)
	got, err := LoadPrivateKey(data, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if !priv.Equal(got.(*rsa.PrivateKey)) {
		t.Fatal("decrypted key differs")
	}
	if _, err := LoadPrivateKey(data, "wrong"); err == nil {
		t.Fatal("wrong passphrase accepted")
	}
}

func TestEncryptedECDSAKeyRoundTrip(t *testing.T) {
	priv, err := GenerateECDSAKey()
	if err != nil {
		t.Fatal(err)
	}
	block, err := ConvertECDSAKeyToBlock(priv, "secret")
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadPrivateKey(pem.EncodeToMemory(block), "secret")
	if err != nil {
		t.Fatal(err)
	}
	if !priv.Equal(got.(*ecdsa.PrivateKey)) {
		t.Fatal("decrypted key differs")
	}
}

// Keys written by earlier versions must stay readable.
func TestLegacyEncryptedRSAKeyStillLoads(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	block, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(priv), []byte("secret"), x509.PEMCipherAES256)
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadPrivateKey(pem.EncodeToMemory(block), "secret")
	if err != nil {
		t.Fatal(err)
	}
	if !priv.Equal(got.(*rsa.PrivateKey)) {
		t.Fatal("decrypted key differs")
	}
}

func TestKeyFileIsPrivate(t *testing.T) {
	if os.PathSeparator != '/' {
		t.Skip("POSIX file modes only")
	}
	file := filepath.Join(t.TempDir(), "k.key")
	if err := os.WriteFile(file, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateECDSAKeyFile("", file); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("mode = %o, want 600", mode)
	}
}

func TestGenerateRSAKeyRejectsWeakSize(t *testing.T) {
	if _, err := GenerateRSAKey(1024); err == nil {
		t.Fatal("1024-bit key generated")
	}
}

func TestCheckPublicKeyStrength(t *testing.T) {
	weakEC, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPublicKeyStrength(&weakEC.PublicKey); err == nil {
		t.Fatal("P-224 accepted")
	}
	goodEC, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPublicKeyStrength(&goodEC.PublicKey); err != nil {
		t.Fatal(err)
	}
}
