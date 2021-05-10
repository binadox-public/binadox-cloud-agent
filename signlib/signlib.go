package signlib

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

var defaultCurve elliptic.Curve = elliptic.P256()

//GeneratePrivateKey : ecdsa.PrivateKey
func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	var privateKey *ecdsa.PrivateKey
	var privateKeyGenerationError error
	privateKey, privateKeyGenerationError = ecdsa.GenerateKey(defaultCurve, rand.Reader)
	if privateKeyGenerationError != nil {
		return privateKey, privateKeyGenerationError
	}
	return privateKey, nil
}

//GeneratePublicKey :
func GeneratePublicKey(privateKey *ecdsa.PrivateKey) *ecdsa.PublicKey {
	var pri ecdsa.PrivateKey
	pri.D, _ = new(big.Int).SetString(fmt.Sprintf("%x", privateKey.D), 16)
	pri.PublicKey.Curve = defaultCurve
	pri.PublicKey.X, pri.PublicKey.Y = pri.PublicKey.Curve.ScalarBaseMult(pri.D.Bytes())

	publicKey := &ecdsa.PublicKey{}
	publicKey.Curve = defaultCurve
	publicKey.X = pri.PublicKey.X
	publicKey.Y = pri.PublicKey.Y
	return publicKey
}

//Signature :
type Signature struct {
	R *big.Int
	S *big.Int
}

//SignMessage : Generates a valid digital signature for golang's ecdsa library
func SignMessage(message string, privateKey *ecdsa.PrivateKey) (Signature, error) {
	var result Signature
	msgHash := fmt.Sprintf(
		"%x",
		sha256.Sum256([]byte(message)),
	)

	signatureR, signatureS, signatureGenerationError := ecdsa.Sign(rand.Reader, privateKey, []byte(msgHash))
	if signatureGenerationError != nil {
		return result, signatureGenerationError
	}
	result.R = signatureR
	result.S = signatureS
	return result, nil
}

//VerifyMessage : Verifies signatures generated using golang's ecdsa function
func VerifyMessage(message string, publicKey *ecdsa.PublicKey, signature Signature) (bool, error) {
	msgHash := fmt.Sprintf(
		"%x",
		sha256.Sum256([]byte(message)),
	)
	return ecdsa.Verify(publicKey, []byte(msgHash), signature.R, signature.S), nil
}

func EncodePrivateKey(privateKey *ecdsa.PrivateKey) (string, error) {
	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	return string(pemEncoded), nil
}

func EncodePublicKey(publicKey *ecdsa.PublicKey) (string, error) {
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	return string(pemEncodedPub), nil
}

func DecodePrivateKey(pemEncoded string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func DecodePublicKey(pemEncoded string) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode([]byte(pemEncoded))
	x509EncodedPub := blockPub.Bytes
	pub, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		return pub, nil
	}
	return nil, errors.New("Unsupported public key type.")
}
