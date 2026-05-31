use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{
    ParseContext, Schema, DescribeOpts, add_metadata_to_node, build_describe_metadata,
    merge_metadata, reserved_metadata_keys, validate_metadata_keys,
};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for tuple (fixed-length array with typed elements) validation.
#[derive(Debug, Clone)]
pub struct TupleSchema {
    pub elements: Vec<Box<dyn Schema>>,
    pub metadata: Option<Value>,
}

impl TupleSchema {
    pub fn new(elements: Vec<Box<dyn Schema>>) -> Self {
        TupleSchema {
            elements,
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

impl Schema for TupleSchema {
    fn kind(&self) -> &str {
        "tuple"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        let arr = match input {
            Value::Array(a) => a,
            other => {
                return Err(vec![ValidationIssue {
                    code: INVALID_TYPE.to_string(),
                    path: path.to_vec(),
                    expected: "tuple".to_string(),
                    received: value_type_name(other).to_string(),
                    meta: None,
                }]);
            }
        };

        let expected_len = self.elements.len();
        if arr.len() < expected_len {
            return Err(vec![ValidationIssue {
                code: TOO_SMALL.to_string(),
                path: path.to_vec(),
                expected: expected_len.to_string(),
                received: arr.len().to_string(),
                meta: None,
            }]);
        }
        if arr.len() > expected_len {
            return Err(vec![ValidationIssue {
                code: TOO_LARGE.to_string(),
                path: path.to_vec(),
                expected: expected_len.to_string(),
                received: arr.len().to_string(),
                meta: None,
            }]);
        }

        let mut issues = Vec::new();
        let mut result = Vec::new();

        for (i, (elem, schema)) in arr.iter().zip(self.elements.iter()).enumerate() {
            let mut elem_path = path.to_vec();
            elem_path.push(PathSegment::Index(i));
            match schema.parse_value(elem, &elem_path, ctx) {
                Ok(v) => result.push(v),
                Err(mut errs) => issues.append(&mut errs),
            }
        }

        if issues.is_empty() {
            Ok(Value::Array(result))
        } else {
            Err(issues)
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "tuple",
            "elements": self.elements.iter().map(|e| e.export_node()).collect::<Vec<_>>()
        });
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
