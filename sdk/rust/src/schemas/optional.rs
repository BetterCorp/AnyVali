use serde_json::{json, Value};

use crate::schema::{
    add_metadata_to_node, build_describe_metadata, merge_metadata, reserved_metadata_keys,
    validate_metadata_keys, DescribeOpts, ParseContext, Schema,
};
use crate::types::{PathSegment, ValidationIssue};

/// Schema wrapper that makes a field optional (absent is allowed, but if present must validate).
/// Note: optional is only meaningful inside object schemas. At the top level,
/// it delegates to its inner schema.
#[derive(Debug, Clone)]
pub struct OptionalSchema {
    pub schema: Box<dyn Schema>,
    pub default_value: Option<Value>,
    pub metadata: Option<Value>,
}

impl OptionalSchema {
    pub fn new(schema: Box<dyn Schema>) -> Self {
        OptionalSchema {
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

impl Schema for OptionalSchema {
    fn kind(&self) -> &str {
        "optional"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        // When used at field level, absent is handled by the object schema.
        // If we get here, the value is present and must validate.
        self.schema.parse_value(input, path, ctx)
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "optional",
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

    fn default_value(&self) -> Option<&Value> {
        self.default_value.as_ref()
    }
}
