package wallet

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"log"

	"golang.org/x/crypto/sha3"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  []byte
}

func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address
}

func NewKeyPair() (ed25519.PrivateKey, []byte) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	return private, public
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := sha3.New224()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicSHA256 := hasher.Sum(nil)

	return publicSHA256
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

// Serialize serializes the wallet
func (w *Wallet) Serialize() []byte {
	data, err := json.Marshal(w)
	if err != nil {
		log.Panic(err)
	}
	return data
}

// DeserializeWallet deserializes a wallet
func DeserializeWallet(data []byte) Wallet {
	var wallet Wallet
	err := json.Unmarshal(data, &wallet)
	if err != nil {
		log.Panic(err)
	}
	return wallet
}

// MarshalJSON custom marshaler for Wallet
func (w Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey []byte
		PublicKey  []byte
	}{
		PrivateKey: w.PrivateKey,
		PublicKey:  w.PublicKey,
	})
}

// UnmarshalJSON custom unmarshaler for Wallet
func (w *Wallet) UnmarshalJSON(data []byte) error {
	aux := struct {
		PrivateKey []byte
		PublicKey  []byte
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	w.PrivateKey = ed25519.PrivateKey(aux.PrivateKey)
	w.PublicKey = ed25519.PublicKey(aux.PublicKey)

	return nil
}
