use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for tuple (fixed-length array with typed elements) validation.
#[derive(Debug, Clone)]
pub struct TupleSchema {
    pub elements: Vec<Box<dyn Schema>>,
}

impl TupleSchema {
    pub fn new(elements: Vec<Box<dyn Schema>>) -> Self {
        TupleSchema { elements }
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
        json!({
            "kind": "tuple",
            "elements": self.elements.iter().map(|e| e.export_node()).collect::<Vec<_>>()
        })
    }
}
