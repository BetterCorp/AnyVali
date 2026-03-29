use serde_json::Value;

use crate::issue_codes::*;
use crate::types::ValidationIssue;

/// Apply a coercion to a value. Returns the coerced value or a coercion_failed issue.
pub fn apply_coercion(
    value: &Value,
    coercion: &str,
    target_kind: &str,
) -> Result<Value, ValidationIssue> {
    match coercion {
        "string->int" => {
            if let Value::String(s) = value {
                let trimmed = s.trim();
                match trimmed.parse::<i64>() {
                    Ok(n) => Ok(serde_json::json!(n)),
                    Err(_) => Err(ValidationIssue {
                        code: COERCION_FAILED.to_string(),
                        path: vec![],
                        expected: target_kind.to_string(),
                        received: s.clone(),
                        meta: None,
                    }),
                }
            } else {
                Ok(value.clone())
            }
        }
        "string->number" => {
            if let Value::String(s) = value {
                let trimmed = s.trim();
                match trimmed.parse::<f64>() {
                    Ok(n) => Ok(serde_json::json!(n)),
                    Err(_) => Err(ValidationIssue {
                        code: COERCION_FAILED.to_string(),
                        path: vec![],
                        expected: target_kind.to_string(),
                        received: s.clone(),
                        meta: None,
                    }),
                }
            } else {
                Ok(value.clone())
            }
        }
        "string->bool" => {
            if let Value::String(s) = value {
                let lower = s.trim().to_lowercase();
                match lower.as_str() {
                    "true" | "1" => Ok(Value::Bool(true)),
                    "false" | "0" => Ok(Value::Bool(false)),
                    _ => Err(ValidationIssue {
                        code: COERCION_FAILED.to_string(),
                        path: vec![],
                        expected: target_kind.to_string(),
                        received: s.clone(),
                        meta: None,
                    }),
                }
            } else {
                Ok(value.clone())
            }
        }
        "trim" => {
            if let Value::String(s) = value {
                Ok(Value::String(s.trim().to_string()))
            } else {
                Ok(value.clone())
            }
        }
        "lower" => {
            if let Value::String(s) = value {
                Ok(Value::String(s.to_lowercase()))
            } else {
                Ok(value.clone())
            }
        }
        "upper" => {
            if let Value::String(s) = value {
                Ok(Value::String(s.to_uppercase()))
            } else {
                Ok(value.clone())
            }
        }
        _ => Ok(value.clone()),
    }
}

/// Parse coercion config from a JSON value (string or array of strings).
pub fn parse_coerce_config(v: &Value) -> Vec<String> {
    match v {
        Value::String(s) => vec![s.clone()],
        Value::Array(arr) => arr
            .iter()
            .filter_map(|v| v.as_str().map(|s| s.to_string()))
            .collect(),
        _ => vec![],
    }
}
