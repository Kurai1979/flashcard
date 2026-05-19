package auth

import (
	"strings"
	"testing"
)

func TestGenerateAndAuthorizeUser(t *testing.T) {
	hash, err := GenerateHash("password123", DefaultParams)
	if err != nil {
		t.Fatalf("GenerateHash: %v", err)
	}

	match, err := AuthorizeUser(hash, "password123")
	if err != nil {
		t.Fatalf("AuthorizeUser: %v", err)
	}
	if !match {
		t.Fatal("expected password to match")
	}
}

func TestAuthorizeUser_WrongPassword(t *testing.T) {
	hash, err := GenerateHash("password123", DefaultParams)
	if err != nil {
		t.Fatalf("GenerateHash: %v", err)
	}

	match, err := AuthorizeUser(hash, "wrongpassword")
	if err != nil {
		t.Fatalf("AuthorizeUser: %v", err)
	}
	if match {
		t.Fatal("expected password to not match")
	}
}

func TestAuthorizeUser_InvalidHash(t *testing.T) {
	_, err := AuthorizeUser("notahash", "password123")
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func TestGenerateHash_UniqueHashes(t *testing.T) {
	h1, err := GenerateHash("password123", DefaultParams)
	if err != nil {
		t.Fatalf("GenerateHash: %v", err)
	}
	h2, err := GenerateHash("password123", DefaultParams)
	if err != nil {
		t.Fatalf("GenerateHash: %v", err)
	}
	if h1 == h2 {
		t.Fatal("expected unique hashes for same password")
	}
}

func TestGenerateHash_Format(t *testing.T) {
	hash, err := GenerateHash("password123", DefaultParams)
	if err != nil {
		t.Fatalf("GenerateHash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected PHC format, got: %s", hash)
	}
}

func TestParseHash_UnsupportedVariant(t *testing.T) {
	_, _, _, err := parseHash("$argon2i$v=19$m=65536,t=1,p=4$c29tZXNhbHQ$aGFzaA")
	if err == nil {
		t.Fatal("expected error for unsupported variant")
	}
}

func TestParseHash_MalformedVariant(t *testing.T) {
	_, _, _, err := parseHash("$argon2idXYZ$v=19$m=65536,t=1,p=4$c29tZXNhbHQ$aGFzaA")
	if err == nil {
		t.Fatal("expected error for malformed variant")
	}
}

func TestGenerateHash_EmptyPassword(t *testing.T) {
	_, err := GenerateHash("", DefaultParams)
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestGenerateHash_ZeroParams(t *testing.T) {
	_, err := GenerateHash("password123", Argon2Params{})
	if err == nil {
		t.Fatal("expected error for zero-value params")
	}
}

func TestAuthorizeUser_EmptyPassword(t *testing.T) {
	hash, err := GenerateHash("password123", DefaultParams)
	if err != nil {
		t.Fatalf("GenerateHash: %v", err)
	}
	_, err = AuthorizeUser(hash, "")
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}
