package blockchain

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"github.com/ddulesov/gogost/gost3410"
	"github.com/ddulesov/gogost/gost34112012256"
)

type Signer struct {
	prvKey *gost3410.PrivateKey
	PubKey *gost3410.PublicKey
}

func NewSigner() (*Signer, error) {
	s := &Signer{}
	var err error = nil
	s.prvKey, s.PubKey, err = genKeys()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Signer) Sign(msg []byte) ([]byte, error) {
	sig, err := s.prvKey.Sign(rand.Reader, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to sign message")
	}
	return sig, nil
}

func (s *Signer) SignString(msg string) ([]byte, error) {
	return s.Sign([]byte(msg))
}

func (s *Signer) SignToString(msg []byte) (string, error) {
	sig, err := s.Sign(msg)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

func (s *Signer) SignStringToString(msg string) (string, error) {
	m := []byte(msg)
	sig, err := s.Sign(m)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

func (s *Signer) Check(msg, sign []byte) bool {
	res, err := s.PubKey.VerifyDigest(msg, sign)
	if err != nil {
		return false
	}
	return res
}

func (s *Signer) CheckStrings(msg, sign string) bool {
	m := []byte(msg)
	si, err := hex.DecodeString(sign)
	if err != nil {
		return false
	}
	return s.Check(m, si)
}

func (s *Signer) RegenKeys() error {
	prvKey, PubKey, err := genKeys()
	if err != nil {
		return err
	}
	s.prvKey, s.PubKey = prvKey, PubKey
	return nil
}

func GetHashGost3411(inp []byte) ([]byte, error) {
	hasher := gost34112012256.New()
	_, err := hasher.Write(inp)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash given value")
	}
	return hasher.Sum(nil), nil
}

func genKeys() (*gost3410.PrivateKey, *gost3410.PublicKey, error) {
	curve := gost3410.CurveDefault()
	raw := make([]byte, gost3410.Mode2012)
	_, err := io.ReadFull(rand.Reader, raw)
	prv, err := gost3410.NewPrivateKey(curve, gost3410.Mode2012, raw)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate private key")
	}
	pub, err := prv.PublicKey()
	pubRaw := pub.Raw()
	pub, err = gost3410.NewPublicKey(curve, gost3410.Mode2012, pubRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate public key")
	}
	return prv, pub, nil
}

func PubKeyFromRaw(raw []byte) (*gost3410.PublicKey, error) {
	if len(raw) != 2*int(gost3410.Mode2012) {
		return nil, fmt.Errorf("Invalid pub key length")
	}
	r := make([]byte, len(raw))
	copy(r, raw)
	reverseSlice(r)

	y := trimLeadingZeros(r[:int(gost3410.Mode2012)])
	x := trimLeadingZeros(r[int(gost3410.Mode2012):])
	yBig := (&big.Int{}).SetBytes(y)
	xBig := (&big.Int{}).SetBytes(x)

	curve := gost3410.CurveDefault()
	pk := &gost3410.PublicKey{
		C:    curve,
		Mode: gost3410.Mode2012,
		X:    xBig,
		Y:    yBig,
	}
	return pk, nil
}

func PubKeyFromString(raw string) (*gost3410.PublicKey, error) {
	r, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode string")
	}
	return PubKeyFromRaw(r)
}

func CheckSign(msg, sign, pubKey []byte) bool {
	pk, err := PubKeyFromRaw(pubKey)
	if err != nil {
		return false
	}
	res, err := pk.VerifyDigest(msg, sign)
	if err != nil {
		return false
	}
	return res
}

func CheckSigString(msg, sign, pubKey string) bool {
	pk, err := PubKeyFromString(pubKey)
	if err != nil {
		return false
	}
	s, err := hex.DecodeString(sign)
	if err != nil {
		return false
	}
	res, err := pk.VerifyDigest([]byte(msg), s)
	if err != nil {
		return false
	}
	return res
}

func MerkleRoot(b [][]byte) ([]byte, error) {
	l := len(b)
	if l == 1 {
		return GetHashGost3411(b[0])
	}
	if l%2 != 0 {
		b = append(b, b[l-1])
		l++
	} else {
		b = append(b, nil)
	}
	writeIdx := 0
	for i := 0; i < l-1; i = i + 2 {
		h1, err := GetHashGost3411(b[i])
		if err != nil {
			return nil, err
		}
		h2, err := GetHashGost3411(b[i+1])
		if err != nil {
			return nil, err
		}
		b[writeIdx] = bytes.Join([][]byte{h1, h2}, nil)
		writeIdx++
	}
	return MerkleRoot(b[:l/2])
}
