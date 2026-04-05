package security

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("this_is_a_32_byte_secret_key_123") // exactly 32 bytes
	if len(key) != 32 {
		t.Fatalf("Test key size is %d, expected 32", len(key))
	}

	plaintext := []byte("sensitive context data inside shmClaw")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Fatal("Ciphertext should not equal plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("Decrypted text mismatch. Expected %q, got %q", plaintext, decrypted)
	}
}

func TestInvalidKeySize(t *testing.T) {
	key := []byte("short_key")
	plaintext := []byte("data")

	_, err := Encrypt(plaintext, key)
	if err != ErrInvalidKeySize {
		t.Fatalf("Expected ErrInvalidKeySize for Encrypt, got %v", err)
	}

	_, err = Decrypt([]byte("some_cipher_text"), key)
	if err != ErrInvalidKeySize {
		t.Fatalf("Expected ErrInvalidKeySize for Decrypt, got %v", err)
	}
}

func TestTamperedCiphertext(t *testing.T) {
	key := []byte("this_is_a_32_byte_secret_key_123")
	plaintext := []byte("data")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with the ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = Decrypt(ciphertext, key)
	if err == nil {
		t.Fatal("Decryption should fail when ciphertext is tampered")
	}
}
