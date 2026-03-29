use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that matches one of a set of allowed values.
#[derive(Debug, Clone)]
pub struct EnumSchema {
    pub values: Vec<Value>,
}

impl EnumSchema {
    pub fn new(values: Vec<Value>) -> Self {
        EnumSchema { values }
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
        json!({"kind": "enum", "values": self.values})
    }
}
