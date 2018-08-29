package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"reflect"
	"strings"
)

// ListCommonAddressToListString converts a list of common.address to list of string
func ListCommonAddressToListString(addresses []common.Address) []string {
	addressesString := make([]string, len(addresses))
	for i, address := range addresses {
		addressesString[i] = address.Hex()
	}
	return addressesString
}

// ListStringToListCommonAddress converts a list of strings to list of common.address
func ListStringToListCommonAddress(addresses []string) []common.Address {
	addressesCommon := make([]common.Address, len(addresses))
	for i, address := range addresses {
		addressesCommon[i] = common.HexToAddress(address)
	}
	return addressesCommon
}

// ListCommonAddressesToString converts a list of common.address to string
func ListCommonAddressesToString(addresses []common.Address) string {
	addressesString := ListCommonAddressToListString(addresses)
	return strings.Join(addressesString, ",")
}

// StringToCommonAddressesList converts a list of common.address to comma delimited string
func StringToCommonAddressesList(addresses string) []common.Address {
	addressesString := strings.Split(addresses, ",")
	return ListStringToListCommonAddress(addressesString)
}

// DbFieldNameFromModelName gets the field name from db given postgres model struct
func DbFieldNameFromModelName(exampleStruct interface{}, fieldName string) (string, error) {
	sType := reflect.TypeOf(exampleStruct)
	field, ok := sType.FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("%s may not exist in struct", fieldName)
	}
	return field.Tag.Get("db"), nil
}

// StructFieldsForQuery is a generic Insert statement for any table
// TODO(IS): gosec linting errors for bytes.buffer use here. figure out if it's inefficient
func StructFieldsForQuery(exampleStruct interface{}, colon bool) (string, string) {
	var fields bytes.Buffer
	var fieldsWithColon bytes.Buffer
	valStruct := reflect.ValueOf(exampleStruct)
	typeOf := valStruct.Type()
	for i := 0; i < valStruct.NumField(); i++ {
		dbFieldName := typeOf.Field(i).Tag.Get("db")
		fields.WriteString(dbFieldName) // nolint: gosec
		if colon {
			fieldsWithColon.WriteString(":")         // nolint: gosec
			fieldsWithColon.WriteString(dbFieldName) // nolint: gosec
		}
		if i+1 < valStruct.NumField() {
			fields.WriteString(", ") // nolint: gosec
			if colon {
				fieldsWithColon.WriteString(", ") // nolint: gosec
			}
		}
	}
	return fields.String(), fieldsWithColon.String()
}
