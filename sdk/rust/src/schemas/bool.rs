use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for boolean validation.
#[derive(Debug, Clone)]
pub struct BoolSchema {
    pub coerce: Option<Vec<String>>,
    pub default_value: Option<Value>,
    pub metadata: Option<Value>,
}

impl BoolSchema {
    pub fn new() -> Self {
        BoolSchema {
            coerce: None,
            default_value: None,
            metadata: None,
        }
    }

    pub fn coerce(mut self, c: Vec<String>) -> Self {
        self.coerce = Some(c);
        self
    }

    /// Enable coercion with the target inferred from the schema kind. A bare
    /// string input is coerced as `string->bool`. Equivalent to `.coerce(vec![])`,
    /// reading as "coerce from string into this kind".
    pub fn coerce_default(self) -> Self {
        self.coerce(vec![])
    }

    pub fn default(mut self, v: Value) -> Self {
        self.default_value = Some(v);
        self
    }

    pub fn describe(mut self, description: &str, opts: Option<&DescribeOpts>) -> Self {
        let new_meta = build_describe_metadata(description, opts);
        match &mut self.metadata {
            Some(existing) => merge_metadata(existing, &new_meta),
            None => self.metadata = Some(new_meta),
        }
        self
    }

    pub fn with_metadata(mut self, meta: Value, replace: bool) -> Self {
        validate_metadata_keys(&meta);
        if replace {
            let reserved = reserved_metadata_keys();
            let mut preserved = serde_json::Map::new();
            if let Some(Value::Object(existing)) = &self.metadata {
                for (k, v) in existing {
                    if reserved.contains(k.as_str()) {
                        preserved.insert(k.clone(), v.clone());
                    }
                }
            }
            if let Value::Object(new_map) = &meta {
                for (k, v) in new_map {
                    preserved.insert(k.clone(), v.clone());
                }
            }
            self.metadata = Some(Value::Object(preserved));
        } else {
            match &mut self.metadata {
                Some(existing) => merge_metadata(existing, &meta),
                None => self.metadata = Some(meta),
            }
        }
        self
    }
}

impl Default for BoolSchema {
    fn default() -> Self {
        Self::new()
    }
}

/// Coerce a string value into a bool (trim + case-insensitive; true<-"true"/"1",
/// false<-"false"/"0"). Non-string values pass through unchanged.
fn coerce_string_to_bool(
    value: &Value,
    path: &[PathSegment],
) -> Result<Value, Vec<ValidationIssue>> {
    if let Value::String(s) = value {
        let lower = s.trim().to_lowercase();
        match lower.as_str() {
            "true" | "1" => Ok(Value::Bool(true)),
            "false" | "0" => Ok(Value::Bool(false)),
            _ => Err(vec![ValidationIssue {
                code: COERCION_FAILED.to_string(),
                path: path.to_vec(),
                expected: "bool".to_string(),
                received: s.clone(),
                meta: None,
            }]),
        }
    } else {
        Ok(value.clone())
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
        // Apply coercions. When coercion is enabled but no explicit typed
        // source token is supplied (e.g. `.coerce(vec![])` / `.coerce_default()`),
        // the only portable source is "string", so a bare string input is
        // coerced as `string->bool` rather than silently no-oping.
        let mut value = input.clone();
        if let Some(coercions) = &self.coerce {
            let has_string_bool = coercions.iter().any(|c| c == "string->bool");
            let infer_default = coercions.is_empty();
            for c in coercions {
                if c != "string->bool" {
                    continue;
                }
                value = coerce_string_to_bool(&value, path)?;
            }
            if infer_default && !has_string_bool {
                value = coerce_string_to_bool(&value, path)?;
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
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
