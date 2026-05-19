package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	SaltLength  uint32
	KeyLength   uint32
	Iterations  uint32
	MemorySize  uint32
	Parallelism uint8
}

var DefaultParams = Argon2Params{
	SaltLength:  16,
	KeyLength:   32,
	Iterations:  3,
	MemorySize:  64 * 1024,
	Parallelism: 4,
}

func AuthorizeUser(storedHash, password string) (bool, error) {
	if password == "" {
		return false, errors.New("password must not be empty")
	}

	params, salt, hash, err := parseHash(storedHash)
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.MemorySize, params.Parallelism, params.KeyLength)
	match := subtle.ConstantTimeCompare(hash, computedHash) == 1

	return match, nil
}

func parseHash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, errors.New("unsupported Argon2 variant")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("unsupported Argon2 version")
	}

	params := &Argon2Params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.MemorySize, &params.Iterations, &params.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid salt encoding: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid hash encoding: %w", err)
	}

	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil

}

func GenerateHash(password string, p Argon2Params) (string, error) {
	if password == "" {
		return "", errors.New("password must not be empty")
	}
	if p.SaltLength == 0 || p.KeyLength == 0 || p.Iterations == 0 || p.MemorySize == 0 || p.Parallelism == 0 {
		return "", errors.New("invalid Argon2 parameters: all fields must be non-zero")
	}

	salt, err := generateRandomSalt(p.SaltLength)
	if err != nil {
		return "", err
	}

	hashRaw := argon2.IDKey([]byte(password), salt, p.Iterations, p.MemorySize, p.Parallelism, p.KeyLength)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.MemorySize,
		p.Iterations,
		p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hashRaw),
	)

	return encodedHash, nil
}

func generateRandomSalt(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
