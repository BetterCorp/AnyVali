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

// ---------------------------------------------------------------------------
// CWE-20 / spec 5.1: non-portable coercion bypass
//
// str::parse is more permissive than the ECMA-262 reference (JS): parse::<i64>
// accepts a leading "+", and parse::<f64> accepts "inf"/"infinity"/"nan". These
// let a string the JS reference rejects coerce into a number. Coercion must
// accept ASCII decimals only so behaviour is identical across SDKs.
// ---------------------------------------------------------------------------

#[test]
fn coerce_int_rejects_non_decimal() {
    let s = int().coerce(vec!["string->int".to_string()]);
    for bad in ["+5", "1_000", "1.0", "1e3", "0x10"] {
        assert!(!s.safe_parse(&json!(bad)).success, "string->int must reject {bad:?}");
    }
    for good in ["42", "-7", "  42  "] {
        assert!(s.safe_parse(&json!(good)).success, "string->int must accept {good:?}");
    }
}

#[test]
fn coerce_number_rejects_non_decimal() {
    let s = number().coerce(vec!["string->number".to_string()]);
    for bad in ["inf", "infinity", "nan", "NaN", "1_000.5"] {
        assert!(!s.safe_parse(&json!(bad)).success, "string->number must reject {bad:?}");
    }
    for good in ["3.14", "+5", ".5", "1e3", "  3.5  "] {
        assert!(s.safe_parse(&json!(good)).success, "string->number must accept {good:?}");
    }
}

// ---------------------------------------------------------------------------
// Cross-SDK regression: enabling coercion with NO explicit source.
//
// Mirrors a JS SDK bug: turning coercion ON without naming a source silently
// no-ops, so string input fails with invalid_type instead of coercing. The
// only portable coercion source is "string", so "coerce, default source"
// MUST behave as string->{int,number,bool}. The idiomatic no-arg ergonomic is
// `.coerce_default()` (equivalent to `.coerce(vec![])`), which infers the
// target from the schema kind.
// ---------------------------------------------------------------------------

#[test]
fn default_coerce_number_from_string() {
    let s = number().coerce_default();
    let result = s.parse(&json!("3.14"));
    assert!(result.is_ok(), "default coerce should coerce \"3.14\": {result:?}");
    assert_eq!(result.unwrap(), json!(3.14));
}

#[test]
fn default_coerce_int_from_string() {
    let s = int().coerce_default();
    let result = s.parse(&json!("42"));
    assert!(result.is_ok(), "default coerce should coerce \"42\": {result:?}");
    assert_eq!(result.unwrap(), json!(42));
}

#[test]
fn default_coerce_bool_true_from_string() {
    let s = bool_().coerce_default();
    let result = s.parse(&json!("true"));
    assert!(result.is_ok(), "default coerce should coerce \"true\": {result:?}");
    assert_eq!(result.unwrap(), json!(true));
}

#[test]
fn default_coerce_bool_false_from_string() {
    let s = bool_().coerce_default();
    let result = s.parse(&json!("false"));
    assert!(result.is_ok(), "default coerce should coerce \"false\": {result:?}");
    assert_eq!(result.unwrap(), json!(false));
}

#[test]
fn default_coerce_object_numeric_fields_from_strings() {
    let s = object()
        .field("age", Box::new(int().coerce_default()))
        .field("score", Box::new(number().coerce_default()))
        .field("active", Box::new(bool_().coerce_default()))
        .required(vec!["age", "score", "active"]);
    let input = json!({ "age": "42", "score": "3.14", "active": "true" });
    let result = s.parse(&input);
    assert!(
        result.is_ok(),
        "default coerce on object numeric/bool fields should coerce all string inputs: {result:?}"
    );
    assert_eq!(result.unwrap(), json!({ "age": 42, "score": 3.14, "active": true }));
}

// The no-arg ergonomic must be interchangeable with the explicit-token form:
// `.coerce_default()` == `.coerce(vec![])`, and an explicit `string->int` token
// still works alongside it.
#[test]
fn coerce_default_equivalent_to_empty_vec() {
    assert_eq!(
        int().coerce_default().safe_parse(&json!("42")).success,
        int().coerce(vec![]).safe_parse(&json!("42")).success,
    );
}

#[test]
fn explicit_token_still_works() {
    let s = int().coerce(vec!["string->int".to_string()]);
    assert_eq!(s.parse(&json!("42")).unwrap(), json!(42));
}

// ---------------------------------------------------------------------------
// CANONICAL COERCION MATRIX (all FROM STRING) via the no-arg `.coerce_default()`
// ergonomic. Every ACCEPT/REJECT row from the spec matrix is exercised here.
// ---------------------------------------------------------------------------

// string->int: ASCII ^-?\d+$ trimmed.
#[test]
fn matrix_default_coerce_int() {
    let s = int().coerce_default();
    for good in ["42", "  42  ", "-7"] {
        let r = s.safe_parse(&json!(good));
        assert!(r.success, "string->int must ACCEPT {good:?}: {:?}", r.issues);
    }
    for bad in ["3.14", "0x10", "1_000", "+5", "Infinity", "", "abc"] {
        let r = s.safe_parse(&json!(bad));
        assert!(!r.success, "string->int must REJECT {bad:?}");
        assert_eq!(r.issues[0].code, "coercion_failed", "{bad:?}");
    }
}

// string->number: ASCII decimal float incl exponent, trimmed.
#[test]
fn matrix_default_coerce_number() {
    let s = number().coerce_default();
    for good in ["3.14", "-1.5e3", "  2  ", "0"] {
        let r = s.safe_parse(&json!(good));
        assert!(r.success, "string->number must ACCEPT {good:?}: {:?}", r.issues);
    }
    for bad in ["0x10", "Infinity", "NaN", "", "1_000", "abc"] {
        let r = s.safe_parse(&json!(bad));
        assert!(!r.success, "string->number must REJECT {bad:?}");
        assert_eq!(r.issues[0].code, "coercion_failed", "{bad:?}");
    }
}

// string->bool: trim + case-insensitive; true<-"true"/"TRUE"/"1", false<-"false"/"0".
#[test]
fn matrix_default_coerce_bool() {
    let s = bool_().coerce_default();
    for (input, expected) in [
        ("true", true),
        ("TRUE", true),
        ("1", true),
        ("false", false),
        ("0", false),
    ] {
        let r = s.safe_parse(&json!(input));
        assert!(r.success, "string->bool must ACCEPT {input:?}: {:?}", r.issues);
        assert_eq!(r.value.unwrap(), json!(expected), "{input:?}");
    }
    for bad in ["yes", "no", "on", "off", "t", "f", "2", ""] {
        let r = s.safe_parse(&json!(bad));
        assert!(!r.success, "string->bool must REJECT {bad:?}");
        assert_eq!(r.issues[0].code, "coercion_failed", "{bad:?}");
    }
}

// string transforms (string kind): trim, lower, upper; chainable.
#[test]
fn matrix_string_transforms() {
    assert_eq!(
        string().coerce(vec!["trim".to_string()]).parse(&json!("  hi  ")).unwrap(),
        json!("hi")
    );
    assert_eq!(
        string().coerce(vec!["lower".to_string()]).parse(&json!("HeLLo")).unwrap(),
        json!("hello")
    );
    assert_eq!(
        string().coerce(vec!["upper".to_string()]).parse(&json!("HeLLo")).unwrap(),
        json!("HELLO")
    );
    // chainable
    assert_eq!(
        string()
            .coerce(vec!["trim".to_string(), "upper".to_string()])
            .parse(&json!("  HeLLo  "))
            .unwrap(),
        json!("HELLO")
    );
}
