use serde_json::{json, Value};

use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema wrapper that makes a field optional (absent is allowed, but if present must validate).
/// Note: optional is only meaningful inside object schemas. At the top level,
/// it delegates to its inner schema.
#[derive(Debug, Clone)]
pub struct OptionalSchema {
    pub schema: Box<dyn Schema>,
    pub default_value: Option<Value>,
}

impl OptionalSchema {
    pub fn new(schema: Box<dyn Schema>) -> Self {
        OptionalSchema {
            schema,
            default_value: None,
        }
    }

    pub fn default(mut self, v: Value) -> Self {
        self.default_value = Some(v);
        self
    }
}

impl Schema for OptionalSchema {
    fn kind(&self) -> &str {
        "optional"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        // When used at field level, absent is handled by the object schema.
        // If we get here, the value is present and must validate.
        self.schema.parse_value(input, path, ctx)
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "optional",
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
