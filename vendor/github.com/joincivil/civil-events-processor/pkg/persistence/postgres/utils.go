package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"reflect"
	"strconv"
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

// ListCommonAddressesToString converts a list of common.address to a comma delimited string
func ListCommonAddressesToString(addresses []common.Address) string {
	addressesString := ListCommonAddressToListString(addresses)
	return strings.Join(addressesString, ",")
}

// StringToCommonAddressesList converts a comma delimited string to a list of common.address
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

// BigIntToFloat64 converts big.Int to float64 type
func BigIntToFloat64(bigInt *big.Int) float64 {
	f := new(big.Float).SetInt(bigInt)
	val, _ := f.Float64()
	return val
}

// Float64ToBigInt converts float64 to big.Int type
func Float64ToBigInt(float float64) *big.Int {
	bigInt := new(big.Int)
	bigInt.SetString(strconv.FormatFloat(float, 'f', -1, 64), 10)
	return bigInt
}

// ListIntToListString converts a list of big.int to a list of string
func ListIntToListString(listInt []int) []string {
	listString := make([]string, len(listInt))
	for idx, i := range listInt {
		listString[idx] = strconv.Itoa(i)
	}
	return listString
}
