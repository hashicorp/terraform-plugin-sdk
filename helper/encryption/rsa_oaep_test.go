package encryption_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/encryption"
)

func TestRSAOAEPEncryptValue(t *testing.T) {
	plainTexts := []string{
		"",
		"A",
		"ABC",
		repeat("A", 382),
		repeat("A", 383),
	}

	for _, plainText := range plainTexts {
		t.Run(plainText, func(t *testing.T) {
			testRSAOAEPEncryptValue(t, plainText)
		})
	}
}

func testRSAOAEPEncryptValue(t *testing.T, plaintext string) {
	const maxPlainTextLength = 382
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	encodedPubKey, err := pemEncode(&rsaPrivateKey.PublicKey)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	encodedPrivKey, err := pemEncode(rsaPrivateKey)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if encodedPrivKey == "" {
		t.Fatalf("blah")
	}

	encrypted, err := encryption.RSAOAEPEncryptValue(encodedPubKey, plaintext, "description")
	if len(plaintext) > maxPlainTextLength {
		if err == nil {
			t.Fatalf("plain text of length %d cannot be encrypted with rsa oaep", len(plaintext))
		}
		return
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	b64, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	decrypted, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, rsaPrivateKey, b64, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if plaintext != string(decrypted) {
		t.Fatalf("decryption did not return original plain text: %#v\n\t%#v", decrypted, plaintext)
	}
}

func pemEncode(key interface{}) (string, error) {

	if _, ok := key.(*rsa.PublicKey); ok {
		return pemEncodeRSAPublicKey(key)
	} else if _, ok := key.(*rsa.PrivateKey); ok {
		return pemEncodeRSAPrivateKey(key)
	}

	return "", fmt.Errorf("not supported key type %T", key)
}

func pemEncodeRSAPrivateKey(key interface{}) (string, error) {
	rsaPrivateKey := key.(*rsa.PrivateKey)

	b := x509.MarshalPKCS1PrivateKey(rsaPrivateKey)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Bytes:   b,
	})), nil
}

func pemEncodeRSAPublicKey(key interface{}) (string, error) {
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Bytes:   b,
	})), nil
}

func repeat(s string, count int) string {
	repeated := ""
	for i := 0; i < count; i++ {
		repeated += s
	}
	return repeated
}
