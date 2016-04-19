package crypto

import (
	"crypto/cipher"
	"encoding/binary"
	"errors"

	"github.com/lucas-clemente/quic-go/crypto/chacha20poly1305trunc12"
	"github.com/lucas-clemente/quic-go/protocol"
)

type aeadChacha20Poly1305 struct {
	otherIV   []byte
	myIV      []byte
	encrypter cipher.AEAD
	decrypter cipher.AEAD
}

// NewAEADChacha20Poly1305 creates a AEAD using chacha20poly1305
func NewAEADChacha20Poly1305(otherKey []byte, myKey []byte, otherIV []byte, myIV []byte) (AEAD, error) {
	if len(myKey) != 32 || len(otherKey) != 32 || len(myIV) != 4 || len(otherIV) != 4 {
		return nil, errors.New("chacha20poly1305: expected 32-byte keys and 4-byte IVs")
	}
	encrypter, err := chacha20poly1305trunc12.NewAEAD(myKey)
	if err != nil {
		return nil, err
	}
	decrypter, err := chacha20poly1305trunc12.NewAEAD(otherKey)
	if err != nil {
		return nil, err
	}
	return &aeadChacha20Poly1305{
		otherIV:   otherIV,
		myIV:      myIV,
		encrypter: encrypter,
		decrypter: decrypter,
	}, nil
}

func (aead *aeadChacha20Poly1305) Open(packetNumber protocol.PacketNumber, associatedData []byte, ciphertext []byte) ([]byte, error) {
	plaintext, err := aead.decrypter.Open(make([]byte, len(ciphertext)), makeNonce(aead.otherIV, packetNumber), ciphertext, associatedData)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func (aead *aeadChacha20Poly1305) Seal(packetNumber protocol.PacketNumber, associatedData []byte, plaintext []byte) []byte {
	return aead.encrypter.Seal(make([]byte, len(plaintext)+12), makeNonce(aead.myIV, packetNumber), plaintext, associatedData)
}

func makeNonce(iv []byte, packetNumber protocol.PacketNumber) []byte {
	res := make([]byte, 12)
	copy(res[0:4], iv)
	binary.LittleEndian.PutUint64(res[4:12], uint64(packetNumber))
	return res
}