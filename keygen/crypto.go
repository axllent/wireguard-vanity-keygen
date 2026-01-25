// Package keygen is a library to generate keys
package keygen

import (
	"crypto/rand"
	"encoding/base64"

	curve25519voi "github.com/oasisprotocol/curve25519-voi/curve"
	scalar "github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

// KeySize defines the size of the key
const KeySize = 32

// Key is curve25519 key.
// It is used by WireGuard to represent public and pre-shared keys.
type Key [KeySize]byte

// PrivateKey is curve25519 key.
// It is used by WireGuard to represent private keys.
type PrivateKey [KeySize]byte

// NewPrivateKey generates a new curve25519 secret key (clamped).
func newPrivateKey() (PrivateKey, error) {
	var priv [KeySize]byte
	_, err := rand.Read(priv[:])
	if err != nil {
		return PrivateKey{}, err
	}
	// Clamp as per RFC 7748
	priv[0] &= 248
	priv[31] = (priv[31] & 127) | 64
	return PrivateKey(priv), nil
}

// Public computes the public key matching this curve25519 secret key using curve25519-voi.
func (k *PrivateKey) Public() Key {
	var pub curve25519voi.MontgomeryPoint
	base := *curve25519voi.X25519_BASEPOINT
	s, err := scalar.NewFromBytesModOrder(k[:])
	if err != nil || s == nil {
		panic("invalid private key for scalar.NewFromBytesModOrder: " + err.Error())
	}
	pub.Mul(&base, s)
	return Key(pub)
}

// String returns a private key as a string
func (k *PrivateKey) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

// String returns a public key as a string
func (k Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}
