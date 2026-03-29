use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema that rejects all values.
#[derive(Debug, Clone)]
pub struct NeverSchema;

impl NeverSchema {
    pub fn new() -> Self {
        NeverSchema
    }
}

impl Default for NeverSchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for NeverSchema {
    fn kind(&self) -> &str {
        "never"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        Err(vec![ValidationIssue {
            code: INVALID_TYPE.to_string(),
            path: path.to_vec(),
            expected: "never".to_string(),
            received: value_type_name(input).to_string(),
            meta: None,
        }])
    }

    fn export_node(&self) -> Value {
        json!({"kind": "never"})
    }
}
