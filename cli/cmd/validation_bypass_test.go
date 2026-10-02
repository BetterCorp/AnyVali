package cmd

import "testing"

func TestValidateCommandRejectsRoundedLiteral(t *testing.T) {
	schemaPath := writeTempSchema(t, map[string]any{
		"root": map[string]any{"kind": "literal", "value": int64(9007199254740992)},
	})
	if code := RunValidate([]string{schemaPath, "9007199254740993", "-f", "quiet"}); code != 1 {
		t.Fatalf("rounded neighboring integer passed validation: exit code %d", code)
	}
	if code := RunValidate([]string{schemaPath, "9007199254740992", "-f", "quiet"}); code != 0 {
		t.Fatalf("valid integer rejected: exit code %d", code)
	}
}
