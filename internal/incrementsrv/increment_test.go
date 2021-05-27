package incrementsrv

import (
	"strings"
	"testing"

	"github.com/ezeql/appcues-increment-simple/testdata"
	"github.com/stretchr/testify/assert"
)

type test struct {
	jsonInput string
	output    error
}

func Test_incrementRequestValid(t *testing.T) {

	tests := []test{
		{jsonInput: testdata.ValidIncJSONReq, output: nil}, // no error
		{jsonInput: testdata.InvalidValueJSONReq, output: ErrValue},
		{jsonInput: testdata.InvalidKeyJSONReq, output: ErrKey},
		{jsonInput: testdata.MissingValueJSONReq, output: ErrValue},
		{jsonInput: testdata.MissingKeyJSONReq, output: ErrKey},
	}
	for _, test := range tests {
		_, in := InputFromJSONReader(strings.NewReader(test.jsonInput))
		assert.Equal(t, test.output, in, "they should be equal")
	}

}
