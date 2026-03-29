use anyvali::*;
use serde_json::json;

#[test]
fn string_to_int_coercion() {
    let s = int().coerce(vec!["string->int".to_string()]);
    let result = s.parse(&json!("42"));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(42));
}

#[test]
fn string_to_int_trims_whitespace() {
    let s = int().coerce(vec!["string->int".to_string()]);
    let result = s.parse(&json!("  42  "));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(42));
}

#[test]
fn string_to_number_coercion() {
    let s = number().coerce(vec!["string->number".to_string()]);
    let result = s.parse(&json!("3.14"));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(3.14));
}

#[test]
fn string_to_bool_true() {
    let s = bool_().coerce(vec!["string->bool".to_string()]);
    assert_eq!(s.parse(&json!("true")).unwrap(), json!(true));
}

#[test]
fn string_to_bool_false() {
    let s = bool_().coerce(vec!["string->bool".to_string()]);
    assert_eq!(s.parse(&json!("false")).unwrap(), json!(false));
}

#[test]
fn string_to_bool_1() {
    let s = bool_().coerce(vec!["string->bool".to_string()]);
    assert_eq!(s.parse(&json!("1")).unwrap(), json!(true));
}

#[test]
fn string_to_bool_0() {
    let s = bool_().coerce(vec!["string->bool".to_string()]);
    assert_eq!(s.parse(&json!("0")).unwrap(), json!(false));
}

#[test]
fn string_to_bool_case_insensitive() {
    let s = bool_().coerce(vec!["string->bool".to_string()]);
    assert_eq!(s.parse(&json!("TRUE")).unwrap(), json!(true));
}

#[test]
fn trim_coercion() {
    let s = string().coerce(vec!["trim".to_string()]);
    assert_eq!(s.parse(&json!("  hello  ")).unwrap(), json!("hello"));
}

#[test]
fn lower_coercion() {
    let s = string().coerce(vec!["lower".to_string()]);
    assert_eq!(
        s.parse(&json!("HELLO World")).unwrap(),
        json!("hello world")
    );
}

#[test]
fn upper_coercion() {
    let s = string().coerce(vec!["upper".to_string()]);
    assert_eq!(
        s.parse(&json!("hello world")).unwrap(),
        json!("HELLO WORLD")
    );
}

#[test]
fn coercion_failure() {
    let s = int().coerce(vec!["string->int".to_string()]);
    let result = s.safe_parse(&json!("not-a-number"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "coercion_failed");
    assert_eq!(result.issues[0].expected, "int");
    assert_eq!(result.issues[0].received, "not-a-number");
}

#[test]
fn coercion_before_validation() {
    let s = int()
        .min(10.0)
        .coerce(vec!["string->int".to_string()]);
    let result = s.safe_parse(&json!("5"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_small");
    assert_eq!(result.issues[0].expected, "10");
    assert_eq!(result.issues[0].received, "5");
}

#[test]
fn coercion_then_validation_success() {
    let s = int()
        .min(1.0)
        .max(100.0)
        .coerce(vec!["string->int".to_string()]);
    let result = s.parse(&json!("50"));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(50));
}

#[test]
fn chained_coercions() {
    let s = string().coerce(vec!["trim".to_string(), "lower".to_string()]);
    assert_eq!(s.parse(&json!("  HELLO  ")).unwrap(), json!("hello"));
}

#[test]
fn coercion_with_already_correct_type() {
    // If input is already an int, coercion is a no-op
    let s = int().coerce(vec!["string->int".to_string()]);
    let result = s.parse(&json!(42));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(42));
}

#[test]
fn bool_coercion_failure() {
    let s = bool_().coerce(vec!["string->bool".to_string()]);
    let result = s.safe_parse(&json!("maybe"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "coercion_failed");
}

// Test the standalone coerce module
#[test]
fn coerce_module_string_to_int() {
    use anyvali::parse::coerce::apply_coercion;
    let result = apply_coercion(&json!("42"), "string->int", "int");
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(42));
}

#[test]
fn coerce_module_string_to_number() {
    use anyvali::parse::coerce::apply_coercion;
    let result = apply_coercion(&json!("3.14"), "string->number", "number");
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!(3.14));
}

#[test]
fn coerce_module_trim() {
    use anyvali::parse::coerce::apply_coercion;
    let result = apply_coercion(&json!("  hello  "), "trim", "string");
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!("hello"));
}

#[test]
fn coerce_module_lower() {
    use anyvali::parse::coerce::apply_coercion;
    let result = apply_coercion(&json!("HELLO"), "lower", "string");
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!("hello"));
}

#[test]
fn coerce_module_upper() {
    use anyvali::parse::coerce::apply_coercion;
    let result = apply_coercion(&json!("hello"), "upper", "string");
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!("HELLO"));
}

#[test]
fn coerce_module_parse_config_string() {
    use anyvali::parse::coerce::parse_coerce_config;
    let config = parse_coerce_config(&json!("trim"));
    assert_eq!(config, vec!["trim"]);
}

#[test]
fn coerce_module_parse_config_array() {
    use anyvali::parse::coerce::parse_coerce_config;
    let config = parse_coerce_config(&json!(["trim", "lower"]));
    assert_eq!(config, vec!["trim", "lower"]);
}
