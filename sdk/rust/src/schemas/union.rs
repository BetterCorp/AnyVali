use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for union (any of several variants) validation.
#[derive(Debug, Clone)]
pub struct UnionSchema {
    pub variants: Vec<Box<dyn Schema>>,
}

impl UnionSchema {
    pub fn new(variants: Vec<Box<dyn Schema>>) -> Self {
        UnionSchema { variants }
    }
}

impl Schema for UnionSchema {
    fn kind(&self) -> &str {
        "union"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        // Try each variant, return first success
        for variant in &self.variants {
            if let Ok(v) = variant.parse_value(input, path, ctx) {
                return Ok(v);
            }
        }

        // All variants failed
        let variant_names = self
            .variants
            .iter()
            .map(|v| v.kind().to_string())
            .collect::<Vec<_>>()
            .join(" | ");

        Err(vec![ValidationIssue {
            code: INVALID_UNION.to_string(),
            path: path.to_vec(),
            expected: variant_names,
            received: value_type_name(input).to_string(),
            meta: None,
        }])
    }

    fn export_node(&self) -> Value {
        json!({
            "kind": "union",
            "variants": self.variants.iter().map(|v| v.export_node()).collect::<Vec<_>>()
        })
    }
}
