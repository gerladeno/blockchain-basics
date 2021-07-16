package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	walletFile         = "wallet.dat"
	addressChecksumLen = 4
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func CreateWallet() (*Wallet, error) {
	private, public, err := newKeyPair()
	if err != nil {
		return nil, err
	}
	return &Wallet{*private, public}, nil
}

func newKeyPair() (*ecdsa.PrivateKey, []byte, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return private, pubKey, nil
}

func (w *Wallet) GetAddress() ([]byte, error) {
	pubKeyHash, err := HashPubKey(w.PublicKey)
	if err != nil {
		return nil, err
	}

	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)

	fullPayLoad := append(versionPayload, checksum...)
	address := Base58Encode(fullPayLoad)
	return address, nil
}

func HashPubKey(pubkey []byte) ([]byte, error) {
	publicSHA256 := sha256.Sum256(pubkey)
	RIPEMD160Hasher := ripemd160.New()
	if _, err := RIPEMD160Hasher.Write(publicSHA256[:]); err != nil {
		return nil, err
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160, nil
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}
