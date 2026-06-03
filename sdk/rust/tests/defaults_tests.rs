use anyvali::*;
use serde_json::json;

#[test]
fn missing_field_gets_default() {
    let schema = object()
        .field("name", Box::new(string()))
        .field_with_default("role", Box::new(string()), json!("user"))
        .required(vec!["name"]);

    let result = schema.safe_parse(&json!({"name": "Alice"}));

    assert!(result.success, "{:?}", result.issues);
    assert_eq!(result.value.unwrap()["role"], json!("user"));
}

#[test]
fn present_field_is_not_overwritten() {
    let schema = object().field_with_default("role", Box::new(string()), json!("user"));

    let result = schema.safe_parse(&json!({"role": "admin"}));

    assert!(result.success, "{:?}", result.issues);
    assert_eq!(result.value.unwrap()["role"], json!("admin"));
}

#[test]
fn invalid_default_produces_default_invalid() {
    let schema = object().field_with_default("count", Box::new(int().min(10.0)), json!(5));

    let result = schema.safe_parse(&json!({}));

    assert!(!result.success);
    assert_eq!(result.issues[0].code, "default_invalid");
    assert_eq!(
        result.issues[0].path,
        vec![PathSegment::Key("count".to_string())]
    );
}

#[test]
fn null_is_not_absent_for_nullable_default() {
    let schema = object().field_with_default(
        "value",
        Box::new(nullable(Box::new(string()))),
        json!("fallback"),
    );

    let result = schema.safe_parse(&json!({"value": null}));

    assert!(result.success, "{:?}", result.issues);
    assert_eq!(result.value.unwrap()["value"], json!(null));
}

#[test]
fn falsy_defaults_are_applied() {
    let schema = object()
        .field_with_default("count", Box::new(int()), json!(0))
        .field_with_default("name", Box::new(string()), json!(""))
        .field_with_default("active", Box::new(bool_()), json!(false));

    let result = schema.safe_parse(&json!({}));

    assert!(result.success, "{:?}", result.issues);
    assert_eq!(
        result.value.unwrap(),
        json!({"count": 0, "name": "", "active": false})
    );
}

#[test]
fn optional_wrapper_field_gets_default() {
    let schema = object().field(
        "host",
        Box::new(optional(Box::new(string())).default(json!("localhost"))),
    );

    let result = schema.safe_parse(&json!({}));

    assert!(result.success, "{:?}", result.issues);
    assert_eq!(result.value.unwrap(), json!({"host": "localhost"}));
}

#[test]
fn optional_wrapper_default_does_not_override_present_field() {
    let schema = object().field(
        "host",
        Box::new(optional(Box::new(string())).default(json!("localhost"))),
    );

    let result = schema.safe_parse(&json!({"host": "example.com"}));

    assert!(result.success, "{:?}", result.issues);
    assert_eq!(result.value.unwrap(), json!({"host": "example.com"}));
}

#[test]
fn optional_wrapper_default_is_validated() {
    let schema = object().field(
        "host",
        Box::new(optional(Box::new(string().min_length(5))).default(json!("hi"))),
    );

    let result = schema.safe_parse(&json!({}));

    assert!(!result.success);
    assert_eq!(result.issues[0].code, "default_invalid");
    assert_eq!(
        result.issues[0].path,
        vec![PathSegment::Key("host".to_string())]
    );
}

#[test]
fn optional_wrapper_default_is_exported() {
    let node = optional(Box::new(string()))
        .default(json!("localhost"))
        .export_node();

    assert_eq!(node["kind"], json!("optional"));
    assert_eq!(node["default"], json!("localhost"));
}

#[test]
fn mutable_optional_wrapper_default_is_not_shared_between_parses() {
    let schema = object().field(
        "meta",
        Box::new(optional(Box::new(any())).default(json!({"items": []}))),
    );

    let mut first = schema.safe_parse(&json!({})).value.unwrap();
    first["meta"]["items"]
        .as_array_mut()
        .unwrap()
        .push(json!("mutated"));

    let second = schema.safe_parse(&json!({})).value.unwrap();
    assert_eq!(second, json!({"meta": {"items": []}}));
}

#[test]
fn nested_object_field_gets_default() {
    let schema = object()
        .field(
            "user",
            Box::new(
                object()
                    .field("name", Box::new(string()))
                    .field_with_default("role", Box::new(string()), json!("guest"))
                    .required(vec!["name"]),
            ),
        )
        .required(vec!["user"]);

    let result = schema.safe_parse(&json!({"user": {"name": "Bob"}}));

    assert!(result.success, "{:?}", result.issues);
    assert_eq!(result.value.unwrap()["user"]["role"], json!("guest"));
}
