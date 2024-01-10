package auth

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"strings"
)

var (
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrInvalidKeyType = errors.New("HS256 key is invalid")
)

// Verify the signature of HS256 tokens. Returns nil if the signature is valid.
func VerifyToken(signingString, signature string, key interface{}) error {
	// Verify the key is the right type
	keyBytes, ok := key.([]byte)
	if !ok {
		return ErrInvalidKeyType
	}

	// Decode signature, for comparison
	sig, err := DecodeSegment(signature)
	if err != nil {
		return err
	}

	// This signing method is symmetric, so we validate the signature
	// by reproducing the signature from the signing string and key, then
	// comparing that against the provided signature.
	hasher := hmac.New(crypto.SHA256.New, keyBytes)
	hasher.Write([]byte(signingString))
	if !hmac.Equal(sig, hasher.Sum(nil)) {
		return ErrTokenInvalid
	}

	// No validation errors.  Signature is good.
	return nil
}

// Implements the Sign method from SigningMethod for this signing method.
// Key must be []byte
func SignToken(signingString string, key interface{}) (string, error) {
	if keyBytes, ok := key.([]byte); ok {
		hasher := hmac.New(crypto.SHA256.New, keyBytes)
		hasher.Write([]byte(signingString))

		return EncodeSegment(hasher.Sum(nil)), nil
	}

	return "", ErrInvalidKeyType
}

// Encode JWT specific base64url encoding with padding stripped
func EncodeSegment(seg []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(seg), "=")
}

// Decode JWT specific base64url encoding with padding stripped
func DecodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}
