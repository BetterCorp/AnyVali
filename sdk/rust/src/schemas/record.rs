use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for record (string-keyed map with uniform value type) validation.
#[derive(Debug, Clone)]
pub struct RecordSchema {
    pub values: Box<dyn Schema>,
    pub metadata: Option<Value>,
}

impl RecordSchema {
    pub fn new(values: Box<dyn Schema>) -> Self {
        RecordSchema {
            values,
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

impl Schema for RecordSchema {
    fn kind(&self) -> &str {
        "record"
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
                    expected: "record".to_string(),
                    received: value_type_name(other).to_string(),
                    meta: None,
                }]);
            }
        };

        let mut issues = Vec::new();
        let mut result = serde_json::Map::new();

        for (key, value) in obj {
            let mut key_path = path.to_vec();
            key_path.push(PathSegment::Key(key.clone()));
            match self.values.parse_value(value, &key_path, ctx) {
                Ok(v) => {
                    result.insert(key.clone(), v);
                }
                Err(mut errs) => issues.append(&mut errs),
            }
        }

        if issues.is_empty() {
            Ok(Value::Object(result))
        } else {
            Err(issues)
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "record",
            "values": self.values.export_node()
        });
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
