package utils

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/golang/glog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/joincivil/civil-events-crawler/pkg/persistence/postgres"
)

// MarshalJsonbPayloadScalar takes a JsonbPayload and converst it to graphql
func MarshalJsonbPayloadScalar(payload postgres.JsonbPayload) graphql.Marshaler {
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
func UnmarshalJsonbPayloadScalar(v interface{}) (postgres.JsonbPayload, error) {
	kv, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("points must be a map of string->interface")
	}

	return kv, nil
}
