use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for union (any of several variants) validation.
#[derive(Debug, Clone)]
pub struct UnionSchema {
    pub variants: Vec<Box<dyn Schema>>,
    pub metadata: Option<Value>,
}

impl UnionSchema {
    pub fn new(variants: Vec<Box<dyn Schema>>) -> Self {
        UnionSchema {
            variants,
            metadata: None,
        }
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

impl Schema for UnionSchema {
    fn kind(&self) -> &str {
        "union"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        // Try each variant, return first success
        for variant in &self.variants {
            if let Ok(v) = variant.parse_value(input, path, ctx) {
                return Ok(v);
            }
        }

        // All variants failed
        let variant_names = self
            .variants
            .iter()
            .map(|v| v.kind().to_string())
            .collect::<Vec<_>>()
            .join(" | ");

        Err(vec![ValidationIssue {
            code: INVALID_UNION.to_string(),
            path: path.to_vec(),
            expected: variant_names,
            received: value_type_name(input).to_string(),
            meta: None,
        }])
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "union",
            "variants": self.variants.iter().map(|v| v.export_node()).collect::<Vec<_>>()
        });
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
