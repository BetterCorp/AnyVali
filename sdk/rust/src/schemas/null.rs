use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema that only accepts null.
#[derive(Debug, Clone)]
pub struct NullSchema;

impl NullSchema {
    pub fn new() -> Self {
        NullSchema
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
        json!({"kind": "null"})
    }
}
