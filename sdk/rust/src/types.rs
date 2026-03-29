use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::fmt;

/// A single validation issue with path, code, expected/received info.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ValidationIssue {
    pub code: String,
    pub path: Vec<PathSegment>,
    pub expected: String,
    pub received: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub meta: Option<Value>,
}

/// A segment in a validation path - either a string key or an integer index.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(untagged)]
pub enum PathSegment {
    Key(String),
    Index(usize),
}

impl fmt::Display for PathSegment {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            PathSegment::Key(k) => write!(f, "{}", k),
            PathSegment::Index(i) => write!(f, "{}", i),
        }
    }
}

/// Error returned by parse() on validation failure.
#[derive(Debug, Clone)]
pub struct ValidationError {
    pub issues: Vec<ValidationIssue>,
}

impl fmt::Display for ValidationError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Validation failed with {} issue(s)", self.issues.len())
    }
}

impl std::error::Error for ValidationError {}

/// Result of safe_parse: either success with a value or failure with issues.
#[derive(Debug, Clone)]
pub struct ParseResult {
    pub success: bool,
    pub value: Option<Value>,
    pub issues: Vec<ValidationIssue>,
}

impl ParseResult {
    pub fn ok(value: Value) -> Self {
        ParseResult {
            success: true,
            value: Some(value),
            issues: vec![],
        }
    }

    pub fn err(issues: Vec<ValidationIssue>) -> Self {
        ParseResult {
            success: false,
            value: None,
            issues,
        }
    }
}

/// The canonical AnyVali JSON interchange document.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AnyValiDocument {
    pub anyvali_version: String,
    pub schema_version: String,
    pub root: Value,
    #[serde(default)]
    pub definitions: serde_json::Map<String, Value>,
    #[serde(default)]
    pub extensions: serde_json::Map<String, Value>,
}

/// Export mode for schemas.
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum ExportMode {
    Portable,
    Extended,
}

/// Unknown keys handling mode for objects.
#[derive(Debug, Clone, Copy, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum UnknownKeyMode {
    Reject,
    Strip,
    Allow,
}

impl Default for UnknownKeyMode {
    fn default() -> Self {
        UnknownKeyMode::Reject
    }
}

/// Get the JSON type name of a serde_json::Value.
pub fn value_type_name(v: &Value) -> &'static str {
    match v {
        Value::Null => "null",
        Value::Bool(_) => "boolean",
        Value::Number(_) => "number",
        Value::String(_) => "string",
        Value::Array(_) => "array",
        Value::Object(_) => "object",
    }
}
