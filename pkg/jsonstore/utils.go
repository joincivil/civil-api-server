package jsonstore

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

// NamespacePlusIDHashKey generates a unique hash of a namespace with an arbitrary ID.
// Namespace used as to prevent from overwriting other keys.
func NamespacePlusIDHashKey(namespace string, ID string) (string, error) {
	val := struct {
		NS string
		ID string
	}{
		NS: namespace,
		ID: ID,
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
	hsh.Write(bys)
	return base64.RawStdEncoding.EncodeToString(
		hsh.Sum(nil)), nil
}
