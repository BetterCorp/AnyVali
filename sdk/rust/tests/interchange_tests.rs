use anyvali::*;
use serde_json::json;
use std::collections::HashMap;

#[test]
fn export_string_schema() {
    let s = string().min_length(1).max_length(100);
    let doc = export(&s, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.anyvali_version, "1.0");
    assert_eq!(doc.schema_version, "1");
    assert_eq!(doc.root["kind"], "string");
    assert_eq!(doc.root["minLength"], 1);
    assert_eq!(doc.root["maxLength"], 100);
}

#[test]
fn export_number_schema() {
    let n = number().min(0.0).max(100.0);
    let doc = export(&n, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "number");
}

#[test]
fn export_int_schema() {
    let i = int().min(1.0).max(10.0);
    let doc = export(&i, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "int");
}

#[test]
fn export_object_schema() {
    let o = object()
        .field("name", Box::new(string()))
        .field("age", Box::new(int()))
        .required(vec!["name", "age"]);
    let doc = export(&o, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "object");
    assert!(doc.root["properties"]["name"].is_object());
    assert!(doc.root["properties"]["age"].is_object());
}

#[test]
fn export_array_schema() {
    let a = array(Box::new(string())).min_items(1);
    let doc = export(&a, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "array");
    assert_eq!(doc.root["items"]["kind"], "string");
    assert_eq!(doc.root["minItems"], 1);
}

#[test]
fn export_with_definitions() {
    let s = ref_("#/definitions/User");
    let mut defs: HashMap<String, Box<dyn Schema>> = HashMap::new();
    defs.insert(
        "User".to_string(),
        Box::new(
            object()
                .field("name", Box::new(string()))
                .required(vec!["name"]),
        ),
    );
    let doc = export(&s, ExportMode::Portable, &defs).unwrap();
    assert!(doc.definitions.contains_key("User"));
    assert_eq!(doc.definitions["User"]["kind"], "object");
}

#[test]
fn export_union() {
    let u = union(vec![Box::new(string()), Box::new(int())]);
    let doc = export(&u, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "union");
    assert_eq!(doc.root["variants"].as_array().unwrap().len(), 2);
}

#[test]
fn export_intersection() {
    let i = intersection(vec![Box::new(number().min(0.0)), Box::new(number().max(100.0))]);
    let doc = export(&i, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "intersection");
}

#[test]
fn export_nullable() {
    let n = nullable(Box::new(string()));
    let doc = export(&n, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "nullable");
    assert_eq!(doc.root["schema"]["kind"], "string");
}

#[test]
fn export_optional() {
    let o = optional(Box::new(int()));
    let doc = export(&o, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "optional");
}

#[test]
fn export_literal() {
    let l = literal(json!("hello"));
    let doc = export(&l, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "literal");
    assert_eq!(doc.root["value"], "hello");
}

#[test]
fn export_enum() {
    let e = enum_(vec![json!("a"), json!("b")]);
    let doc = export(&e, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "enum");
    assert_eq!(doc.root["values"], json!(["a", "b"]));
}

#[test]
fn export_tuple() {
    let t = tuple(vec![Box::new(string()), Box::new(int())]);
    let doc = export(&t, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "tuple");
    assert_eq!(doc.root["elements"].as_array().unwrap().len(), 2);
}

#[test]
fn export_record() {
    let r = record(Box::new(string()));
    let doc = export(&r, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "record");
    assert_eq!(doc.root["values"]["kind"], "string");
}

#[test]
fn export_ref() {
    let r = ref_("#/definitions/User");
    let doc = export(&r, ExportMode::Portable, &HashMap::new()).unwrap();
    assert_eq!(doc.root["kind"], "ref");
    assert_eq!(doc.root["ref"], "#/definitions/User");
}

#[test]
fn export_primitives() {
    for (schema, expected_kind): (Box<dyn Schema>, &str) in vec![
        (Box::new(any()) as Box<dyn Schema>, "any"),
        (Box::new(unknown()), "unknown"),
        (Box::new(never()), "never"),
        (Box::new(null()), "null"),
        (Box::new(bool_()), "bool"),
    ] {
        let doc = export(schema.as_ref(), ExportMode::Portable, &HashMap::new()).unwrap();
        assert_eq!(doc.root["kind"], expected_kind);
    }
}

// ===== Import Tests =====

#[test]
fn import_string_schema() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "string", "minLength": 3 },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert_eq!(schema.kind(), "string");
    assert!(schema.parse(&json!("abc")).is_ok());
    assert!(!schema.safe_parse(&json!("ab")).success);
}

#[test]
fn import_number_schema() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "number", "min": 0, "max": 100 },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!(50)).is_ok());
    assert!(!schema.safe_parse(&json!(-1)).success);
}

#[test]
fn import_int_schema() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "int" },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!(42)).is_ok());
    assert!(!schema.safe_parse(&json!(3.14)).success);
}

#[test]
fn import_object_schema() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": {
            "kind": "object",
            "properties": {
                "name": { "kind": "string" },
                "age": { "kind": "int" }
            },
            "required": ["name", "age"],
            "unknownKeys": "reject"
        },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!({"name": "Alice", "age": 30})).is_ok());
}

#[test]
fn import_with_definitions() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": {
            "kind": "object",
            "properties": {
                "user": { "kind": "ref", "ref": "#/definitions/User" }
            },
            "required": ["user"],
            "unknownKeys": "reject"
        },
        "definitions": {
            "User": {
                "kind": "object",
                "properties": {
                    "name": { "kind": "string" }
                },
                "required": ["name"],
                "unknownKeys": "reject"
            }
        },
        "extensions": {}
    }"#;
    let (schema, ctx) = import(json_str).unwrap();
    let result = schema.parse_with_context(
        &json!({"user": {"name": "Alice"}}),
        &ctx,
    );
    assert!(result.is_ok());
}

#[test]
fn import_coerce_config() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "int", "coerce": "string->int" },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    let result = schema.parse(&json!("42"));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(42));
}

#[test]
fn import_coerce_array_config() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "string", "coerce": ["trim", "lower"] },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    let result = schema.parse(&json!("  HELLO  "));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!("hello"));
}

#[test]
fn import_all_int_widths() {
    for kind in &["int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64"] {
        let json_str = format!(
            r#"{{"anyvaliVersion":"1.0","schemaVersion":"1","root":{{"kind":"{}"}},"definitions":{{}},"extensions":{{}}}}"#,
            kind
        );
        let (schema, _ctx) = import(&json_str).unwrap();
        assert_eq!(schema.kind(), *kind);
    }
}

#[test]
fn import_float_widths() {
    for kind in &["float32", "float64"] {
        let json_str = format!(
            r#"{{"anyvaliVersion":"1.0","schemaVersion":"1","root":{{"kind":"{}"}},"definitions":{{}},"extensions":{{}}}}"#,
            kind
        );
        let (schema, _ctx) = import(&json_str).unwrap();
        assert_eq!(schema.kind(), *kind);
    }
}

#[test]
fn import_all_simple_kinds() {
    for kind in &["any", "unknown", "never", "null", "bool"] {
        let json_str = format!(
            r#"{{"anyvaliVersion":"1.0","schemaVersion":"1","root":{{"kind":"{}"}},"definitions":{{}},"extensions":{{}}}}"#,
            kind
        );
        let (schema, _ctx) = import(&json_str).unwrap();
        assert_eq!(schema.kind(), *kind);
    }
}

#[test]
fn import_literal() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "literal", "value": "hello" },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!("hello")).is_ok());
    assert!(!schema.safe_parse(&json!("world")).success);
}

#[test]
fn import_enum() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "enum", "values": ["a", "b", "c"] },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!("a")).is_ok());
    assert!(!schema.safe_parse(&json!("d")).success);
}

#[test]
fn import_array() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "array", "items": { "kind": "int" }, "minItems": 1 },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!([1, 2])).is_ok());
    assert!(!schema.safe_parse(&json!([])).success);
}

#[test]
fn import_tuple() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "tuple", "elements": [{ "kind": "string" }, { "kind": "int" }] },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!(["hello", 42])).is_ok());
}

#[test]
fn import_record() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "record", "values": { "kind": "string" } },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!({"a": "b"})).is_ok());
}

#[test]
fn import_union() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "union", "variants": [{ "kind": "string" }, { "kind": "int" }] },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!("hello")).is_ok());
    assert!(schema.parse(&json!(42)).is_ok());
}

#[test]
fn import_intersection() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "intersection", "allOf": [{ "kind": "number", "min": 0 }, { "kind": "number", "max": 100 }] },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!(50)).is_ok());
}

#[test]
fn import_optional() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "optional", "schema": { "kind": "string" } },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!("hello")).is_ok());
}

#[test]
fn import_nullable() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "nullable", "schema": { "kind": "string" } },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    assert!(schema.parse(&json!(null)).is_ok());
    assert!(schema.parse(&json!("hello")).is_ok());
}

#[test]
fn import_unknown_keys_strip() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": {
            "kind": "object",
            "properties": { "name": { "kind": "string" } },
            "required": ["name"],
            "unknownKeys": "strip"
        },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    let result = schema.parse(&json!({"name": "Alice", "extra": true}));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!({"name": "Alice"}));
}

#[test]
fn import_unknown_keys_allow() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": {
            "kind": "object",
            "properties": { "name": { "kind": "string" } },
            "required": ["name"],
            "unknownKeys": "allow"
        },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    let result = schema.parse(&json!({"name": "Alice", "extra": true}));
    assert!(result.is_ok());
}

#[test]
fn roundtrip_export_import() {
    let original = object()
        .field("name", Box::new(string().min_length(1)))
        .field("age", Box::new(int().min(0.0)))
        .required(vec!["name", "age"]);

    let doc = export(&original, ExportMode::Portable, &HashMap::new()).unwrap();
    let json_str = interchange::exporter::document_to_json(&doc).unwrap();

    let (imported, _ctx) = import(&json_str).unwrap();
    assert!(imported
        .parse(&json!({"name": "Alice", "age": 30}))
        .is_ok());
    assert!(!imported.safe_parse(&json!({"name": "", "age": 30})).success);
}

#[test]
fn import_from_value_works() {
    let doc_value = json!({
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "string" },
        "definitions": {},
        "extensions": {}
    });
    let (schema, _ctx) = import_value(&doc_value).unwrap();
    assert!(schema.parse(&json!("hello")).is_ok());
}

#[test]
fn import_rejects_unsupported_kind() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "fantasy_type" },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
    assert!(result.unwrap_err().contains("Unsupported schema kind"));
}

#[test]
fn import_rejects_missing_kind() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": {},
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
}

#[test]
fn import_default_with_object() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": {
            "kind": "object",
            "properties": {
                "role": { "kind": "string", "default": "user" }
            },
            "required": [],
            "unknownKeys": "reject"
        },
        "definitions": {},
        "extensions": {}
    }"#;
    let (schema, _ctx) = import(json_str).unwrap();
    let result = schema.parse(&json!({}));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!({"role": "user"}));
}
