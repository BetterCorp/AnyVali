use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that matches one of a set of allowed values.
#[derive(Debug, Clone)]
pub struct EnumSchema {
    pub values: Vec<Value>,
    pub metadata: Option<Value>,
}

impl EnumSchema {
    pub fn new(values: Vec<Value>) -> Self {
        EnumSchema {
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

fn display_value(v: &Value) -> String {
    match v {
        Value::String(s) => s.clone(),
        Value::Number(n) => n.to_string(),
        Value::Bool(b) => b.to_string(),
        Value::Null => "null".to_string(),
        other => other.to_string(),
    }
}

impl Schema for EnumSchema {
    fn kind(&self) -> &str {
        "enum"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        for v in &self.values {
            if input == v {
                return Ok(input.clone());
            }
        }

        let values_str = self
            .values
            .iter()
            .map(|v| display_value(v))
            .collect::<Vec<_>>()
            .join(",");

        Err(vec![ValidationIssue {
            code: INVALID_TYPE.to_string(),
            path: path.to_vec(),
            expected: format!("enum({})", values_str),
            received: display_value(input),
            meta: None,
        }])
    }

    fn export_node(&self) -> Value {
        let mut node = json!({"kind": "enum", "values": self.values});
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
