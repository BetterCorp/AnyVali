use serde_json::Value;

/// Apply a default value if the input is absent (None/undefined).
/// Returns the default value if input is None, otherwise the input.
pub fn apply_default(input: Option<&Value>, default: &Value) -> Value {
    match input {
        Some(v) => v.clone(),
        None => default.clone(),
    }
}

/// Check if a value should be treated as absent for default application.
/// Note: null is NOT treated as absent per the spec.
pub fn is_absent(input: Option<&Value>) -> bool {
    input.is_none()
}
