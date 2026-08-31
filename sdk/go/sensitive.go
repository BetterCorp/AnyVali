package anyvali

import (
	"encoding/json"
)

const encryptedPrefix = "encrypted:"

// SensitiveTransform converts one sensitive value. Returning an error aborts the operation.
type SensitiveTransform func(path []any, value any) (any, error)

// SafeParseEncrypted validates storage data while treating sensitive values as opaque envelopes.
func SafeParseEncrypted(schema Schema, data any) ParseResult {
	projected, err := Import(&Document{Root: encryptedNode(schema.ToNode())})
	if err != nil {
		return ParseResult{Success: false, Issues: []ValidationIssue{{
			Code: IssueUnsupportedSchemaKind, Message: err.Error(),
		}}}
	}
	return projected.SafeParse(data)
}

// Encrypt parses plaintext, transforms sensitive values, and validates the storage result.
func Encrypt(schema Schema, data any, transform SensitiveTransform) (any, error) {
	plain, err := schema.Parse(data)
	if err != nil {
		return nil, err
	}
	encrypted, err := transformSensitive(schema, plain, nil, "encrypt", transform, map[string]any{})
	if err != nil {
		return nil, err
	}
	checked := SafeParseEncrypted(schema, encrypted)
	if !checked.Success {
		return nil, &ValidationError{Issues: checked.Issues}
	}
	return checked.Data, nil
}

// Decrypt validates storage data, transforms sensitive values, and parses plaintext.
func Decrypt(schema Schema, data any, transform SensitiveTransform) (any, error) {
	checked := SafeParseEncrypted(schema, data)
	if !checked.Success {
		return nil, &ValidationError{Issues: checked.Issues}
	}
	plain, err := transformSensitive(schema, checked.Data, nil, "decrypt", transform, map[string]any{})
	if err != nil {
		return nil, err
	}
	return schema.Parse(plain)
}

func encryptedNode(node map[string]any) map[string]any {
	if meta, ok := node["metadata"].(map[string]any); ok && meta["sensitive"] == true {
		kind, _ := node["kind"].(string)
		marker := map[string]any{"kind": "string", "startsWith": encryptedPrefix, "minLength": len(encryptedPrefix) + 1}
		if kind == "optional" || kind == "nullable" {
			return map[string]any{"kind": kind, "schema": marker}
		}
		return marker
	}

	out := make(map[string]any, len(node))
	for k, v := range node {
		out[k] = v
	}
	kind, _ := node["kind"].(string)
	switch kind {
	case "array":
		projectChild(out, node, "items", "item")
	case "tuple":
		projectList(out, node, "items", "elements")
	case "object":
		if props, ok := node["properties"].(map[string]any); ok {
			projected := make(map[string]any, len(props))
			for k, v := range props {
				projected[k] = encryptedNode(v.(map[string]any))
			}
			out["properties"] = projected
		}
	case "record":
		projectChild(out, node, "value", "values", "valueSchema")
	case "union":
		projectList(out, node, "variants", "schemas")
	case "intersection":
		projectList(out, node, "allOf", "schemas")
	case "optional", "nullable":
		projectChild(out, node, "schema", "inner")
	}
	return out
}

func projectChild(out, node map[string]any, keys ...string) {
	for _, key := range keys {
		if child, ok := node[key].(map[string]any); ok {
			out[key] = encryptedNode(child)
			return
		}
	}
}

func projectList(out, node map[string]any, keys ...string) {
	for _, key := range keys {
		if children, ok := node[key].([]any); ok {
			projected := make([]any, len(children))
			for i, child := range children {
				projected[i] = encryptedNode(child.(map[string]any))
			}
			out[key] = projected
			return
		}
	}
}

func transformSensitive(schema Schema, value any, path []any, mode string, transform SensitiveTransform, cache map[string]any) (any, error) {
	if marker, ok := schema.(interface{ isSensitive() bool }); ok && marker.isSensitive() && value != nil {
		keyBytes, _ := json.Marshal(path)
		key := string(keyBytes)
		if cached, ok := cache[key]; ok {
			return cached, nil
		}
		if mode == "encrypt" {
			checked := schema.SafeParse(value)
			if !checked.Success {
				return nil, &ValidationError{Issues: checked.Issues}
			}
			value = checked.Data
		}
		result, err := transform(append([]any(nil), path...), value)
		if err != nil {
			return nil, err
		}
		cache[key] = result
		return result, nil
	}

	switch s := schema.(type) {
	case *ObjectSchema:
		obj, ok := value.(map[string]any)
		if !ok {
			return value, nil
		}
		out := make(map[string]any, len(obj))
		for k, v := range obj {
			out[k] = v
		}
		for k, child := range s.properties {
			if v, exists := obj[k]; exists {
				var err error
				out[k], err = transformSensitive(child, v, append(path, k), mode, transform, cache)
				if err != nil {
					return nil, err
				}
			}
		}
		return out, nil
	case *ArraySchema:
		return transformSlice(s.item, value, path, mode, transform, cache)
	case *TupleSchema:
		arr, ok := value.([]any)
		if !ok {
			return value, nil
		}
		out := append([]any(nil), arr...)
		for i, child := range s.items {
			var err error
			out[i], err = transformSensitive(child, arr[i], append(path, i), mode, transform, cache)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case *RecordSchema:
		obj, ok := value.(map[string]any)
		if !ok {
			return value, nil
		}
		out := make(map[string]any, len(obj))
		for k, v := range obj {
			var err error
			out[k], err = transformSensitive(s.valueSchema, v, append(path, k), mode, transform, cache)
			if err != nil {
				return nil, err
			}
		}
		return out, nil
	case *OptionalSchema:
		return transformSensitive(s.inner, value, path, mode, transform, cache)
	case *NullableSchema:
		if value == nil {
			return nil, nil
		}
		return transformSensitive(s.inner, value, path, mode, transform, cache)
	case *UnionSchema:
		for _, child := range s.schemas {
			match := child.SafeParse(value)
			if mode == "decrypt" {
				match = SafeParseEncrypted(child, value)
			}
			if match.Success {
				return transformSensitive(child, value, path, mode, transform, cache)
			}
		}
	case *IntersectionSchema:
		result := value
		for _, child := range s.schemas {
			var err error
			result, err = transformSensitive(child, result, path, mode, transform, cache)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case *RefSchema:
		if s.resolved != nil {
			return transformSensitive(s.resolved, value, path, mode, transform, cache)
		}
	}
	return value, nil
}

func transformSlice(item Schema, value any, path []any, mode string, transform SensitiveTransform, cache map[string]any) (any, error) {
	arr, ok := value.([]any)
	if !ok {
		return value, nil
	}
	out := make([]any, len(arr))
	for i, v := range arr {
		var err error
		out[i], err = transformSensitive(item, v, append(path, i), mode, transform, cache)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
