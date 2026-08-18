package license

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/ygpkg/yg-go/logs"
)

type KeyConfig struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

// ParsePrivateKey will parse a private key (type string) to rsa.PrivateKey
func ParsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded private key: %w", err)
	}
	return key, nil
}

// ParsePublicKey will parse a public key (type string) to rsa.PublicKey
func ParsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

// GenerateKeys generates public and private key files
func GenerateKeys(ctx context.Context) ([]byte, []byte, error) {
	var (
		privateKeyPem, publicKeyPem []byte
	)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logs.ErrorContextf(ctx, "GenerateKeys error: %v", err)
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKeyPem = pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyPem = pem.EncodeToMemory(&pem.Block{
		Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	return privateKeyPem, publicKeyPem, nil
}

// CalculatePublicKeyFingerprint computes the SHA-256 fingerprint of a public key
// The fingerprint is based on the DER-encoded public key, which is the standard
// for this type of verification.
func CalculatePublicKeyFingerprint(pubKey *rsa.PublicKey) (string, error) {
	// Marshal the public key to its standard DER-encoded format
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Compute the SHA-256 hash of the DER bytes
	hash := sha256.Sum256(pubKeyBytes)

	// Convert the hash to a hex string for comparison
	return hex.EncodeToString(hash[:]), nil
}
