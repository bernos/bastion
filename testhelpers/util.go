package testhelpers

import (
	"encoding/json"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

func AsJSON(o any) string {
	b, _ := json.MarshalIndent(o, "", "  ")
	return string(b)
}

func UID(t *testing.T) string {
	t.Helper()
	id, err := gonanoid.New(8)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
