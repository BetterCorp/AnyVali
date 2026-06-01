use anyvali::*;
use serde_json::json;
use std::time::Instant;

// =============================================================================
// CVE-2016-4055: ReDoS (Regular Expression Denial of Service)
//
// The Rust `regex` crate uses a finite automata engine that guarantees linear
// time matching, making it inherently resistant to ReDoS. These tests verify
// that catastrophic backtracking patterns complete promptly even with adversarial
// input, confirming the library does not fall back to an exponential-time engine.
// =============================================================================

#[test]
fn cve_2016_4055_redos_nested_quantifiers_completes_quickly() {
    // Classic ReDoS pattern: (a+)+ against a long string of 'a's followed by '!'
    // In a backtracking engine, this is O(2^n). Rust regex handles it in linear time.
    let s = string().pattern("^(a+)+$");
    let evil_input = "a".repeat(50_000) + "!";
    let start = Instant::now();
    let result = s.safe_parse(&json!(evil_input));
    let elapsed = start.elapsed();

    assert!(!result.success);
    // Should complete well within 1 second; a vulnerable engine would hang
    assert!(
        elapsed.as_secs() < 2,
        "ReDoS: nested quantifier pattern took {:?}, expected < 2s",
        elapsed
    );
}

#[test]
fn cve_2016_4055_redos_alternation_overlap_completes_quickly() {
    // Overlapping alternation: (a|a)* against adversarial input
    let s = string().pattern("^(a|a)*$");
    let evil_input = "a".repeat(50_000) + "!";
    let start = Instant::now();
    let result = s.safe_parse(&json!(evil_input));
    let elapsed = start.elapsed();

    assert!(!result.success);
    assert!(
        elapsed.as_secs() < 2,
        "ReDoS: alternation overlap pattern took {:?}, expected < 2s",
        elapsed
    );
}

#[test]
fn cve_2016_4055_redos_email_like_pattern_completes_quickly() {
    // Email-like ReDoS pattern that is catastrophic in PCRE engines
    let s = string().pattern("^([a-zA-Z0-9._-]+)*@[a-zA-Z0-9.-]+$");
    let evil_input = "a".repeat(50_000) + "@";
    let start = Instant::now();
    let result = s.safe_parse(&json!(evil_input));
    let elapsed = start.elapsed();

    // The pattern might match or not depending on the regex engine's interpretation,
    // but it must complete quickly regardless.
    let _ = result;
    assert!(
        elapsed.as_secs() < 2,
        "ReDoS: email-like pattern took {:?}, expected < 2s",
        elapsed
    );
}

// =============================================================================
// CVE-2003-1564: Recursive/Self-referencing $ref schemas
//
// A self-referencing $ref (where a definition references itself) could cause
// infinite recursion and stack overflow during validation. The importer and
// validator must handle this gracefully without crashing.
// =============================================================================

#[test]
fn cve_2003_1564_recursive_ref_self_referencing_import_does_not_crash() {
    // Schema where a definition references itself: User has a field "friend" that is a ref to User.
    let json_str = r##"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "ref", "ref": "#/definitions/User" },
        "definitions": {
            "User": {
                "kind": "object",
                "properties": {
                    "name": { "kind": "string" },
                    "friend": { "kind": "ref", "ref": "#/definitions/User" }
                },
                "required": ["name"],
                "unknownKeys": "reject"
            }
        },
        "extensions": {}
    }"##;

    // Import should succeed (the schema structure is valid; it is only during
    // unbounded recursive validation that a problem would manifest).
    let result = import(json_str);
    assert!(result.is_ok(), "Importing a self-referencing schema should not crash");
}

#[test]
fn cve_2003_1564_recursive_ref_mutual_reference_import_does_not_crash() {
    // Mutually recursive definitions: A references B, B references A
    let json_str = r##"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "ref", "ref": "#/definitions/A" },
        "definitions": {
            "A": {
                "kind": "object",
                "properties": {
                    "b": { "kind": "ref", "ref": "#/definitions/B" }
                },
                "required": [],
                "unknownKeys": "reject"
            },
            "B": {
                "kind": "object",
                "properties": {
                    "a": { "kind": "ref", "ref": "#/definitions/A" }
                },
                "required": [],
                "unknownKeys": "reject"
            }
        },
        "extensions": {}
    }"##;

    let result = import(json_str);
    assert!(result.is_ok(), "Importing mutually recursive schemas should not crash");
}

#[test]
fn cve_2003_1564_recursive_ref_unresolvable_ref_does_not_crash() {
    // A $ref that points to a nonexistent definition
    let json_str = r##"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "ref", "ref": "#/definitions/DoesNotExist" },
        "definitions": {},
        "extensions": {}
    }"##;

    let result = import(json_str);
    // Import may succeed (ref resolution is deferred) or fail gracefully
    // Either way, it must not panic
    let _ = result;
}

// =============================================================================
// CVE-2003-1564 parse-time: A pure self-cycle (Self -> ref Self) must NOT cause
// `safe_parse` to hang or stack-overflow the host process. Importer accepts the
// shape because resolution is deferred; this test exercises the actual recursion
// at parse time. Run in a sub-thread with a bounded stack so a stack overflow
// kills only the thread, not the test runner.
// =============================================================================

#[test]
fn cve_2003_1564_parse_time_self_cycle_does_not_hang_runtime() {
    let json_str = r##"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "ref", "ref": "#/definitions/Self" },
        "definitions": {
            "Self": { "kind": "ref", "ref": "#/definitions/Self" }
        },
        "extensions": {}
    }"##;

    // Tight 1 MiB stack so a self-cycle exhausts quickly.
    let handle = std::thread::Builder::new()
        .stack_size(1024 * 1024)
        .spawn(move || {
            let schema = import(json_str).expect("import");
            // Result here is irrelevant; we only need this call to *terminate*.
            let _ = schema.safe_parse(&json!("anything"));
        })
        .expect("spawn");

    // Join with a generous wall-clock bound. If the parse hangs, the test
    // hangs the runner — caught by the test framework's own timeout — but
    // we surface a clearer message via thread.is_finished polling.
    let start = Instant::now();
    while !handle.is_finished() {
        if start.elapsed().as_secs() > 5 {
            panic!("parse on self-cycle hung > 5s — runtime DoS");
        }
        std::thread::sleep(std::time::Duration::from_millis(50));
    }

    // join() yields Ok(()) on clean exit or Err on panic (e.g. stack
    // overflow). Either is acceptable — both prove non-hang.
    let _ = handle.join();
}

// =============================================================================
// CWE-190: Integer Overflow
//
// Validate that all integer width schemas correctly enforce their boundary
// values. Values at MAX+1 and MIN-1 must be rejected. Values at exactly MAX
// and MIN must be accepted.
// =============================================================================

#[test]
fn cwe_190_integer_overflow_int8_boundaries() {
    let s = int8();
    // Accept boundary values
    assert!(s.parse(&json!(i8::MAX as i64)).is_ok());   // 127
    assert!(s.parse(&json!(i8::MIN as i64)).is_ok());   // -128

    // Reject overflow values
    let above = s.safe_parse(&json!((i8::MAX as i64) + 1)); // 128
    assert!(!above.success);
    assert_eq!(above.issues[0].code, "too_large");

    let below = s.safe_parse(&json!((i8::MIN as i64) - 1)); // -129
    assert!(!below.success);
    assert_eq!(below.issues[0].code, "too_small");
}

#[test]
fn cwe_190_integer_overflow_int16_boundaries() {
    let s = int16();
    assert!(s.parse(&json!(i16::MAX as i64)).is_ok());  // 32767
    assert!(s.parse(&json!(i16::MIN as i64)).is_ok());  // -32768

    let above = s.safe_parse(&json!((i16::MAX as i64) + 1)); // 32768
    assert!(!above.success);
    assert_eq!(above.issues[0].code, "too_large");

    let below = s.safe_parse(&json!((i16::MIN as i64) - 1)); // -32769
    assert!(!below.success);
    assert_eq!(below.issues[0].code, "too_small");
}

#[test]
fn cwe_190_integer_overflow_int32_boundaries() {
    let s = int32();
    assert!(s.parse(&json!(i32::MAX as i64)).is_ok());  // 2147483647
    assert!(s.parse(&json!(i32::MIN as i64)).is_ok());  // -2147483648

    let above = s.safe_parse(&json!((i32::MAX as i64) + 1)); // 2147483648
    assert!(!above.success);
    assert_eq!(above.issues[0].code, "too_large");

    let below = s.safe_parse(&json!((i32::MIN as i64) - 1)); // -2147483649
    assert!(!below.success);
    assert_eq!(below.issues[0].code, "too_small");
}

#[test]
fn cwe_190_integer_overflow_int64_boundaries() {
    let s = int64();
    // i64::MAX = 9223372036854775807 -- JSON numbers may lose precision at this range
    // but the schema should still accept the value as parsed by serde_json
    assert!(s.parse(&json!(i64::MAX)).is_ok());
    assert!(s.parse(&json!(i64::MIN)).is_ok());
}

#[test]
fn cwe_190_integer_overflow_uint8_boundaries() {
    let s = uint8();
    assert!(s.parse(&json!(0)).is_ok());
    assert!(s.parse(&json!(u8::MAX as i64)).is_ok());   // 255

    let above = s.safe_parse(&json!((u8::MAX as i64) + 1)); // 256
    assert!(!above.success);
    assert_eq!(above.issues[0].code, "too_large");

    let below = s.safe_parse(&json!(-1));
    assert!(!below.success);
    assert_eq!(below.issues[0].code, "too_small");
}

#[test]
fn cwe_190_integer_overflow_uint16_boundaries() {
    let s = uint16();
    assert!(s.parse(&json!(0)).is_ok());
    assert!(s.parse(&json!(u16::MAX as i64)).is_ok());  // 65535

    let above = s.safe_parse(&json!((u16::MAX as i64) + 1)); // 65536
    assert!(!above.success);
    assert_eq!(above.issues[0].code, "too_large");

    let below = s.safe_parse(&json!(-1));
    assert!(!below.success);
    assert_eq!(below.issues[0].code, "too_small");
}

#[test]
fn cwe_190_integer_overflow_uint32_boundaries() {
    let s = uint32();
    assert!(s.parse(&json!(0)).is_ok());
    assert!(s.parse(&json!(u32::MAX as i64)).is_ok());  // 4294967295

    let above = s.safe_parse(&json!((u32::MAX as i64) + 1)); // 4294967296
    assert!(!above.success);
    assert_eq!(above.issues[0].code, "too_large");

    let below = s.safe_parse(&json!(-1));
    assert!(!below.success);
    assert_eq!(below.issues[0].code, "too_small");
}

#[test]
fn cwe_190_integer_overflow_uint64_rejects_negative() {
    let s = uint64();
    assert!(s.parse(&json!(0)).is_ok());

    let below = s.safe_parse(&json!(-1));
    assert!(!below.success);
    assert_eq!(below.issues[0].code, "too_small");
}

#[test]
fn cwe_190_integer_overflow_float_not_accepted_as_int() {
    // Ensure float values that look like boundary values but have fractional parts
    // are rejected by integer schemas
    let s = int8();
    let result = s.safe_parse(&json!(127.5));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "invalid_type");
}

// =============================================================================
// CWE-20: NaN and Infinity Handling
//
// IEEE 754 special values (NaN, Infinity, -Infinity) are not valid JSON.
// serde_json::json!() will not produce these values. We verify that if such
// values somehow reach the validator, they are handled gracefully.
// =============================================================================

#[test]
fn cwe_20_nan_infinity_serde_json_does_not_serialize_nan() {
    // serde_json::json!(f64::NAN) produces JSON null because NaN is not
    // representable in JSON. Verify the number schema rejects it (as null).
    let n = number();
    let val = json!(f64::NAN);
    // serde_json converts NaN to null
    assert!(val.is_null(), "serde_json should convert NaN to null");
    let result = n.safe_parse(&val);
    assert!(!result.success, "number() should reject null (from NaN)");
}

#[test]
fn cwe_20_nan_infinity_serde_json_does_not_serialize_infinity() {
    // serde_json::json!(f64::INFINITY) produces JSON null
    let n = number();
    let val = json!(f64::INFINITY);
    assert!(val.is_null(), "serde_json should convert Infinity to null");
    let result = n.safe_parse(&val);
    assert!(!result.success, "number() should reject null (from Infinity)");
}

#[test]
fn cwe_20_nan_infinity_serde_json_does_not_serialize_neg_infinity() {
    // serde_json::json!(f64::NEG_INFINITY) produces JSON null
    let n = number();
    let val = json!(f64::NEG_INFINITY);
    assert!(val.is_null(), "serde_json should convert -Infinity to null");
    let result = n.safe_parse(&val);
    assert!(!result.success, "number() should reject null (from -Infinity)");
}

#[test]
fn cwe_20_nan_infinity_int_rejects_null_from_nan() {
    let i = int();
    let val = json!(f64::NAN);
    let result = i.safe_parse(&val);
    assert!(!result.success, "int() should reject null (from NaN)");
}

#[test]
fn cwe_20_nan_infinity_float64_rejects_null_from_nan() {
    let f = float64();
    let val = json!(f64::NAN);
    let result = f.safe_parse(&val);
    assert!(!result.success, "float64() should reject null (from NaN)");
}

#[test]
fn cwe_20_nan_infinity_float32_rejects_null_from_infinity() {
    let f = float32();
    let val = json!(f64::INFINITY);
    let result = f.safe_parse(&val);
    assert!(!result.success, "float32() should reject null (from Infinity)");
}

// =============================================================================
// CWE-20: Format Validation Bypass
//
// Edge cases in format validation (email, url, ipv4) that could allow
// malicious or malformed values to pass validation.
// =============================================================================

#[test]
fn cwe_20_format_bypass_email_rejects_empty_local_part() {
    let s = string().format("email");
    let result = s.safe_parse(&json!("@example.com"));
    assert!(!result.success, "email format should reject empty local part");
}

#[test]
fn cwe_20_format_bypass_tampered_email_format_name_not_ignored() {
    let s = string().format("email\0");
    let result = s.safe_parse(&json!("not-an-email"));
    assert!(
        !result.success,
        "tampered format name must not strip email validation"
    );
}

#[test]
fn cwe_20_format_bypass_imported_tampered_email_format_not_unconstrained() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "string", "format": "email\u0000" },
        "definitions": {},
        "extensions": {}
    }"#;
    let schema = import(json_str).expect("import");
    let result = schema.safe_parse(&json!("not-an-email"));
    assert!(
        !result.success,
        "imported tampered format name must not strip email validation"
    );
}

#[test]
fn cwe_20_format_bypass_email_rejects_spaces() {
    let s = string().format("email");
    let result = s.safe_parse(&json!("user @example.com"));
    assert!(!result.success, "email format should reject spaces in local part");
}

#[test]
fn cwe_20_format_bypass_email_rejects_no_tld() {
    let s = string().format("email");
    let result = s.safe_parse(&json!("user@localhost"));
    assert!(!result.success, "email format should reject addresses without a TLD");
}

#[test]
fn cwe_20_format_bypass_email_rejects_double_at() {
    let s = string().format("email");
    let result = s.safe_parse(&json!("user@@example.com"));
    assert!(!result.success, "email format should reject double @ signs");
}

#[test]
fn cwe_20_format_bypass_email_accepts_valid_complex_address() {
    let s = string().format("email");
    assert!(s.parse(&json!("user.name+tag@example.co.uk")).is_ok());
}

#[test]
fn cwe_20_format_bypass_url_rejects_javascript_protocol() {
    let s = string().format("url");
    let result = s.safe_parse(&json!("javascript:alert(1)"));
    assert!(!result.success, "url format should reject javascript: protocol");
}

#[test]
fn cwe_20_format_bypass_url_rejects_data_protocol() {
    let s = string().format("url");
    let result = s.safe_parse(&json!("data:text/html,<script>alert(1)</script>"));
    assert!(!result.success, "url format should reject data: protocol");
}

#[test]
fn cwe_20_format_bypass_url_rejects_ftp_protocol() {
    let s = string().format("url");
    let result = s.safe_parse(&json!("ftp://files.example.com"));
    assert!(!result.success, "url format should reject ftp: protocol");
}

#[test]
fn cwe_20_format_bypass_url_rejects_empty_string() {
    let s = string().format("url");
    let result = s.safe_parse(&json!(""));
    assert!(!result.success, "url format should reject empty string");
}

#[test]
fn cwe_20_format_bypass_url_rejects_bare_domain() {
    let s = string().format("url");
    let result = s.safe_parse(&json!("example.com"));
    assert!(!result.success, "url format should reject bare domains without protocol");
}

#[test]
fn cwe_20_format_bypass_ipv4_rejects_leading_zeros() {
    let s = string().format("ipv4");
    // Leading zeros can be interpreted as octal in some systems (e.g., 010 = 8)
    let result = s.safe_parse(&json!("192.168.01.1"));
    assert!(!result.success, "ipv4 format should reject leading zeros (octal ambiguity)");
}

#[test]
fn cwe_20_format_bypass_ipv4_rejects_overflow_octet() {
    let s = string().format("ipv4");
    let result = s.safe_parse(&json!("256.1.1.1"));
    assert!(!result.success, "ipv4 format should reject octets > 255");
}

#[test]
fn cwe_20_format_bypass_ipv4_rejects_too_few_octets() {
    let s = string().format("ipv4");
    let result = s.safe_parse(&json!("192.168.1"));
    assert!(!result.success, "ipv4 format should reject too few octets");
}

#[test]
fn cwe_20_format_bypass_ipv4_rejects_too_many_octets() {
    let s = string().format("ipv4");
    let result = s.safe_parse(&json!("192.168.1.1.1"));
    assert!(!result.success, "ipv4 format should reject too many octets");
}

#[test]
fn cwe_20_format_bypass_ipv4_rejects_negative_octet() {
    let s = string().format("ipv4");
    let result = s.safe_parse(&json!("192.168.-1.1"));
    assert!(!result.success, "ipv4 format should reject negative octets");
}

#[test]
fn cwe_20_format_bypass_ipv4_rejects_non_numeric() {
    let s = string().format("ipv4");
    let result = s.safe_parse(&json!("192.168.abc.1"));
    assert!(!result.success, "ipv4 format should reject non-numeric octets");
}

#[test]
fn cwe_20_format_bypass_ipv4_accepts_boundary_octets() {
    let s = string().format("ipv4");
    assert!(s.parse(&json!("0.0.0.0")).is_ok());
    assert!(s.parse(&json!("255.255.255.255")).is_ok());
}

// =============================================================================
// Unicode length constraints
//
// minLength/maxLength are defined in Unicode code points, not UTF-8 bytes or
// UTF-16 code units. One astral symbol must count as length 1.
// =============================================================================

#[test]
fn unicode_length_astral_code_point_counts_as_one_character() {
    let emoji = "😀";
    assert!(string().max_length(1).safe_parse(&json!(emoji)).success);
    assert!(!string().min_length(2).safe_parse(&json!(emoji)).success);
}

#[test]
fn unicode_length_imported_max_length_uses_code_points() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "string", "maxLength": 1 },
        "definitions": {},
        "extensions": {}
    }"#;
    let schema = import(json_str).expect("import");
    assert!(schema.safe_parse(&json!("😀")).success);
}

// =============================================================================
// CWE-400: Resource Exhaustion via Large Inputs
//
// Verify that the library handles very large inputs without crashing,
// consuming excessive memory, or taking unreasonable time.
// =============================================================================

#[test]
fn cwe_400_large_input_very_long_string_does_not_crash() {
    let s = string();
    let long_string = "x".repeat(1_000_000); // 1MB string
    let start = Instant::now();
    let result = s.safe_parse(&json!(long_string));
    let elapsed = start.elapsed();

    assert!(result.success, "string() should accept a long string");
    assert!(
        elapsed.as_secs() < 5,
        "Parsing a 1MB string took {:?}, expected < 5s",
        elapsed
    );
}

#[test]
fn cwe_400_large_input_very_long_string_with_pattern_does_not_crash() {
    let s = string().pattern("^[a-z]+$");
    let long_string = "a".repeat(1_000_000);
    let start = Instant::now();
    let result = s.safe_parse(&json!(long_string));
    let elapsed = start.elapsed();

    assert!(result.success);
    assert!(
        elapsed.as_secs() < 5,
        "Pattern matching a 1MB string took {:?}, expected < 5s",
        elapsed
    );
}

#[test]
fn cwe_400_large_input_very_long_string_with_max_length_rejects() {
    let s = string().max_length(100);
    let long_string = "x".repeat(1_000_000);
    let result = s.safe_parse(&json!(long_string));
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
}

#[test]
fn cwe_400_large_input_large_array_does_not_crash() {
    let a = array(Box::new(int()));
    // Build a large array of integers
    let large_array: Vec<serde_json::Value> = (0..100_000).map(|i| json!(i)).collect();
    let val = serde_json::Value::Array(large_array);

    let start = Instant::now();
    let result = a.safe_parse(&val);
    let elapsed = start.elapsed();

    assert!(result.success, "array(int()) should accept a large array");
    assert!(
        elapsed.as_secs() < 10,
        "Parsing a 100k-element array took {:?}, expected < 10s",
        elapsed
    );
}

#[test]
fn cwe_400_large_input_large_array_with_max_items_rejects() {
    let a = array(Box::new(int())).max_items(10);
    let large_array: Vec<serde_json::Value> = (0..10_000).map(|i| json!(i)).collect();
    let val = serde_json::Value::Array(large_array);

    let result = a.safe_parse(&val);
    assert!(!result.success);
    assert_eq!(result.issues[0].code, "too_large");
}

#[test]
fn cwe_400_large_input_deeply_nested_object_does_not_crash() {
    // Build a deeply nested JSON object: {"a": {"a": {"a": ... }}}
    let mut val = json!("leaf");
    for _ in 0..100 {
        val = json!({"a": val});
    }

    // An any() schema should accept deeply nested values
    let a = any();
    let result = a.safe_parse(&val);
    assert!(result.success);
}

#[test]
fn cwe_400_large_input_large_object_many_keys_does_not_crash() {
    let r = record(Box::new(string()));
    let mut map = serde_json::Map::new();
    for i in 0..10_000 {
        map.insert(format!("key_{}", i), json!(format!("value_{}", i)));
    }
    let val = serde_json::Value::Object(map);

    let start = Instant::now();
    let result = r.safe_parse(&val);
    let elapsed = start.elapsed();

    assert!(result.success);
    assert!(
        elapsed.as_secs() < 10,
        "Parsing a 10k-key record took {:?}, expected < 10s",
        elapsed
    );
}

// =============================================================================
// Schema Import Injection: Unknown/Malicious Schema Kinds
//
// The importer must reject schema documents containing unrecognized or
// potentially malicious "kind" values, preventing schema injection attacks.
// =============================================================================

#[test]
fn schema_import_injection_rejects_unknown_kind() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "exec" },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
    assert!(
        result.unwrap_err().contains("Unsupported schema kind"),
        "Import should report unsupported kind"
    );
}

#[test]
fn schema_import_injection_rejects_sql_injection_kind() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "'; DROP TABLE schemas;--" },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
    assert!(result.unwrap_err().contains("Unsupported schema kind"));
}

#[test]
fn schema_import_injection_rejects_path_traversal_kind() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "../../../etc/passwd" },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
    assert!(result.unwrap_err().contains("Unsupported schema kind"));
}

#[test]
fn schema_import_injection_rejects_empty_kind() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "" },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
}

#[test]
fn schema_import_injection_rejects_missing_kind_field() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "value": "hello" },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
}

#[test]
fn schema_import_injection_rejects_numeric_kind() {
    // kind must be a string, not a number
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": 42 },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
}

#[test]
fn schema_import_injection_rejects_null_kind() {
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": null },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err());
}

#[test]
fn schema_import_injection_unknown_kind_in_definition() {
    // Malicious kind inside a definition, not just the root
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": { "kind": "string" },
        "definitions": {
            "Evil": { "kind": "system_exec" }
        },
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err(), "Unknown kinds in definitions must also be rejected");
}

#[test]
fn schema_import_injection_unknown_kind_in_nested_schema() {
    // Unknown kind nested inside an array items schema
    let json_str = r#"{
        "anyvaliVersion": "1.0",
        "schemaVersion": "1",
        "root": {
            "kind": "array",
            "items": { "kind": "remote_code_execution" }
        },
        "definitions": {},
        "extensions": {}
    }"#;
    let result = import(json_str);
    assert!(result.is_err(), "Unknown kinds in nested schemas must also be rejected");
}

#[test]
fn schema_import_injection_valid_kinds_all_accepted() {
    // Verify that all legitimate kinds are still accepted
    let valid_kinds = vec![
        "string", "number", "float64", "float32",
        "int", "int8", "int16", "int32", "int64",
        "uint8", "uint16", "uint32", "uint64",
        "bool", "null", "any", "unknown", "never",
    ];
    for kind in valid_kinds {
        let json_str = format!(
            r#"{{"anyvaliVersion":"1.0","schemaVersion":"1","root":{{"kind":"{}"}},"definitions":{{}},"extensions":{{}}}}"#,
            kind
        );
        let result = import(&json_str);
        assert!(
            result.is_ok(),
            "Legitimate kind '{}' should be accepted, but got error: {:?}",
            kind,
            result.err()
        );
    }
}
