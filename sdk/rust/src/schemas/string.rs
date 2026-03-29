use regex::Regex;
use serde_json::{json, Value};

use crate::format::validators::validate_format;
use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for string validation with optional constraints.
#[derive(Debug, Clone)]
pub struct StringSchema {
    pub min_length: Option<usize>,
    pub max_length: Option<usize>,
    pub pattern: Option<String>,
    pub starts_with: Option<String>,
    pub ends_with: Option<String>,
    pub includes: Option<String>,
    pub format: Option<String>,
    pub coerce: Option<Vec<String>>,
    pub default_value: Option<Value>,
}

impl StringSchema {
    pub fn new() -> Self {
        StringSchema {
            min_length: None,
            max_length: None,
            pattern: None,
            starts_with: None,
            ends_with: None,
            includes: None,
            format: None,
            coerce: None,
            default_value: None,
        }
    }

    pub fn min_length(mut self, n: usize) -> Self {
        self.min_length = Some(n);
        self
    }

    pub fn max_length(mut self, n: usize) -> Self {
        self.max_length = Some(n);
        self
    }

    pub fn pattern(mut self, p: &str) -> Self {
        self.pattern = Some(p.to_string());
        self
    }

    pub fn starts_with(mut self, s: &str) -> Self {
        self.starts_with = Some(s.to_string());
        self
    }

    pub fn ends_with(mut self, s: &str) -> Self {
        self.ends_with = Some(s.to_string());
        self
    }

    pub fn includes(mut self, s: &str) -> Self {
        self.includes = Some(s.to_string());
        self
    }

    pub fn format(mut self, f: &str) -> Self {
        self.format = Some(f.to_string());
        self
    }

    pub fn coerce(mut self, c: Vec<String>) -> Self {
        self.coerce = Some(c);
        self
    }

    pub fn default(mut self, v: Value) -> Self {
        self.default_value = Some(v);
        self
    }
}

impl Default for StringSchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for StringSchema {
    fn kind(&self) -> &str {
        "string"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        // Apply coercions
        let mut value = input.clone();
        if let Some(coercions) = &self.coerce {
            for c in coercions {
                match c.as_str() {
                    "trim" => {
                        if let Value::String(s) = &value {
                            value = Value::String(s.trim().to_string());
                        }
                    }
                    "lower" => {
                        if let Value::String(s) = &value {
                            value = Value::String(s.to_lowercase());
                        }
                    }
                    "upper" => {
                        if let Value::String(s) = &value {
                            value = Value::String(s.to_uppercase());
                        }
                    }
                    _ => {}
                }
            }
        }

        let s = match &value {
            Value::String(s) => s.clone(),
            other => {
                return Err(vec![ValidationIssue {
                    code: INVALID_TYPE.to_string(),
                    path: path.to_vec(),
                    expected: "string".to_string(),
                    received: value_type_name(other).to_string(),
                    meta: None,
                }]);
            }
        };

        let mut issues = Vec::new();

        if let Some(min) = self.min_length {
            if s.len() < min {
                issues.push(ValidationIssue {
                    code: TOO_SMALL.to_string(),
                    path: path.to_vec(),
                    expected: min.to_string(),
                    received: s.len().to_string(),
                    meta: None,
                });
            }
        }

        if let Some(max) = self.max_length {
            if s.len() > max {
                issues.push(ValidationIssue {
                    code: TOO_LARGE.to_string(),
                    path: path.to_vec(),
                    expected: max.to_string(),
                    received: s.len().to_string(),
                    meta: None,
                });
            }
        }

        if let Some(pat) = &self.pattern {
            if let Ok(re) = Regex::new(pat) {
                if !re.is_match(&s) {
                    issues.push(ValidationIssue {
                        code: INVALID_STRING.to_string(),
                        path: path.to_vec(),
                        expected: pat.clone(),
                        received: s.clone(),
                        meta: None,
                    });
                }
            }
        }

        if let Some(prefix) = &self.starts_with {
            if !s.starts_with(prefix.as_str()) {
                issues.push(ValidationIssue {
                    code: INVALID_STRING.to_string(),
                    path: path.to_vec(),
                    expected: prefix.clone(),
                    received: s.clone(),
                    meta: None,
                });
            }
        }

        if let Some(suffix) = &self.ends_with {
            if !s.ends_with(suffix.as_str()) {
                issues.push(ValidationIssue {
                    code: INVALID_STRING.to_string(),
                    path: path.to_vec(),
                    expected: suffix.clone(),
                    received: s.clone(),
                    meta: None,
                });
            }
        }

        if let Some(sub) = &self.includes {
            if !s.contains(sub.as_str()) {
                issues.push(ValidationIssue {
                    code: INVALID_STRING.to_string(),
                    path: path.to_vec(),
                    expected: sub.clone(),
                    received: s.clone(),
                    meta: None,
                });
            }
        }

        if let Some(fmt) = &self.format {
            if !validate_format(fmt, &s) {
                issues.push(ValidationIssue {
                    code: INVALID_STRING.to_string(),
                    path: path.to_vec(),
                    expected: fmt.clone(),
                    received: s.clone(),
                    meta: None,
                });
            }
        }

        if issues.is_empty() {
            Ok(value)
        } else {
            Err(issues)
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({"kind": "string"});
        let obj = node.as_object_mut().unwrap();
        if let Some(v) = self.min_length {
            obj.insert("minLength".to_string(), json!(v));
        }
        if let Some(v) = self.max_length {
            obj.insert("maxLength".to_string(), json!(v));
        }
        if let Some(v) = &self.pattern {
            obj.insert("pattern".to_string(), json!(v));
        }
        if let Some(v) = &self.starts_with {
            obj.insert("startsWith".to_string(), json!(v));
        }
        if let Some(v) = &self.ends_with {
            obj.insert("endsWith".to_string(), json!(v));
        }
        if let Some(v) = &self.includes {
            obj.insert("includes".to_string(), json!(v));
        }
        if let Some(v) = &self.format {
            obj.insert("format".to_string(), json!(v));
        }
        if let Some(v) = &self.coerce {
            if v.len() == 1 {
                obj.insert("coerce".to_string(), json!(v[0]));
            } else {
                obj.insert("coerce".to_string(), json!(v));
            }
        }
        if let Some(v) = &self.default_value {
            obj.insert("default".to_string(), v.clone());
        }
        node
    }
}
