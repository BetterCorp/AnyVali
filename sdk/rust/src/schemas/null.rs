use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema that only accepts null.
#[derive(Debug, Clone)]
pub struct NullSchema {
    pub metadata: Option<Value>,
}

impl NullSchema {
    pub fn new() -> Self {
        NullSchema { metadata: None }
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

impl Default for NullSchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for NullSchema {
    fn kind(&self) -> &str {
        "null"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        match input {
            Value::Null => Ok(Value::Null),
            other => Err(vec![ValidationIssue {
                code: INVALID_TYPE.to_string(),
                path: path.to_vec(),
                expected: "null".to_string(),
                received: value_type_name(other).to_string(),
                meta: None,
            }]),
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({"kind": "null"});
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
