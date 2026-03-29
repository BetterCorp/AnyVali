use serde_json::Value;
use std::fs;
use std::path::Path;

/// Run all conformance test cases from a corpus JSON file.
fn run_corpus_file(file_path: &str) {
    let path = Path::new(file_path);
    if !path.exists() {
        eprintln!("Corpus file not found, skipping: {}", file_path);
        return;
    }

    let content = fs::read_to_string(path).expect("Failed to read corpus file");
    let corpus: Value = serde_json::from_str(&content).expect("Failed to parse corpus JSON");

    let suite = corpus["suite"].as_str().unwrap_or("unknown");
    let cases = corpus["cases"].as_array().expect("cases must be an array");

    for (i, case) in cases.iter().enumerate() {
        let description = case["description"].as_str().unwrap_or("no description");
        let schema_doc = &case["schema"];
        let input = &case["input"];
        let expected_valid = case["valid"].as_bool().unwrap();
        let expected_output = &case["output"];
        let expected_issues = case["issues"].as_array();

        // Import the schema from the document
        let (schema, ctx) =
            anyvali::import_value(schema_doc).unwrap_or_else(|e| {
                panic!(
                    "Failed to import schema for suite={} case={} ({}): {}",
                    suite, i, description, e
                );
            });

        // Run safe_parse
        let result = schema.safe_parse_with_context(input, &ctx);

        // Check validity
        assert_eq!(
            result.success, expected_valid,
            "Suite={} Case={} ({}): expected valid={} but got valid={}\nIssues: {:?}",
            suite, i, description, expected_valid, result.success, result.issues
        );

        if expected_valid {
            // Check output matches
            let actual_output = result
                .value
                .as_ref()
                .expect("Expected value on success");
            assert_eq!(
                actual_output, expected_output,
                "Suite={} Case={} ({}): output mismatch.\nExpected: {}\nActual: {}",
                suite, i, description, expected_output, actual_output
            );
        } else {
            // Check that output is None
            assert!(
                result.value.is_none(),
                "Suite={} Case={} ({}): expected no value on failure",
                suite, i, description
            );

            // Check issue codes match
            if let Some(expected) = expected_issues {
                assert_eq!(
                    result.issues.len(),
                    expected.len(),
                    "Suite={} Case={} ({}): issue count mismatch.\nExpected: {:?}\nActual: {:?}",
                    suite,
                    i,
                    description,
                    expected,
                    result.issues
                );

                for (j, (actual_issue, expected_issue)) in
                    result.issues.iter().zip(expected.iter()).enumerate()
                {
                    let expected_code = expected_issue["code"].as_str().unwrap();
                    assert_eq!(
                        actual_issue.code, expected_code,
                        "Suite={} Case={} ({}) Issue={}: code mismatch. Expected={} Got={}",
                        suite, i, description, j, expected_code, actual_issue.code
                    );

                    // Check path
                    let expected_path = expected_issue["path"]
                        .as_array()
                        .expect("path must be an array");
                    assert_eq!(
                        actual_issue.path.len(),
                        expected_path.len(),
                        "Suite={} Case={} ({}) Issue={}: path length mismatch.\nExpected: {:?}\nActual: {:?}",
                        suite, i, description, j, expected_path, actual_issue.path
                    );

                    for (k, (actual_seg, expected_seg)) in
                        actual_issue.path.iter().zip(expected_path.iter()).enumerate()
                    {
                        match expected_seg {
                            Value::String(s) => {
                                assert_eq!(
                                    actual_seg,
                                    &anyvali::PathSegment::Key(s.clone()),
                                    "Suite={} Case={} ({}) Issue={} Path={}: segment mismatch",
                                    suite, i, description, j, k
                                );
                            }
                            Value::Number(n) => {
                                let idx = n.as_u64().unwrap() as usize;
                                assert_eq!(
                                    actual_seg,
                                    &anyvali::PathSegment::Index(idx),
                                    "Suite={} Case={} ({}) Issue={} Path={}: segment mismatch",
                                    suite, i, description, j, k
                                );
                            }
                            _ => panic!("Unexpected path segment type"),
                        }
                    }

                    // Check expected/received
                    if let Some(exp) = expected_issue["expected"].as_str() {
                        assert_eq!(
                            actual_issue.expected, exp,
                            "Suite={} Case={} ({}) Issue={}: expected field mismatch",
                            suite, i, description, j
                        );
                    }
                    if let Some(recv) = expected_issue["received"].as_str() {
                        assert_eq!(
                            actual_issue.received, recv,
                            "Suite={} Case={} ({}) Issue={}: received field mismatch",
                            suite, i, description, j
                        );
                    }
                }
            }
        }
    }
}

// Determine the corpus base path relative to the test file location.
// The test runs from sdk/rust/, and corpus is at ../../spec/corpus/
fn corpus_path(relative: &str) -> String {
    let base = env!("CARGO_MANIFEST_DIR");
    format!("{}/../../spec/corpus/{}", base, relative)
}

// ===== Primitives =====

#[test]
fn conformance_string_valid() {
    run_corpus_file(&corpus_path("primitives/string-valid.json"));
}

#[test]
fn conformance_string_invalid() {
    run_corpus_file(&corpus_path("primitives/string-invalid.json"));
}

#[test]
fn conformance_number_valid() {
    run_corpus_file(&corpus_path("primitives/number-valid.json"));
}

#[test]
fn conformance_number_invalid() {
    run_corpus_file(&corpus_path("primitives/number-invalid.json"));
}

#[test]
fn conformance_int_valid() {
    run_corpus_file(&corpus_path("primitives/int-valid.json"));
}

#[test]
fn conformance_int_widths() {
    run_corpus_file(&corpus_path("primitives/int-widths.json"));
}

#[test]
fn conformance_float_widths() {
    run_corpus_file(&corpus_path("primitives/float-widths.json"));
}

#[test]
fn conformance_bool() {
    run_corpus_file(&corpus_path("primitives/bool.json"));
}

#[test]
fn conformance_null() {
    run_corpus_file(&corpus_path("primitives/null.json"));
}

#[test]
fn conformance_literal() {
    run_corpus_file(&corpus_path("primitives/literal.json"));
}

#[test]
fn conformance_enum() {
    run_corpus_file(&corpus_path("primitives/enum.json"));
}

#[test]
fn conformance_any() {
    run_corpus_file(&corpus_path("primitives/any.json"));
}

#[test]
fn conformance_unknown() {
    run_corpus_file(&corpus_path("primitives/unknown.json"));
}

#[test]
fn conformance_never() {
    run_corpus_file(&corpus_path("primitives/never.json"));
}

// ===== Constraints =====

#[test]
fn conformance_string_constraints() {
    run_corpus_file(&corpus_path("constraints/string-constraints.json"));
}

#[test]
fn conformance_string_format() {
    run_corpus_file(&corpus_path("constraints/string-format.json"));
}

#[test]
fn conformance_numeric_constraints() {
    run_corpus_file(&corpus_path("constraints/numeric-constraints.json"));
}

#[test]
fn conformance_array_constraints() {
    run_corpus_file(&corpus_path("constraints/array-constraints.json"));
}

// ===== Objects =====

#[test]
fn conformance_object_required() {
    run_corpus_file(&corpus_path("objects/object-required.json"));
}

#[test]
fn conformance_object_unknown_keys() {
    run_corpus_file(&corpus_path("objects/object-unknown-keys.json"));
}

// ===== Composition =====

#[test]
fn conformance_union() {
    run_corpus_file(&corpus_path("composition/union.json"));
}

#[test]
fn conformance_intersection() {
    run_corpus_file(&corpus_path("composition/intersection.json"));
}

#[test]
fn conformance_optional() {
    run_corpus_file(&corpus_path("composition/optional.json"));
}

#[test]
fn conformance_nullable() {
    run_corpus_file(&corpus_path("composition/nullable.json"));
}

#[test]
fn conformance_array() {
    run_corpus_file(&corpus_path("composition/array.json"));
}

#[test]
fn conformance_tuple() {
    run_corpus_file(&corpus_path("composition/tuple.json"));
}

#[test]
fn conformance_record() {
    run_corpus_file(&corpus_path("composition/record.json"));
}

// ===== Defaults =====

#[test]
fn conformance_defaults() {
    run_corpus_file(&corpus_path("defaults/defaults.json"));
}

// ===== Coercions =====

#[test]
fn conformance_coercions() {
    run_corpus_file(&corpus_path("coercions/coercions.json"));
}

// ===== Refs =====

#[test]
fn conformance_refs() {
    run_corpus_file(&corpus_path("refs/refs.json"));
}

// ===== Numeric Safety =====

#[test]
fn conformance_numeric_safety() {
    run_corpus_file(&corpus_path("numeric-safety/numeric-safety.json"));
}
