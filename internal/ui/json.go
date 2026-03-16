package ui

import (
	"bytes"
	"encoding/json"
)

func PrettyJSON(text string) (string, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(text), "", "  "); err != nil {
		return "", err
	}
	return out.String(), nil
}
