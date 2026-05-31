use serde_json::{json, Value};

use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that accepts any value (like any, but semantically different in type systems).
#[derive(Debug, Clone)]
pub struct UnknownSchema {
    pub metadata: Option<Value>,
}

impl UnknownSchema {
    pub fn new() -> Self {
        UnknownSchema { metadata: None }
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

impl Default for UnknownSchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for UnknownSchema {
    fn kind(&self) -> &str {
        "unknown"
    }

    fn parse_value(
        &self,
        input: &Value,
        _path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        Ok(input.clone())
    }

    fn export_node(&self) -> Value {
        let mut node = json!({"kind": "unknown"});
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
