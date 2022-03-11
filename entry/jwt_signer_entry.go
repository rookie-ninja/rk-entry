package rkentry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"strings"
)

// validAlgorithm a simple function which will check input string in slice
func validAlgorithm(e string, list []string) bool {
	for i := range list {
		if e == list[i] {
			return true
		}
	}

	return false
}

// RegisterSymmetricJwtSigner create symmetricJwtSigner
func RegisterSymmetricJwtSigner(entryName, algo string, rawKey []byte) *symmetricJwtSigner {
	res := &symmetricJwtSigner{
		entryName: entryName,
	}

	if !validAlgorithm(algo, res.Algorithms()) {
		return nil
	}

	res.Algorithm = algo
	res.key = rawKey

	switch res.Algorithm {
	case jwt.SigningMethodHS256.Name:
		res.SigningMethod = jwt.SigningMethodHS256
	case jwt.SigningMethodHS384.Name:
		res.SigningMethod = jwt.SigningMethodHS384
	case jwt.SigningMethodHS512.Name:
		res.SigningMethod = jwt.SigningMethodHS512
	}

	GlobalAppCtx.AddEntry(res)

	return res
}

// symmetricJwtSigner a signer which will use symmetric key
type symmetricJwtSigner struct {
	entryName     string            `yaml:"-" json:"-"`
	Algorithm     string            `yaml:"-" json:"-"`
	SigningMethod jwt.SigningMethod `yaml:"-" json:"-"`
	key           []byte            `yaml:"-" json:"-"`
}

func (s *symmetricJwtSigner) Bootstrap(ctx context.Context) {}

func (s *symmetricJwtSigner) Interrupt(ctx context.Context) {}

func (s *symmetricJwtSigner) GetName() string {
	return s.entryName
}

func (s *symmetricJwtSigner) GetType() string {
	return SignerJwtEntryType
}

func (s *symmetricJwtSigner) GetDescription() string {
	return "Symmetric jwt signer"
}

func (s *symmetricJwtSigner) String() string {
	m := map[string]string{
		"name":                s.entryName,
		"algorithm":           s.Algorithm,
		"signingMethod":       s.SigningMethod.Alg(),
		"supportedAlgorithms": strings.Join(s.Algorithms(), ","),
	}

	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// SignJwt sign jwt with key
func (s *symmetricJwtSigner) SignJwt(claim jwt.Claims) (string, error) {
	if claim == nil {
		return "", errors.New("Nil jwt claim")
	}

	token := jwt.NewWithClaims(s.SigningMethod, claim)
	return token.SignedString(s.PubKey())
}

// VerifyJwt verify jwt with key
func (s *symmetricJwtSigner) VerifyJwt(raw string) (*jwt.Token, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != s.Algorithm {
			return nil, fmt.Errorf("unexpected jwt signing algorithm=%v", t.Header["alg"])
		}

		return s.PubKey(), nil
	})

	// return error
	if err != nil {
		return nil, err
	}

	// invalid token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// PubKey return raw token
func (s *symmetricJwtSigner) PubKey() []byte {
	return s.key
}

// Algorithms supported algorithms
func (s *symmetricJwtSigner) Algorithms() []string {
	return []string{
		jwt.SigningMethodHS256.Name,
		jwt.SigningMethodHS384.Name,
		jwt.SigningMethodHS512.Name,
	}
}

// RegisterAsymmetricJwtSigner create asymmetricJwtSigner
func RegisterAsymmetricJwtSigner(entryName, algo string, privPEM, pubPEM []byte) *asymmetricJwtSigner {
	res := &asymmetricJwtSigner{
		entryName: entryName,
	}

	if !validAlgorithm(algo, res.Algorithms()) {
		return nil
	}

	res.Algorithm = algo

	switch res.Algorithm {
	case jwt.SigningMethodRS256.Name, jwt.SigningMethodRS384.Name, jwt.SigningMethodRS512.Name:
		parsedPrivKey, err := jwt.ParseRSAPrivateKeyFromPEM(privPEM)
		if err != nil {
			ShutdownWithError(err)
		}

		parsedPubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubPEM)
		if err != nil {
			ShutdownWithError(err)
		}
		res.privKey = parsedPrivKey
		res.pubKey = parsedPubKey

		switch res.Algorithm {
		case jwt.SigningMethodRS256.Name:
			res.SigningMethod = jwt.SigningMethodRS256
		case jwt.SigningMethodRS384.Name:
			res.SigningMethod = jwt.SigningMethodRS384
		case jwt.SigningMethodRS512.Name:
			res.SigningMethod = jwt.SigningMethodRS512
		}

	case jwt.SigningMethodES256.Name, jwt.SigningMethodES384.Name, jwt.SigningMethodES512.Name:
		parsedPrivKey, err := jwt.ParseECPrivateKeyFromPEM(privPEM)
		if err != nil {
			ShutdownWithError(err)
		}

		parsedPubKey, err := jwt.ParseECPublicKeyFromPEM(pubPEM)
		if err != nil {
			ShutdownWithError(err)
		}
		res.privKey = parsedPrivKey
		res.pubKey = parsedPubKey

		switch res.Algorithm {
		case jwt.SigningMethodES256.Name:
			res.SigningMethod = jwt.SigningMethodES256
		case jwt.SigningMethodES384.Name:
			res.SigningMethod = jwt.SigningMethodES384
		case jwt.SigningMethodES512.Name:
			res.SigningMethod = jwt.SigningMethodES512
		}
	}

	GlobalAppCtx.AddEntry(res)

	return res
}

// asymmetricJwtSigner a signer which will use asymmetric key
type asymmetricJwtSigner struct {
	entryName     string            `yaml:"-" json:"-"`
	Algorithm     string            `yaml:"-" json:"-"`
	SigningMethod jwt.SigningMethod `yaml:"-" json:"-"`
	pubKey        interface{}       `yaml:"-" json:"-"`
	privKey       interface{}       `yaml:"-" json:"-"`
}

func (s *asymmetricJwtSigner) Bootstrap(ctx context.Context) {}

func (s *asymmetricJwtSigner) Interrupt(ctx context.Context) {}

func (s *asymmetricJwtSigner) GetName() string {
	return s.entryName
}

func (s *asymmetricJwtSigner) GetType() string {
	return SignerJwtEntryType
}

func (s *asymmetricJwtSigner) GetDescription() string {
	return "Symmetric jwt signer"
}

func (s *asymmetricJwtSigner) String() string {
	m := map[string]string{
		"name":                s.entryName,
		"algorithm":           s.Algorithm,
		"signingMethod":       s.SigningMethod.Alg(),
		"supportedAlgorithms": strings.Join(s.Algorithms(), ","),
	}

	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// SignJwt sign jwt with key
func (s *asymmetricJwtSigner) SignJwt(claim jwt.Claims) (string, error) {
	if claim == nil {
		return "", errors.New("Nil jwt claim")
	}

	token := jwt.NewWithClaims(s.SigningMethod, claim)
	return token.SignedString(s.privKey)
}

// VerifyJwt verify jwt with key
func (s *asymmetricJwtSigner) VerifyJwt(raw string) (*jwt.Token, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != s.Algorithm {
			return nil, fmt.Errorf("unexpected jwt signing algorithm=%v", t.Header["alg"])
		}

		return s.pubKey, nil
	})

	// return error
	if err != nil {
		return nil, err
	}

	// invalid token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// PubKey return public key
func (s *asymmetricJwtSigner) PubKey() []byte {
	return nil
}

// Algorithms supported algorithms
func (s *asymmetricJwtSigner) Algorithms() []string {
	return []string{
		jwt.SigningMethodRS256.Name,
		jwt.SigningMethodRS384.Name,
		jwt.SigningMethodRS512.Name,
		jwt.SigningMethodES256.Name,
		jwt.SigningMethodES384.Name,
		jwt.SigningMethodES512.Name,
	}
}
