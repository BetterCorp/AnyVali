use serde_json::{json, Value};
use std::collections::BTreeMap;

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, UnknownKeyMode, ValidationIssue, value_type_name};

/// A field definition in an object schema.
#[derive(Debug, Clone)]
pub struct ObjectField {
    pub schema: Box<dyn Schema>,
    pub default_value: Option<Value>,
}

/// Schema for object validation.
#[derive(Debug, Clone)]
pub struct ObjectSchema {
    pub properties: BTreeMap<String, ObjectField>,
    pub required: Vec<String>,
    pub unknown_keys: UnknownKeyMode,
}

impl ObjectSchema {
    pub fn new() -> Self {
        ObjectSchema {
            properties: BTreeMap::new(),
            required: Vec::new(),
            unknown_keys: UnknownKeyMode::Reject,
        }
    }

    pub fn field(mut self, name: &str, schema: Box<dyn Schema>) -> Self {
        self.properties.insert(
            name.to_string(),
            ObjectField {
                schema,
                default_value: None,
            },
        );
        self
    }

    pub fn field_with_default(
        mut self,
        name: &str,
        schema: Box<dyn Schema>,
        default: Value,
    ) -> Self {
        self.properties.insert(
            name.to_string(),
            ObjectField {
                schema,
                default_value: Some(default),
            },
        );
        self
    }

    pub fn required(mut self, fields: Vec<&str>) -> Self {
        self.required = fields.iter().map(|s| s.to_string()).collect();
        self
    }

    pub fn unknown_keys(mut self, mode: UnknownKeyMode) -> Self {
        self.unknown_keys = mode;
        self
    }
}

impl Default for ObjectSchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for ObjectSchema {
    fn kind(&self) -> &str {
        "object"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        let obj = match input {
            Value::Object(o) => o,
            other => {
                return Err(vec![ValidationIssue {
                    code: INVALID_TYPE.to_string(),
                    path: path.to_vec(),
                    expected: "object".to_string(),
                    received: value_type_name(other).to_string(),
                    meta: None,
                }]);
            }
        };

        let mut issues = Vec::new();
        let mut result = serde_json::Map::new();

        // Check required fields and validate all properties
        for (field_name, field_def) in &self.properties {
            let mut field_path = path.to_vec();
            field_path.push(PathSegment::Key(field_name.clone()));

            if let Some(value) = obj.get(field_name) {
                // Field is present - validate it
                match field_def.schema.parse_value(value, &field_path, ctx) {
                    Ok(v) => {
                        result.insert(field_name.clone(), v);
                    }
                    Err(mut errs) => issues.append(&mut errs),
                }
            } else if let Some(default) = &field_def.default_value {
                // Field absent, has default - apply default then validate
                match field_def.schema.parse_value(default, &field_path, ctx) {
                    Ok(v) => {
                        result.insert(field_name.clone(), v);
                    }
                    Err(_) => {
                        // Default is invalid
                        issues.push(ValidationIssue {
                            code: crate::issue_codes::DEFAULT_INVALID.to_string(),
                            path: field_path,
                            expected: format_default_expected(&field_def.schema, default),
                            received: format_default_received(default),
                            meta: None,
                        });
                    }
                }
            } else if self.required.contains(field_name) {
                // Required but missing
                issues.push(ValidationIssue {
                    code: REQUIRED.to_string(),
                    path: field_path,
                    expected: field_def.schema.kind().to_string(),
                    received: "undefined".to_string(),
                    meta: None,
                });
            }
            // else: optional and absent, skip
        }

        // Check unknown keys
        for key in obj.keys() {
            if !self.properties.contains_key(key) {
                match self.unknown_keys {
                    UnknownKeyMode::Reject => {
                        let mut key_path = path.to_vec();
                        key_path.push(PathSegment::Key(key.clone()));
                        issues.push(ValidationIssue {
                            code: UNKNOWN_KEY.to_string(),
                            path: key_path,
                            expected: "undefined".to_string(),
                            received: key.clone(),
                            meta: None,
                        });
                    }
                    UnknownKeyMode::Strip => {
                        // silently drop
                    }
                    UnknownKeyMode::Allow => {
                        result.insert(key.clone(), obj[key].clone());
                    }
                }
            }
        }

        if issues.is_empty() {
            Ok(Value::Object(result))
        } else {
            Err(issues)
        }
    }

    fn export_node(&self) -> Value {
        let mut props = json!({});
        for (name, field) in &self.properties {
            let mut node = field.schema.export_node();
            if let Some(default) = &field.default_value {
                node.as_object_mut()
                    .unwrap()
                    .insert("default".to_string(), default.clone());
            }
            props
                .as_object_mut()
                .unwrap()
                .insert(name.clone(), node);
        }

        let mut node = json!({
            "kind": "object",
            "properties": props,
            "required": self.required,
        });

        let mode_str = match self.unknown_keys {
            UnknownKeyMode::Reject => "reject",
            UnknownKeyMode::Strip => "strip",
            UnknownKeyMode::Allow => "allow",
        };
        node.as_object_mut()
            .unwrap()
            .insert("unknownKeys".to_string(), json!(mode_str));

        node
    }
}

/// Extract the constraint value from a schema for default_invalid reporting.
fn format_default_expected(schema: &Box<dyn Schema>, _default: &Value) -> String {
    // Look at the exported node to find the first constraint
    let node = schema.export_node();
    if let Some(obj) = node.as_object() {
        // Check min, max, minLength, maxLength
        for key in &["min", "max", "minLength", "maxLength"] {
            if let Some(v) = obj.get(*key) {
                return format_constraint_value(v);
            }
        }
    }
    schema.kind().to_string()
}

fn format_default_received(default: &Value) -> String {
    match default {
        Value::Number(n) => {
            let f = n.as_f64().unwrap();
            if f.fract() == 0.0 {
                format!("{}", f as i64)
            } else {
                format!("{}", f)
            }
        }
        Value::String(s) => s.clone(),
        Value::Bool(b) => b.to_string(),
        Value::Null => "null".to_string(),
        other => other.to_string(),
    }
}

fn format_constraint_value(v: &Value) -> String {
    match v {
        Value::Number(n) => {
            let f = n.as_f64().unwrap();
            if f.fract() == 0.0 {
                format!("{}", f as i64)
            } else {
                format!("{}", f)
            }
        }
        other => other.to_string(),
    }
}
