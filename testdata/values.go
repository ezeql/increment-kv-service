package testdata

import (
	"fmt"
	"strings"
)

const (
	ValidKey     string = "12345678-1234-5678-1234-567812345678"
	ValidValue   uint64 = 1
	InvalidKey   string = "NOTVALID-1234-567812345678"
	InvalidValue uint64 = 0

	mask string = "00000000-0000-0000-0000-000000000000"
)

var (
	ValidIncJSONReq     = fmt.Sprintf(`{"key": "%s", "value": %v}`, ValidKey, ValidValue)
	InvalidKeyJSONReq   = fmt.Sprintf(`{"key": "%s", "value": %v}`, InvalidKey, ValidValue)
	InvalidValueJSONReq = fmt.Sprintf(`{"key": "%s", "value": %v}`, ValidKey, InvalidValue)
	MissingValueJSONReq = fmt.Sprintf(`{"key": "%s"}`, ValidKey)
	MissingKeyJSONReq   = fmt.Sprintf(`{"value": %v}`, ValidValue)
	InvalidJSONReq      = "{ brokenRequest :) }"

	ListValidKeys = []string{
		fillMask(1), //generate valids UUIDs filled with 1's, 2's, etc
		fillMask(2),
		fillMask(3),
		fillMask(4),
		fillMask(5),
	}
)

func BuildIncrementJSON(k string, v uint64) string {
	return fmt.Sprintf(`{"key": "%s", "value": %v}`, k, v)
}

func fillMask(n int) string {
	return strings.ReplaceAll(mask, "0", fmt.Sprint(n))
}
