use serde_json::{json, Value};

use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that accepts any value (like any, but semantically different in type systems).
#[derive(Debug, Clone)]
pub struct UnknownSchema;

impl UnknownSchema {
    pub fn new() -> Self {
        UnknownSchema
    }
}

impl Default for UnknownSchema {
    fn default() -> Self {
        Self::new()
    }
}

impl Schema for UnknownSchema {
    fn kind(&self) -> &str {
        "unknown"
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
        json!({"kind": "unknown"})
    }
}
