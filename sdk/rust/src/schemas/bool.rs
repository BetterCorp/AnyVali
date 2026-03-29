use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for boolean validation.
#[derive(Debug, Clone)]
pub struct BoolSchema {
    pub coerce: Option<Vec<String>>,
    pub default_value: Option<Value>,
}

impl BoolSchema {
    pub fn new() -> Self {
        BoolSchema {
            coerce: None,
            default_value: None,
        }
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

impl Default for BoolSchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for BoolSchema {
    fn kind(&self) -> &str {
        "bool"
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
                if c == "string->bool" {
                    if let Value::String(s) = &value {
                        let lower = s.trim().to_lowercase();
                        match lower.as_str() {
                            "true" | "1" => {
                                value = Value::Bool(true);
                            }
                            "false" | "0" => {
                                value = Value::Bool(false);
                            }
                            _ => {
                                return Err(vec![ValidationIssue {
                                    code: COERCION_FAILED.to_string(),
                                    path: path.to_vec(),
                                    expected: "bool".to_string(),
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
            Value::Bool(_) => Ok(value),
            other => Err(vec![ValidationIssue {
                code: INVALID_TYPE.to_string(),
                path: path.to_vec(),
                expected: "bool".to_string(),
                received: value_type_name(other).to_string(),
                meta: None,
            }]),
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({"kind": "bool"});
        if let Some(v) = &self.coerce {
            let obj = node.as_object_mut().unwrap();
            if v.len() == 1 {
                obj.insert("coerce".to_string(), json!(v[0]));
            } else {
                obj.insert("coerce".to_string(), json!(v));
            }
        }
        if let Some(v) = &self.default_value {
            node.as_object_mut()
                .unwrap()
                .insert("default".to_string(), v.clone());
        }
        node
    }
}
