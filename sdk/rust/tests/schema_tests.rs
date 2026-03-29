use anyvali::*;
use serde_json::json;

// ===== String Tests =====

#[test]
fn string_accepts_simple_string() {
    let s = string();
    let result = s.parse(&json!("hello"));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!("hello"));
}

#[test]
fn string_accepts_empty_string() {
    let s = string();
    assert!(s.parse(&json!("")).is_ok());
}

#[test]
fn string_rejects_number() {
    let s = string();
    let result = s.safe_parse(&json!(42));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].expected, "string");
    assert_eq!(result.issues[0].received, "number");
}

#[test]
fn string_rejects_boolean() {
    let s = string();
    let result = s.safe_parse(&json!(true));
    assert!(!result.success);
    assert_eq!(result.issues[0].received, "boolean");
}

#[test]
fn string_rejects_null() {
    let s = string();
    let result = s.safe_parse(&json!(null));
    assert!(!result.success);
    assert_eq!(result.issues[0].received, "null");
}

#[test]
fn string_rejects_array() {
    let s = string();
    let result = s.safe_parse(&json!(["a", "b"]));
    assert!(!result.success);
    assert_eq!(result.issues[0].received, "array");
}

#[test]
fn string_rejects_object() {
    let s = string();
    let result = s.safe_parse(&json!({"key": "value"}));
    assert!(!result.success);
    assert_eq!(result.issues[0].received, "object");
}

#[test]
fn string_min_length_passes() {
    let s = string().min_length(3);
    assert!(s.parse(&json!("abc")).is_ok());
}

#[test]
fn string_min_length_fails() {
    let s = string().min_length(3);
    let result = s.safe_parse(&json!("ab"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_small");
    assert_eq!(result.issues[0].expected, "3");
    assert_eq!(result.issues[0].received, "2");
}

#[test]
fn string_max_length_passes() {
    let s = string().max_length(5);
    assert!(s.parse(&json!("hello")).is_ok());
}

#[test]
fn string_max_length_fails() {
    let s = string().max_length(5);
    let result = s.safe_parse(&json!("hello!"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
}

#[test]
fn string_pattern_passes() {
    let s = string().pattern("^[a-z]+$");
    assert!(s.parse(&json!("abc")).is_ok());
}

#[test]
fn string_pattern_fails() {
    let s = string().pattern("^[a-z]+$");
    let result = s.safe_parse(&json!("ABC"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_string");
}

#[test]
fn string_starts_with_passes() {
    let s = string().starts_with("hello");
    assert!(s.parse(&json!("hello world")).is_ok());
}

#[test]
fn string_starts_with_fails() {
    let s = string().starts_with("hello");
    let result = s.safe_parse(&json!("world hello"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_string");
}

#[test]
fn string_ends_with_passes() {
    let s = string().ends_with(".json");
    assert!(s.parse(&json!("file.json")).is_ok());
}

#[test]
fn string_ends_with_fails() {
    let s = string().ends_with(".json");
    let result = s.safe_parse(&json!("file.xml"));
    assert!(!result.success);
}

#[test]
fn string_includes_passes() {
    let s = string().includes("world");
    assert!(s.parse(&json!("hello world!")).is_ok());
}

#[test]
fn string_includes_fails() {
    let s = string().includes("world");
    let result = s.safe_parse(&json!("hello there"));
    assert!(!result.success);
}

#[test]
fn string_combined_constraints() {
    let s = string().min_length(1).max_length(100);
    assert!(s.parse(&json!("test")).is_ok());
    assert!(s.safe_parse(&json!("")).issues.len() > 0);
}

// ===== Number Tests =====

#[test]
fn number_accepts_positive() {
    let n = number();
    assert!(n.parse(&json!(42)).is_ok());
}

#[test]
fn number_accepts_zero() {
    let n = number();
    assert!(n.parse(&json!(0)).is_ok());
}

#[test]
fn number_accepts_negative() {
    let n = number();
    assert!(n.parse(&json!(-3.14)).is_ok());
}

#[test]
fn number_accepts_large_float() {
    let n = number();
    assert!(n.parse(&json!(1.7976931348623157e+308)).is_ok());
}

#[test]
fn number_rejects_string() {
    let n = number();
    let result = n.safe_parse(&json!("42"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].expected, "number");
    assert_eq!(result.issues[0].received, "string");
}

#[test]
fn number_rejects_boolean() {
    let n = number();
    let result = n.safe_parse(&json!(true));
    assert!(!result.success);
}

#[test]
fn number_rejects_null() {
    let n = number();
    let result = n.safe_parse(&json!(null));
    assert!(!result.success);
}

#[test]
fn number_rejects_object() {
    let n = number();
    let result = n.safe_parse(&json!({}));
    assert!(!result.success);
}

#[test]
fn number_rejects_array() {
    let n = number();
    let result = n.safe_parse(&json!([1, 2, 3]));
    assert!(!result.success);
}

#[test]
fn number_min_passes() {
    let n = number().min(10.0);
    assert!(n.parse(&json!(10)).is_ok());
}

#[test]
fn number_min_fails() {
    let n = number().min(10.0);
    let result = n.safe_parse(&json!(9));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_small");
}

#[test]
fn number_max_passes() {
    let n = number().max(100.0);
    assert!(n.parse(&json!(100)).is_ok());
}

#[test]
fn number_max_fails() {
    let n = number().max(100.0);
    let result = n.safe_parse(&json!(101));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
}

#[test]
fn number_exclusive_min_passes() {
    let n = number().exclusive_min(0.0);
    assert!(n.parse(&json!(0.001)).is_ok());
}

#[test]
fn number_exclusive_min_fails() {
    let n = number().exclusive_min(0.0);
    let result = n.safe_parse(&json!(0));
    assert!(!result.success);
}

#[test]
fn number_exclusive_max_passes() {
    let n = number().exclusive_max(100.0);
    assert!(n.parse(&json!(99.999)).is_ok());
}

#[test]
fn number_exclusive_max_fails() {
    let n = number().exclusive_max(100.0);
    let result = n.safe_parse(&json!(100));
    assert!(!result.success);
}

#[test]
fn number_multiple_of_passes() {
    let n = number().multiple_of(3.0);
    assert!(n.parse(&json!(9)).is_ok());
}

#[test]
fn number_multiple_of_fails() {
    let n = number().multiple_of(3.0);
    let result = n.safe_parse(&json!(10));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_number");
}

#[test]
fn number_multiple_of_float() {
    let n = number().multiple_of(0.5);
    assert!(n.parse(&json!(2.5)).is_ok());
}

// ===== Int Tests =====

#[test]
fn int_accepts_positive() {
    let i = int();
    assert!(i.parse(&json!(42)).is_ok());
}

#[test]
fn int_accepts_zero() {
    let i = int();
    assert!(i.parse(&json!(0)).is_ok());
}

#[test]
fn int_accepts_negative() {
    let i = int();
    assert!(i.parse(&json!(-100)).is_ok());
}

#[test]
fn int_rejects_float() {
    let i = int();
    let result = i.safe_parse(&json!(3.14));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].expected, "int");
    assert_eq!(result.issues[0].received, "number");
}

#[test]
fn int_rejects_string() {
    let i = int();
    let result = i.safe_parse(&json!("42"));
    assert!(!result.success);
    assert_eq!(result.issues[0].received, "string");
}

#[test]
fn int_with_constraints() {
    let i = int().min(1.0).max(10.0);
    assert!(i.parse(&json!(5)).is_ok());
    assert!(!i.safe_parse(&json!(0)).success);
    assert!(!i.safe_parse(&json!(11)).success);
}

// ===== Int Width Tests =====

#[test]
fn int8_accepts_in_range() {
    let i = int8();
    assert!(i.parse(&json!(127)).is_ok());
    assert!(i.parse(&json!(-128)).is_ok());
}

#[test]
fn int8_rejects_above_range() {
    let i = int8();
    let result = i.safe_parse(&json!(128));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
    assert_eq!(result.issues[0].expected, "int8");
}

#[test]
fn int8_rejects_below_range() {
    let i = int8();
    let result = i.safe_parse(&json!(-129));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_small");
}

#[test]
fn int16_accepts_in_range() {
    let i = int16();
    assert!(i.parse(&json!(32767)).is_ok());
}

#[test]
fn int16_rejects_above_range() {
    let i = int16();
    let result = i.safe_parse(&json!(32768));
    assert!(!result.success);
}

#[test]
fn int32_accepts_max() {
    let i = int32();
    assert!(i.parse(&json!(2147483647)).is_ok());
}

#[test]
fn int32_rejects_above_range() {
    let i = int32();
    let result = i.safe_parse(&json!(2147483648_i64));
    assert!(!result.success);
}

#[test]
fn uint8_accepts_valid() {
    let u = uint8();
    assert!(u.parse(&json!(0)).is_ok());
    assert!(u.parse(&json!(255)).is_ok());
}

#[test]
fn uint8_rejects_negative() {
    let u = uint8();
    let result = u.safe_parse(&json!(-1));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_small");
}

#[test]
fn uint8_rejects_256() {
    let u = uint8();
    let result = u.safe_parse(&json!(256));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
}

#[test]
fn uint16_accepts_max() {
    let u = uint16();
    assert!(u.parse(&json!(65535)).is_ok());
}

#[test]
fn uint32_accepts_max() {
    let u = uint32();
    assert!(u.parse(&json!(4294967295_u64)).is_ok());
}

#[test]
fn uint64_rejects_negative() {
    let u = uint64();
    let result = u.safe_parse(&json!(-1));
    assert!(!result.success);
}

// ===== Float Width Tests =====

#[test]
fn float64_accepts_float() {
    let f = float64();
    assert!(f.parse(&json!(3.141592653589793)).is_ok());
}

#[test]
fn float64_accepts_int_as_float() {
    let f = float64();
    assert!(f.parse(&json!(42)).is_ok());
}

#[test]
fn float64_rejects_string() {
    let f = float64();
    let result = f.safe_parse(&json!("3.14"));
    assert!(!result.success);
    assert_eq!(result.issues[0].expected, "float64");
}

#[test]
fn float32_accepts_float() {
    let f = float32();
    assert!(f.parse(&json!(1.5)).is_ok());
}

#[test]
fn float32_rejects_boolean() {
    let f = float32();
    let result = f.safe_parse(&json!(true));
    assert!(!result.success);
    assert_eq!(result.issues[0].expected, "float32");
}

// ===== Bool Tests =====

#[test]
fn bool_accepts_true() {
    let b = bool_();
    assert!(b.parse(&json!(true)).is_ok());
}

#[test]
fn bool_accepts_false() {
    let b = bool_();
    assert!(b.parse(&json!(false)).is_ok());
}

#[test]
fn bool_rejects_number() {
    let b = bool_();
    let result = b.safe_parse(&json!(1));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
}

#[test]
fn bool_rejects_string() {
    let b = bool_();
    let result = b.safe_parse(&json!("true"));
    assert!(!result.success);
}

#[test]
fn bool_rejects_null() {
    let b = bool_();
    let result = b.safe_parse(&json!(null));
    assert!(!result.success);
}

// ===== Null Tests =====

#[test]
fn null_accepts_null() {
    let n = null();
    assert!(n.parse(&json!(null)).is_ok());
}

#[test]
fn null_rejects_string() {
    let n = null();
    let result = n.safe_parse(&json!("null"));
    assert!(!result.success);
}

#[test]
fn null_rejects_zero() {
    let n = null();
    let result = n.safe_parse(&json!(0));
    assert!(!result.success);
}

#[test]
fn null_rejects_false() {
    let n = null();
    let result = n.safe_parse(&json!(false));
    assert!(!result.success);
}

// ===== Any Tests =====

#[test]
fn any_accepts_string() {
    let a = any();
    assert!(a.parse(&json!("hello")).is_ok());
}

#[test]
fn any_accepts_number() {
    let a = any();
    assert!(a.parse(&json!(42)).is_ok());
}

#[test]
fn any_accepts_null() {
    let a = any();
    assert!(a.parse(&json!(null)).is_ok());
}

#[test]
fn any_accepts_object() {
    let a = any();
    assert!(a.parse(&json!({"key": "value"})).is_ok());
}

#[test]
fn any_accepts_array() {
    let a = any();
    assert!(a.parse(&json!([1, "two", true])).is_ok());
}

// ===== Unknown Tests =====

#[test]
fn unknown_accepts_all_types() {
    let u = unknown();
    assert!(u.parse(&json!("hello")).is_ok());
    assert!(u.parse(&json!(99)).is_ok());
    assert!(u.parse(&json!(null)).is_ok());
    assert!(u.parse(&json!(false)).is_ok());
    assert!(u.parse(&json!({"a": [1, {"b": true}]})).is_ok());
}

// ===== Never Tests =====

#[test]
fn never_rejects_everything() {
    let n = never();
    assert!(!n.safe_parse(&json!("hello")).success);
    assert!(!n.safe_parse(&json!(0)).success);
    assert!(!n.safe_parse(&json!(null)).success);
    assert!(!n.safe_parse(&json!(true)).success);
    assert!(!n.safe_parse(&json!({})).success);
}

#[test]
fn never_correct_issue() {
    let n = never();
    let result = n.safe_parse(&json!("hello"));
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].expected, "never");
    assert_eq!(result.issues[0].received, "string");
}

// ===== Literal Tests =====

#[test]
fn literal_accepts_matching_string() {
    let l = literal(json!("hello"));
    assert!(l.parse(&json!("hello")).is_ok());
}

#[test]
fn literal_rejects_nonmatching_string() {
    let l = literal(json!("hello"));
    let result = l.safe_parse(&json!("world"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_literal");
}

#[test]
fn literal_accepts_matching_number() {
    let l = literal(json!(42));
    assert!(l.parse(&json!(42)).is_ok());
}

#[test]
fn literal_accepts_matching_bool() {
    let l = literal(json!(true));
    assert!(l.parse(&json!(true)).is_ok());
}

#[test]
fn literal_rejects_wrong_type() {
    let l = literal(json!(42));
    let result = l.safe_parse(&json!("42"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_literal");
}

#[test]
fn literal_accepts_null() {
    let l = literal(json!(null));
    assert!(l.parse(&json!(null)).is_ok());
}

// ===== Enum Tests =====

#[test]
fn enum_accepts_valid_value() {
    let e = enum_(vec![json!("red"), json!("green"), json!("blue")]);
    assert!(e.parse(&json!("red")).is_ok());
    assert!(e.parse(&json!("blue")).is_ok());
}

#[test]
fn enum_rejects_invalid_value() {
    let e = enum_(vec![json!("red"), json!("green"), json!("blue")]);
    let result = e.safe_parse(&json!("yellow"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].expected, "enum(red,green,blue)");
}

#[test]
fn enum_accepts_numeric() {
    let e = enum_(vec![json!(1), json!(2), json!(3)]);
    assert!(e.parse(&json!(2)).is_ok());
}

#[test]
fn enum_rejects_wrong_type() {
    let e = enum_(vec![json!(1), json!(2), json!(3)]);
    let result = e.safe_parse(&json!("1"));
    assert!(!result.success);
}

// ===== Array Tests =====

#[test]
fn array_accepts_valid_elements() {
    let a = array(Box::new(string()));
    assert_eq!(
        a.parse(&json!(["a", "b", "c"])).unwrap(),
        json!(["a", "b", "c"])
    );
}

#[test]
fn array_accepts_empty() {
    let a = array(Box::new(int()));
    assert!(a.parse(&json!([])).is_ok());
}

#[test]
fn array_rejects_non_array() {
    let a = array(Box::new(string()));
    let result = a.safe_parse(&json!("not an array"));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].expected, "array");
}

#[test]
fn array_rejects_invalid_element() {
    let a = array(Box::new(int()));
    let result = a.safe_parse(&json!([1, 2, "three"]));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].path, vec![PathSegment::Index(2)]);
}

#[test]
fn array_reports_multiple_invalid_elements() {
    let a = array(Box::new(bool_()));
    let result = a.safe_parse(&json!([true, "yes", false, 1]));
    assert!(!result.success);
    assert_eq!(result.issues.len(), 2);
    assert_eq!(result.issues[0].path, vec![PathSegment::Index(1)]);
    assert_eq!(result.issues[1].path, vec![PathSegment::Index(3)]);
}

#[test]
fn array_min_items_passes() {
    let a = array(Box::new(int())).min_items(2);
    assert!(a.parse(&json!([1, 2])).is_ok());
}

#[test]
fn array_min_items_fails() {
    let a = array(Box::new(int())).min_items(2);
    let result = a.safe_parse(&json!([1]));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_small");
}

#[test]
fn array_max_items_passes() {
    let a = array(Box::new(string())).max_items(3);
    assert!(a.parse(&json!(["a", "b", "c"])).is_ok());
}

#[test]
fn array_max_items_fails() {
    let a = array(Box::new(string())).max_items(3);
    let result = a.safe_parse(&json!(["a", "b", "c", "d"]));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
}

// ===== Tuple Tests =====

#[test]
fn tuple_accepts_valid() {
    let t = tuple(vec![Box::new(string()), Box::new(int())]);
    assert_eq!(
        t.parse(&json!(["hello", 42])).unwrap(),
        json!(["hello", 42])
    );
}

#[test]
fn tuple_rejects_too_few() {
    let t = tuple(vec![Box::new(string()), Box::new(int())]);
    let result = t.safe_parse(&json!(["hello"]));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_small");
}

#[test]
fn tuple_rejects_too_many() {
    let t = tuple(vec![Box::new(string()), Box::new(int())]);
    let result = t.safe_parse(&json!(["hello", 42, true]));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
}

#[test]
fn tuple_rejects_wrong_element_type() {
    let t = tuple(vec![Box::new(string()), Box::new(int())]);
    let result = t.safe_parse(&json!([42, "hello"]));
    assert!(!result.success);
    assert_eq!(result.issues.len(), 2);
}

#[test]
fn tuple_rejects_non_array() {
    let t = tuple(vec![Box::new(string())]);
    let result = t.safe_parse(&json!("not a tuple"));
    assert!(!result.success);
    assert_eq!(result.issues[0].expected, "tuple");
}

// ===== Object Tests =====

#[test]
fn object_accepts_valid() {
    let o = object()
        .field("name", Box::new(string()))
        .field("age", Box::new(int()))
        .required(vec!["name", "age"]);
    let result = o.parse(&json!({"name": "Alice", "age": 30}));
    assert!(result.is_ok());
}

#[test]
fn object_rejects_missing_required() {
    let o = object()
        .field("name", Box::new(string()))
        .field("age", Box::new(int()))
        .required(vec!["name", "age"]);
    let result = o.safe_parse(&json!({"name": "Alice"}));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "required");
    assert_eq!(
        result.issues[0].path,
        vec![PathSegment::Key("age".to_string())]
    );
}

#[test]
fn object_rejects_all_missing() {
    let o = object()
        .field("name", Box::new(string()))
        .field("age", Box::new(int()))
        .required(vec!["name", "age"]);
    let result = o.safe_parse(&json!({}));
    assert!(!result.success);
    assert_eq!(result.issues.len(), 2);
}

#[test]
fn object_accepts_optional_absent() {
    let o = object()
        .field("name", Box::new(string()))
        .field("nickname", Box::new(string()))
        .required(vec!["name"]);
    let result = o.parse(&json!({"name": "Alice"}));
    assert!(result.is_ok());
}

#[test]
fn object_rejects_non_object() {
    let o = object()
        .field("name", Box::new(string()))
        .required(vec!["name"]);
    let result = o.safe_parse(&json!("not an object"));
    assert!(!result.success);
    assert_eq!(result.issues[0].expected, "object");
}

#[test]
fn object_unknown_keys_reject_default() {
    let o = object()
        .field("name", Box::new(string()))
        .required(vec!["name"]);
    let result = o.safe_parse(&json!({"name": "Alice", "extra": "value"}));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "unknown_key");
}

#[test]
fn object_unknown_keys_strip() {
    let o = object()
        .field("name", Box::new(string()))
        .required(vec!["name"])
        .unknown_keys(UnknownKeyMode::Strip);
    let result = o.parse(&json!({"name": "Alice", "extra": "value", "another": 42}));
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), json!({"name": "Alice"}));
}

#[test]
fn object_unknown_keys_allow() {
    let o = object()
        .field("name", Box::new(string()))
        .required(vec!["name"])
        .unknown_keys(UnknownKeyMode::Allow);
    let result = o.parse(&json!({"name": "Alice", "extra": "value"}));
    assert!(result.is_ok());
    let val = result.unwrap();
    assert_eq!(val["name"], json!("Alice"));
    assert_eq!(val["extra"], json!("value"));
}

#[test]
fn object_unknown_keys_reject_multiple() {
    let o = object()
        .field("id", Box::new(int()))
        .required(vec!["id"]);
    let result = o.safe_parse(&json!({"id": 1, "foo": "bar", "baz": true}));
    assert!(!result.success);
    assert_eq!(result.issues.len(), 2);
    assert!(result.issues.iter().all(|i| i.code == "unknown_key"));
}

// ===== Record Tests =====

#[test]
fn record_accepts_valid() {
    let r = record(Box::new(int()));
    assert_eq!(
        r.parse(&json!({"a": 1, "b": 2, "c": 3})).unwrap(),
        json!({"a": 1, "b": 2, "c": 3})
    );
}

#[test]
fn record_accepts_empty() {
    let r = record(Box::new(string()));
    assert!(r.parse(&json!({})).is_ok());
}

#[test]
fn record_rejects_invalid_value() {
    let r = record(Box::new(int()));
    let result = r.safe_parse(&json!({"a": 1, "b": "two"}));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(
        result.issues[0].path,
        vec![PathSegment::Key("b".to_string())]
    );
}

#[test]
fn record_rejects_non_object() {
    let r = record(Box::new(string()));
    let result = r.safe_parse(&json!([1, 2, 3]));
    assert!(!result.success);
    assert_eq!(result.issues[0].expected, "record");
    assert_eq!(result.issues[0].received, "array");
}

// ===== Union Tests =====

#[test]
fn union_accepts_first_variant() {
    let u = union(vec![Box::new(string()), Box::new(int())]);
    assert!(u.parse(&json!("hello")).is_ok());
}

#[test]
fn union_accepts_second_variant() {
    let u = union(vec![Box::new(string()), Box::new(int())]);
    assert!(u.parse(&json!(42)).is_ok());
}

#[test]
fn union_rejects_no_match() {
    let u = union(vec![Box::new(string()), Box::new(int())]);
    let result = u.safe_parse(&json!(true));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_union");
    assert_eq!(result.issues[0].expected, "string | int");
    assert_eq!(result.issues[0].received, "boolean");
}

#[test]
fn union_first_match_wins() {
    let u = union(vec![Box::new(number()), Box::new(int())]);
    // 5 matches number first
    assert!(u.parse(&json!(5)).is_ok());
}

#[test]
fn union_with_null() {
    let u = union(vec![Box::new(string()), Box::new(null())]);
    assert!(u.parse(&json!(null)).is_ok());
}

// ===== Intersection Tests =====

#[test]
fn intersection_accepts_satisfying_all() {
    let i = intersection(vec![
        Box::new(
            object()
                .field("name", Box::new(string()))
                .required(vec!["name"])
                .unknown_keys(UnknownKeyMode::Allow),
        ),
        Box::new(
            object()
                .field("age", Box::new(int()))
                .required(vec!["age"])
                .unknown_keys(UnknownKeyMode::Allow),
        ),
    ]);
    assert!(i.parse(&json!({"name": "Alice", "age": 30})).is_ok());
}

#[test]
fn intersection_rejects_missing_from_second() {
    let i = intersection(vec![
        Box::new(
            object()
                .field("name", Box::new(string()))
                .required(vec!["name"])
                .unknown_keys(UnknownKeyMode::Allow),
        ),
        Box::new(
            object()
                .field("age", Box::new(int()))
                .required(vec!["age"])
                .unknown_keys(UnknownKeyMode::Allow),
        ),
    ]);
    let result = i.safe_parse(&json!({"name": "Alice"}));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "required");
}

#[test]
fn intersection_numeric_ranges() {
    let i = intersection(vec![
        Box::new(number().min(0.0)),
        Box::new(number().max(100.0)),
    ]);
    assert!(i.parse(&json!(50)).is_ok());
    assert!(!i.safe_parse(&json!(-5)).success);
}

// ===== Optional Tests =====

#[test]
fn optional_accepts_present_valid() {
    let o = object()
        .field("name", Box::new(optional(Box::new(string()))))
        .required(vec![]);
    assert!(o.parse(&json!({"name": "Alice"})).is_ok());
}

#[test]
fn optional_accepts_absent() {
    let o = object()
        .field("name", Box::new(optional(Box::new(string()))))
        .required(vec![]);
    assert!(o.parse(&json!({})).is_ok());
}

#[test]
fn optional_rejects_present_invalid() {
    let o = object()
        .field("name", Box::new(optional(Box::new(string()))))
        .required(vec![]);
    let result = o.safe_parse(&json!({"name": 123}));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
}

#[test]
fn optional_null_is_not_absent() {
    let o = object()
        .field("name", Box::new(optional(Box::new(string()))))
        .required(vec![]);
    let result = o.safe_parse(&json!({"name": null}));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].received, "null");
}

// ===== Nullable Tests =====

#[test]
fn nullable_accepts_null() {
    let n = nullable(Box::new(string()));
    assert!(n.parse(&json!(null)).is_ok());
}

#[test]
fn nullable_accepts_valid_nonnull() {
    let n = nullable(Box::new(string()));
    assert!(n.parse(&json!("hello")).is_ok());
}

#[test]
fn nullable_rejects_invalid_nonnull() {
    let n = nullable(Box::new(string()));
    let result = n.safe_parse(&json!(42));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
    assert_eq!(result.issues[0].expected, "string");
}

#[test]
fn nullable_int_accepts_null() {
    let n = nullable(Box::new(int()));
    assert!(n.parse(&json!(null)).is_ok());
}

#[test]
fn nullable_int_accepts_valid() {
    let n = nullable(Box::new(int()));
    assert!(n.parse(&json!(99)).is_ok());
}

// ===== Ref Tests (manual) =====

#[test]
fn ref_resolves_definition() {
    use std::collections::HashMap;

    let user_schema = object()
        .field("name", Box::new(string()))
        .field("age", Box::new(int()))
        .required(vec!["name", "age"]);

    let root = object()
        .field("user", Box::new(ref_("#/definitions/User")))
        .required(vec!["user"]);

    let mut definitions: HashMap<String, Box<dyn Schema>> = HashMap::new();
    definitions.insert("User".to_string(), Box::new(user_schema));

    let ctx = ParseContext { definitions };
    let result = root.parse_with_context(
        &json!({"user": {"name": "Alice", "age": 30}}),
        &ctx,
    );
    assert!(result.is_ok());
}

#[test]
fn ref_validates_against_definition() {
    use std::collections::HashMap;

    let user_schema = object()
        .field("name", Box::new(string()))
        .field("age", Box::new(int()))
        .required(vec!["name", "age"]);

    let root = object()
        .field("user", Box::new(ref_("#/definitions/User")))
        .required(vec!["user"]);

    let mut definitions: HashMap<String, Box<dyn Schema>> = HashMap::new();
    definitions.insert("User".to_string(), Box::new(user_schema));

    let ctx = ParseContext { definitions };
    let result = root.safe_parse_with_context(
        &json!({"user": {"name": "Alice"}}),
        &ctx,
    );
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "required");
}

// ===== Schema Clone Tests =====

#[test]
fn schema_clone_works() {
    let s1 = string().min_length(3);
    let s2 = s1.clone();
    assert!(s2.parse(&json!("abc")).is_ok());
    assert!(!s2.safe_parse(&json!("ab")).success);
}

// ===== Safe Parse vs Parse Tests =====

#[test]
fn parse_returns_error_on_failure() {
    let s = string();
    let result = s.parse(&json!(42));
    assert!(result.is_err());
    let err = result.unwrap_err();
    assert_eq!(err.issues.len(), 1);
}

#[test]
fn safe_parse_returns_parse_result() {
    let s = string();

    let ok = s.safe_parse(&json!("hello"));
    assert!(ok.success);
    assert_eq!(ok.value, Some(json!("hello")));
    assert!(ok.issues.is_empty());

    let fail = s.safe_parse(&json!(42));
    assert!(!fail.success);
    assert!(fail.value.is_none());
    assert!(!fail.issues.is_empty());
}
