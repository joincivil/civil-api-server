package jsonstore

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

const (
	// NoSaltValue is just a nice placeholder for an empty salt value
	NoSaltValue = ""

	// DefaultNamespaceValue is just a nice placeholder for an empty namespace value
	DefaultNamespaceValue = "default"
)

// NamespaceIDSaltHashKey generates a unique hash of an arbitrary ID. A namespace
// value can be optionally added to the hash. Adding a salt helps to add uniqueness
// within a namespace but is optional.
func NamespaceIDSaltHashKey(namespace string, ID string, salt string) (string, error) {
	val := struct {
		NS   string
		ID   string
		Salt string
	}{
		NS:   namespace,
		ID:   ID,
		Salt: salt,
	}

	bys, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	return CreateHashKey(bys)
}

// CreateHashKey generates the key as a sha256 hash
func CreateHashKey(bys []byte) (string, error) {
	hsh := sha256.New()
	hsh.Write(bys) // nolint: errcheck
	return base64.RawStdEncoding.EncodeToString(
		hsh.Sum(nil)), nil
}
