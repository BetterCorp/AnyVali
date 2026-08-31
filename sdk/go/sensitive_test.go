package anyvali

import (
	"encoding/json"
	"testing"
)

func TestSensitiveDataRoundTripAndValidation(t *testing.T) {
	sensitive := DescribeOpts{Sensitive: true}
	schema := Object(map[string]Schema{
		"name":    String(),
		"secret":  String().Describe("secret", sensitive),
		"profile": Object(map[string]Schema{"token": String()}).Describe("profile", sensitive),
		"aliases": Array(String().Describe("alias", sensitive)),
	})
	plain := map[string]any{
		"name": "Ada", "secret": "abc",
		"profile": map[string]any{"token": "xyz"}, "aliases": []any{"one", "two"},
	}
	enc := func(_ []any, value any) (any, error) {
		b, _ := json.Marshal(value)
		return encryptedPrefix + string(b), nil
	}
	encrypted, err := Encrypt(schema, plain, enc)
	if err != nil {
		t.Fatal(err)
	}
	if !SafeParseEncrypted(schema, encrypted).Success {
		t.Fatal("encrypted value should validate")
	}
	decrypted, err := Decrypt(schema, encrypted, func(_ []any, value any) (any, error) {
		var out any
		err := json.Unmarshal([]byte(value.(string)[len(encryptedPrefix):]), &out)
		return out, err
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := json.Marshal(decrypted); string(got) != `{"aliases":["one","two"],"name":"Ada","profile":{"token":"xyz"},"secret":"abc"}` {
		t.Fatalf("unexpected round trip: %s", got)
	}
	if _, err := Encrypt(schema, plain, func([]any, any) (any, error) { return "broken", nil }); err == nil {
		t.Fatal("broken envelope should fail")
	}
}
