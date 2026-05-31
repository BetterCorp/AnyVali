use serde_json::{json, Value};

use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue};

/// Schema wrapper that allows null in addition to the inner schema's type.
#[derive(Debug, Clone)]
pub struct NullableSchema {
    pub schema: Box<dyn Schema>,
    pub default_value: Option<Value>,
    pub metadata: Option<Value>,
}

impl NullableSchema {
    pub fn new(schema: Box<dyn Schema>) -> Self {
        NullableSchema {
            schema,
            default_value: None,
            metadata: None,
        }
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

impl Schema for NullableSchema {
    fn kind(&self) -> &str {
        "nullable"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        if input.is_null() {
            return Ok(Value::Null);
        }
        self.schema.parse_value(input, path, ctx)
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "nullable",
            "schema": self.schema.export_node()
        });
        if let Some(v) = &self.default_value {
            node.as_object_mut()
                .unwrap()
                .insert("default".to_string(), v.clone());
        }
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
