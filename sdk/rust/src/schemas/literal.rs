use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that matches a single literal value.
#[derive(Debug, Clone)]
pub struct LiteralSchema {
    pub value: Value,
}

impl LiteralSchema {
    pub fn new(value: Value) -> Self {
        LiteralSchema { value }
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

impl Schema for LiteralSchema {
    fn kind(&self) -> &str {
        "literal"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        if input == &self.value {
            Ok(input.clone())
        } else {
            Err(vec![ValidationIssue {
                code: INVALID_LITERAL.to_string(),
                path: path.to_vec(),
                expected: display_value(&self.value),
                received: display_value(input),
                meta: None,
            }])
        }
    }

    fn export_node(&self) -> Value {
        json!({"kind": "literal", "value": self.value})
    }
}
