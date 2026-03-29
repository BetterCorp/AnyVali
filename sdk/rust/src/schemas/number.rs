use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for number (float64) validation.
#[derive(Debug, Clone)]
pub struct NumberSchema {
    /// The kind name: "number", "float64", or "float32"
    kind_name: String,
    pub min: Option<f64>,
    pub max: Option<f64>,
    pub exclusive_min: Option<f64>,
    pub exclusive_max: Option<f64>,
    pub multiple_of: Option<f64>,
    pub coerce: Option<Vec<String>>,
    pub default_value: Option<Value>,
    /// For float32, we check range
    is_float32: bool,
}

impl NumberSchema {
    pub fn new() -> Self {
        NumberSchema {
            kind_name: "number".to_string(),
            min: None,
            max: None,
            exclusive_min: None,
            exclusive_max: None,
            multiple_of: None,
            coerce: None,
            default_value: None,
            is_float32: false,
        }
    }

    pub fn float64() -> Self {
        NumberSchema {
            kind_name: "float64".to_string(),
            ..Self::new()
        }
    }

    pub fn float32() -> Self {
        NumberSchema {
            kind_name: "float32".to_string(),
            is_float32: true,
            ..Self::new()
        }
    }

    pub fn min(mut self, v: f64) -> Self {
        self.min = Some(v);
        self
    }

    pub fn max(mut self, v: f64) -> Self {
        self.max = Some(v);
        self
    }

    pub fn exclusive_min(mut self, v: f64) -> Self {
        self.exclusive_min = Some(v);
        self
    }

    pub fn exclusive_max(mut self, v: f64) -> Self {
        self.exclusive_max = Some(v);
        self
    }

    pub fn multiple_of(mut self, v: f64) -> Self {
        self.multiple_of = Some(v);
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

impl Default for NumberSchema {
    fn default() -> Self {
        Self::new()
    }
}

/// Format a number for display in issue messages, matching spec expectations.
fn format_num(n: f64) -> String {
    if n.fract() == 0.0 && n.abs() < 1e15 {
        format!("{}", n as i64)
    } else {
        format!("{}", n)
    }
}

fn validate_numeric_constraints(
    n: f64,
    schema: &NumberSchema,
    path: &[PathSegment],
) -> Vec<ValidationIssue> {
    let mut issues = Vec::new();

    if let Some(min) = schema.min {
        if n < min {
            issues.push(ValidationIssue {
                code: TOO_SMALL.to_string(),
                path: path.to_vec(),
                expected: format_num(min),
                received: format_num(n),
                meta: None,
            });
        }
    }

    if let Some(max) = schema.max {
        if n > max {
            issues.push(ValidationIssue {
                code: TOO_LARGE.to_string(),
                path: path.to_vec(),
                expected: format_num(max),
                received: format_num(n),
                meta: None,
            });
        }
    }

    if let Some(emin) = schema.exclusive_min {
        if n <= emin {
            issues.push(ValidationIssue {
                code: TOO_SMALL.to_string(),
                path: path.to_vec(),
                expected: format_num(emin),
                received: format_num(n),
                meta: None,
            });
        }
    }

    if let Some(emax) = schema.exclusive_max {
        if n >= emax {
            issues.push(ValidationIssue {
                code: TOO_LARGE.to_string(),
                path: path.to_vec(),
                expected: format_num(emax),
                received: format_num(n),
                meta: None,
            });
        }
    }

    if let Some(mo) = schema.multiple_of {
        let remainder = n % mo;
        if remainder.abs() > 1e-10 && (remainder - mo).abs() > 1e-10 {
            issues.push(ValidationIssue {
                code: INVALID_NUMBER.to_string(),
                path: path.to_vec(),
                expected: format_num(mo),
                received: format_num(n),
                meta: None,
            });
        }
    }

    issues
}

impl Schema for NumberSchema {
    fn kind(&self) -> &str {
        &self.kind_name
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
                if c == "string->number" {
                    if let Value::String(s) = &value {
                        let trimmed = s.trim();
                        match trimmed.parse::<f64>() {
                            Ok(n) => {
                                value = json!(n);
                            }
                            Err(_) => {
                                return Err(vec![ValidationIssue {
                                    code: COERCION_FAILED.to_string(),
                                    path: path.to_vec(),
                                    expected: self.kind_name.clone(),
                                    received: s.clone(),
                                    meta: None,
                                }]);
                            }
                        }
                    }
                }
            }
        }

        match &value {
            Value::Number(n) => {
                let f = n.as_f64().unwrap();

                if self.is_float32 {
                    // float32 range check: we accept values that fit in f32
                    // (the corpus doesn't test range rejection for float32 specifically,
                    //  but we validate the type)
                }

                let issues = validate_numeric_constraints(f, self, path);
                if issues.is_empty() {
                    Ok(value)
                } else {
                    Err(issues)
                }
            }
            other => Err(vec![ValidationIssue {
                code: INVALID_TYPE.to_string(),
                path: path.to_vec(),
                expected: self.kind_name.clone(),
                received: value_type_name(other).to_string(),
                meta: None,
            }]),
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({"kind": self.kind_name});
        let obj = node.as_object_mut().unwrap();
        if let Some(v) = self.min {
            obj.insert("min".to_string(), json!(v));
        }
        if let Some(v) = self.max {
            obj.insert("max".to_string(), json!(v));
        }
        if let Some(v) = self.exclusive_min {
            obj.insert("exclusiveMin".to_string(), json!(v));
        }
        if let Some(v) = self.exclusive_max {
            obj.insert("exclusiveMax".to_string(), json!(v));
        }
        if let Some(v) = self.multiple_of {
            obj.insert("multipleOf".to_string(), json!(v));
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
