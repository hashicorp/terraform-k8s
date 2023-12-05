package workspacehelper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/snyk/terraform-k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestShouldReturnStringFromNumber(t *testing.T) {
	expected := "12321"
	value := cty.Value(cty.NumberIntVal(12321))
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturnStringFromJSONStringNull(t *testing.T) {
	expected := `"{"hi":null}"`
	value := cty.Value(cty.StringVal(`"{"hi":null}"`))
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

func TestShouldReturnNullFromNullObject(t *testing.T) {
	expected := "null"
	value := cty.NullVal(cty.Map(cty.String))
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturnEmptyFromEmptyArray(t *testing.T) {
	expected := "[]"
	value := cty.StringVal("[]")
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestShouldReturnArrayFromArray(t *testing.T) {
	expected := "[1,2,3]"
	value := cty.StringVal("[1, 2, 3]")
	formatted := convertValueToString(value)
	assert.Equal(t, expected, formatted)
}

func TestEmbeddedStructures(t *testing.T) {
	tests := []struct {
		name    string
		resp    string
		want    []*v1alpha1.OutputStatus
		wantErr bool
	}{
		{
			name: "Embedded map with null value",
			resp: `
    {
      "version": 4,
      "outputs": {
          "map": {
              "value": [
                  {
                      "data": null
                  }
              ],
              "type": [
                  "tuple",
                  [
                      [
                          "object",
                          {
                              "data": [
                                  "map",
                                  "string"
                              ]
                          }
                      ]
                  ]
              ]
          }
      }
  }`,
			want: []*v1alpha1.OutputStatus{{Key: "map", Value: `[{"data":null}]`}},
		},
		{
			name: "Embedded map with an emptly list as value",
			resp: ` {
        "version": 4,
        "outputs": {
            "map": {
                "value": [
                    {
                        "data": {
                            "key": []
                        }
                    }
                ],
                "type": [
                    "tuple",
                    [
                        [
                            "object",
                            {
                                "data": [
                                    "map",
                                    [
                                        "list",
                                        "string"
                                    ]
                                ]
                            }
                        ]
                    ]
                ]
            }
        }
    } `,
			want: []*v1alpha1.OutputStatus{{Key: "map", Value: `[{"data":{"key":[]}}]`}},
		},
		{
			name: "Embedded map with a list of strings as value",
			resp: ` {
        "version": 4,
        "outputs": {
            "map": {
                "value": [
                    {
                        "data": {
                            "key": [1, 2, 3]
                        }
                    }
                ],
                "type": [
                    "tuple",
                    [
                        [
                            "object",
                            {
                                "data": [
                                    "map",
                                    [
                                        "list",
                                        "string"
                                    ]
                                ]
                            }
                        ]
                    ]
                ]
            }
        }
    } `,
			want: []*v1alpha1.OutputStatus{{Key: "map", Value: `[{"data":{"key":["1","2","3"]}}]`}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, tt.resp)
			}))
			config := &tfe.Config{
				Address:    srv.URL,
				Token:      "token1",
				HTTPClient: srv.Client(),
			}
			client, err := tfe.NewClient(config)
			assert.NoError(t, err)

			cloud := &TerraformCloudClient{
				Client: client,
			}

			outputs, err := cloud.GetOutputsFromState(srv.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("TerraformCloudClient.GetOutputsFromState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, outputs)
		})
	}
}
