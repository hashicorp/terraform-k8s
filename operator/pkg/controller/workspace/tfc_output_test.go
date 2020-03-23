package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestShouldReturnStringFromNumber(t *testing.T) {
	expected := "12321"
	value := cty.Value(cty.NumberIntVal(12321))
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturnStringFromBool(t *testing.T) {
	expected := "true"
	value := cty.Value(cty.BoolVal(true))
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturnStringFromList(t *testing.T) {
	expected := `["hello","world"]`
	value := cty.ListVal([]cty.Value{cty.StringVal("hello"), cty.StringVal("world")})
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturn1StringFromList(t *testing.T) {
	expected := `["hello"]`
	value := cty.ListVal([]cty.Value{cty.StringVal("hello")})
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturnStringFromMap(t *testing.T) {
	expected := `{"goodbye":"night","hello":"world"}`
	value := cty.MapVal(map[string]cty.Value{
		"goodbye": cty.StringVal("night"),
		"hello":   cty.StringVal("world"),
	})
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturnStringFromObject(t *testing.T) {
	expected := `{"goodbye":true,"hello":{"user":"me"}}`
	value := cty.ObjectVal(map[string]cty.Value{
		"goodbye": cty.BoolVal(true),
		"hello": cty.MapVal(map[string]cty.Value{
			"user": cty.StringVal("me"),
		}),
	})
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}
