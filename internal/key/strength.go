package key

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"fmt"
)

const (
	// MinRSABits is the smallest RSA key that can be generated or certified.
	MinRSABits = 2048
	// MaxRSABits bounds the cost of verifying a signature made with a key
	// supplied by a client.
	MaxRSABits = 16384
	// minECDSABits rejects curves weaker than P-256.
	minECDSABits = 256
)

// CheckPublicKeyStrength returns an error if a public key is too weak to be
// certified, or of a type that can not be assessed.
func CheckPublicKeyStrength(pub any) error {
	switch k := pub.(type) {
	case *rsa.PublicKey:
		bits := k.N.BitLen()
		if bits < MinRSABits {
			return fmt.Errorf("RSA key too small: %d bits, minimum is %d", bits, MinRSABits)
		}
		if bits > MaxRSABits {
			return fmt.Errorf("RSA key too large: %d bits, maximum is %d", bits, MaxRSABits)
		}
	case *ecdsa.PublicKey:
		if bits := k.Curve.Params().BitSize; bits < minECDSABits {
			return fmt.Errorf("ECDSA curve too small: %d bits, minimum is %d", bits, minECDSABits)
		}
	case ed25519.PublicKey:
	default:
		return fmt.Errorf("unsupported public key type %T", pub)
	}
	return nil
}
