package jsonstore

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/joincivil/civil-api-server/pkg/auth"
)

// TokenPlusIDHashKey generates a unique hash of Token.Sub with an arbitrary ID
// Used as the ID PK for the table to prevent from overwriting other keys.
func TokenPlusIDHashKey(token *auth.Token, ID string) (string, error) {
	val := struct {
		TokenSub string
		ID       string
	}{
		TokenSub: token.Sub,
		ID:       ID,
	}

	bys, err := json.Marshal(val)
	if err != nil {
		return "", err
	}

	hsh := sha256.New()
	hsh.Write(bys)
	return base64.RawStdEncoding.EncodeToString(
		hsh.Sum(nil)), nil
}
