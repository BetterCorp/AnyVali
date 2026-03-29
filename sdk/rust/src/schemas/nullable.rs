use serde_json::{json, Value};

use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema wrapper that allows null in addition to the inner schema's type.
#[derive(Debug, Clone)]
pub struct NullableSchema {
    pub schema: Box<dyn Schema>,
    pub default_value: Option<Value>,
}

impl NullableSchema {
    pub fn new(schema: Box<dyn Schema>) -> Self {
        NullableSchema {
            schema,
            default_value: None,
        }
    }

    pub fn default(mut self, v: Value) -> Self {
        self.default_value = Some(v);
        self
    }
}

impl Schema for NullableSchema {
    fn kind(&self) -> &str {
        "nullable"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        if input.is_null() {
            return Ok(Value::Null);
        }
        self.schema.parse_value(input, path, ctx)
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "nullable",
            "schema": self.schema.export_node()
        });
        if let Some(v) = &self.default_value {
            node.as_object_mut()
                .unwrap()
                .insert("default".to_string(), v.clone());
        }
        node
    }
}
