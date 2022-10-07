package rk

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-entry/v3/util"
)

const SignerJwtKind = "signerJwt"

// NewSymmetricSignerJwt create symmetricSignerJwt
func NewSymmetricSignerJwt(name, algo string, rawKey []byte) *symmetricSignerJwt {
	res := &symmetricSignerJwt{
		name: name,
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

	return res
}

// symmetricSignerJwt a signer which will use symmetric key
type symmetricSignerJwt struct {
	name          string            `yaml:"-" json:"-"`
	key           []byte            `yaml:"-" json:"-"`
	Algorithm     string            `yaml:"-" json:"-"`
	SigningMethod jwt.SigningMethod `yaml:"-" json:"-"`
}

func (s *symmetricSignerJwt) Category() string {
	return CategoryInline
}

func (s *symmetricSignerJwt) Kind() string {
	return SignerJwtKind
}

func (s *symmetricSignerJwt) Name() string {
	return s.name
}

func (s *symmetricSignerJwt) Config() EntryConfig {
	return nil
}

func (s *symmetricSignerJwt) Monitor() *Monitor {
	return nil
}

func (s *symmetricSignerJwt) FS() *embed.FS {
	return nil
}

func (s *symmetricSignerJwt) Apis() []*BuiltinApi {
	return []*BuiltinApi{}
}

func (s *symmetricSignerJwt) Bootstrap(ctx context.Context) {}

func (s *symmetricSignerJwt) Interrupt(ctx context.Context) {}

// SignJwt sign jwt with key
func (s *symmetricSignerJwt) SignJwt(claim jwt.Claims) (string, error) {
	if claim == nil {
		return "", errors.New("nil jwt claim")
	}

	token := jwt.NewWithClaims(s.SigningMethod, claim)
	return token.SignedString(s.PubKey())
}

// VerifyJwt verify jwt with key
func (s *symmetricSignerJwt) VerifyJwt(raw string) (*jwt.Token, error) {
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
func (s *symmetricSignerJwt) PubKey() []byte {
	return s.key
}

// Algorithms supported algorithms
func (s *symmetricSignerJwt) Algorithms() []string {
	return []string{
		jwt.SigningMethodHS256.Name,
		jwt.SigningMethodHS384.Name,
		jwt.SigningMethodHS512.Name,
	}
}

// NewAsymmetricSignerJwt create asymmetricSignerJwt
func NewAsymmetricSignerJwt(name, algo string, privatePEM, pubPEM []byte) *asymmetricSignerJwt {
	res := &asymmetricSignerJwt{
		name: name,
	}

	if !validAlgorithm(algo, res.Algorithms()) {
		return nil
	}

	res.Algorithm = algo

	switch res.Algorithm {
	case jwt.SigningMethodRS256.Name, jwt.SigningMethodRS384.Name, jwt.SigningMethodRS512.Name:
		parsedPrivKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
		if err != nil {
			rku.ShutdownWithError(err)
		}

		parsedPubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubPEM)
		if err != nil {
			rku.ShutdownWithError(err)
		}
		res.privateKey = parsedPrivKey
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
		parsedPrivKey, err := jwt.ParseECPrivateKeyFromPEM(privatePEM)
		if err != nil {
			rku.ShutdownWithError(err)
		}

		parsedPubKey, err := jwt.ParseECPublicKeyFromPEM(pubPEM)
		if err != nil {
			rku.ShutdownWithError(err)
		}
		res.privateKey = parsedPrivKey
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

	return res
}

// asymmetricSignerJwt a signer which will use asymmetric key
type asymmetricSignerJwt struct {
	name          string            `yaml:"-" json:"-"`
	Algorithm     string            `yaml:"-" json:"-"`
	SigningMethod jwt.SigningMethod `yaml:"-" json:"-"`
	pubKey        interface{}       `yaml:"-" json:"-"`
	privateKey    interface{}       `yaml:"-" json:"-"`
}

func (s *asymmetricSignerJwt) Category() string {
	return CategoryInline
}

func (s *asymmetricSignerJwt) Kind() string {
	return SignerJwtKind
}

func (s *asymmetricSignerJwt) Name() string {
	return s.name
}

func (s *asymmetricSignerJwt) Config() EntryConfig {
	return nil
}

func (s *asymmetricSignerJwt) Monitor() *Monitor {
	return nil
}

func (s *asymmetricSignerJwt) FS() *embed.FS {
	return nil
}

func (s *asymmetricSignerJwt) Apis() []*BuiltinApi {
	return []*BuiltinApi{}
}

func (s *asymmetricSignerJwt) Bootstrap(ctx context.Context) {}

func (s *asymmetricSignerJwt) Interrupt(ctx context.Context) {}

// SignJwt sign jwt with key
func (s *asymmetricSignerJwt) SignJwt(claim jwt.Claims) (string, error) {
	if claim == nil {
		return "", errors.New("Nil jwt claim")
	}

	token := jwt.NewWithClaims(s.SigningMethod, claim)
	return token.SignedString(s.privateKey)
}

// VerifyJwt verify jwt with key
func (s *asymmetricSignerJwt) VerifyJwt(raw string) (*jwt.Token, error) {
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
func (s *asymmetricSignerJwt) PubKey() []byte {
	return nil
}

// Algorithms supported algorithms
func (s *asymmetricSignerJwt) Algorithms() []string {
	return []string{
		jwt.SigningMethodRS256.Name,
		jwt.SigningMethodRS384.Name,
		jwt.SigningMethodRS512.Name,
		jwt.SigningMethodES256.Name,
		jwt.SigningMethodES384.Name,
		jwt.SigningMethodES512.Name,
	}
}

// validAlgorithm a simple function which will check input string in slice
func validAlgorithm(e string, list []string) bool {
	for i := range list {
		if e == list[i] {
			return true
		}
	}

	return false
}
