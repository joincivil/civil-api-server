package utils

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/golang/glog"

	"github.com/99designs/gqlgen/graphql"

	cpostgres "github.com/joincivil/go-common/pkg/persistence/postgres"
)

// MarshalJsonbPayloadScalar takes a JsonbPayload and converst it to graphql
func MarshalJsonbPayloadScalar(payload cpostgres.JsonbPayload) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			glog.Error("Error marshalling payload to JSON")
			return
		}
		_, err = w.Write(jsonBytes)
		if err != nil {
			glog.Error("Error writing jsonBytes")
			return
		}
	})
}

// UnmarshalJsonbPayloadScalar takes the graphql data and converts to postgres.JsonbPayload
func UnmarshalJsonbPayloadScalar(v interface{}) (cpostgres.JsonbPayload, error) {
	kv, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("points must be a map of string->interface")
	}

	return kv, nil
}
