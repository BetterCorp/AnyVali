use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for record (string-keyed map with uniform value type) validation.
#[derive(Debug, Clone)]
pub struct RecordSchema {
    pub values: Box<dyn Schema>,
}

impl RecordSchema {
    pub fn new(values: Box<dyn Schema>) -> Self {
        RecordSchema { values }
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
        json!({
            "kind": "record",
            "values": self.values.export_node()
        })
    }
}
