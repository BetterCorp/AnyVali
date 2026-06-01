use serde_json::{json, Value};

use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue};

/// Schema for intersection (all schemas must pass) validation.
#[derive(Debug, Clone)]
pub struct IntersectionSchema {
    pub all_of: Vec<Box<dyn Schema>>,
    pub metadata: Option<Value>,
}

impl IntersectionSchema {
    pub fn new(all_of: Vec<Box<dyn Schema>>) -> Self {
        IntersectionSchema {
            all_of,
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

impl Schema for IntersectionSchema {
    fn kind(&self) -> &str {
        "intersection"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        let mut all_issues = Vec::new();
        let mut last_result = input.clone();

        for schema in &self.all_of {
            match schema.parse_value(input, path, ctx) {
                Ok(v) => {
                    // For objects, merge results
                    if let (Value::Object(existing), Value::Object(new)) =
                        (&mut last_result, &v)
                    {
                        for (k, val) in new {
                            existing.insert(k.clone(), val.clone());
                        }
                    } else {
                        last_result = v;
                    }
                }
                Err(mut errs) => all_issues.append(&mut errs),
            }
        }

        if all_issues.is_empty() {
            Ok(last_result)
        } else {
            Err(all_issues)
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "intersection",
            "allOf": self.all_of.iter().map(|s| s.export_node()).collect::<Vec<_>>()
        });
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
