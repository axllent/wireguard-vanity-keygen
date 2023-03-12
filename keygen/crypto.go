package keygen

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/curve25519"
)

// KeySize defines the size of the key
const KeySize = 32

// Key is curve25519 key.
// It is used by WireGuard to represent public and preshared keys.
type Key [KeySize]byte

// PrivateKey is curve25519 key.
// It is used by WireGuard to represent private keys.
type PrivateKey [KeySize]byte

// NewPrivateKey generates a new curve25519 secret key.
// It conforms to the format described on https://cr.yp.to/ecdh.html.
func newPrivateKey() (PrivateKey, error) {
	k, err := newPresharedKey()
	if err != nil {
		return PrivateKey{}, err
	}
	k[0] &= 248
	k[31] = (k[31] & 127) | 64
	return (PrivateKey)(*k), nil
}

// NewPresharedKey generates a new key
func newPresharedKey() (*Key, error) {
	var k [KeySize]byte
	_, err := rand.Read(k[:])
	if err != nil {
		return nil, err
	}
	return (*Key)(&k), nil
}

// Public computes the public key matching this curve25519 secret key.
func (k *PrivateKey) Public() Key {
	var p [KeySize]byte
	curve25519.ScalarBaseMult(&p, (*[KeySize]byte)(k))
	return (Key)(p)
}

// String returns a private key as a string
func (k *PrivateKey) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

// String returns a public key as a string
func (k Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}
