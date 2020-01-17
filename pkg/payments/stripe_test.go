package payments_test

import (
	"github.com/joincivil/civil-api-server/pkg/payments"
	"strings"
	"testing"
)

var table = []struct {
	groupA     []string
	groupB     []string
	isSubset   bool
	difference []string
}{
	{[]string{}, []string{}, true, []string{}},
	{[]string{"foo"}, []string{}, false, []string{"foo"}},
	{[]string{"foo"}, []string{"foo"}, true, []string{}},
	{[]string{"bar"}, []string{"foo"}, false, []string{"bar"}},
	{[]string{"foo"}, []string{"bar", "baz"}, false, []string{"foo"}},
	{[]string{"foo"}, []string{"foo", "bar", "baz"}, true, []string{}},
	{[]string{"foo", "bar"}, []string{"foo", "bar", "baz"}, true, []string{}},
	{[]string{"foo", "bar"}, []string{"foo", "baz"}, false, []string{"bar"}},
	{[]string{"foo", "bar"}, []string{"foo", "baz", "bat"}, false, []string{"bar"}},
	{[]string{"civil.co", "registry.civil.co"}, []string{"registry.civil.co", "somesite.com", "anothersite.com"}, false, []string{"civil.co"}},
}

func TestIsSubset(t *testing.T) {

	for _, test := range table {
		isSubset := payments.IsSubset(test.groupA, test.groupB)
		if isSubset != test.isSubset {
			t.Errorf("%v, %v | expected: %v | actual: %v", test.groupA, test.groupB, test.isSubset, isSubset)
		}
	}
}

func TestSetDifference(t *testing.T) {

	for _, test := range table {
		difference := payments.SetDifference(test.groupA, test.groupB)
		if strings.Join(difference, "") != strings.Join(test.difference, "") {
			t.Errorf("%v, %v | expected: %v | actual: %v", test.groupA, test.groupB, test.difference, difference)
		}
	}
}
