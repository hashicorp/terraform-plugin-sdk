package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/hashicorp/errwrap"
)

func RSAOAEPEncryptValue(key string, value string, description string) (string, error) {
	rsaPub, err := rsaPublicKey(key)
	if err != nil {
		return "", errwrap.Wrapf(fmt.Sprintf("Error parsing Public Key for %s: {{err}}", description), err)
	}

	encrypted, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, rsaPub, []byte(value), nil)
	if err != nil {
		return "", errwrap.Wrapf(fmt.Sprintf("Error encrypting with rsa oaep using sha512 hash for %s: {{err}}", description), err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func rsaPublicKey(encoded string) (*rsa.PublicKey, error) {
	pubPem, _ := pem.Decode([]byte(encoded))
	if pubPem.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("error parsing pem encoded public key: %s", pubPem.Type)
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubPem.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not a rsa public key: %T", pubKey)
	}
	return rsaPubKey, nil
}
