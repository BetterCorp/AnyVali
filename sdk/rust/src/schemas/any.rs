use serde_json::{json, Value};

use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that accepts any value.
#[derive(Debug, Clone)]
pub struct AnySchema;

impl AnySchema {
    pub fn new() -> Self {
        AnySchema
    }
}

impl Default for AnySchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for AnySchema {
    fn kind(&self) -> &str {
        "any"
    }

    fn parse_value(
        &self,
        input: &Value,
        _path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        Ok(input.clone())
    }

    fn export_node(&self) -> Value {
        json!({"kind": "any"})
    }
}
