use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that refers to a named definition.
#[derive(Debug, Clone)]
pub struct RefSchema {
    pub ref_path: String,
}

impl RefSchema {
    pub fn new(ref_path: &str) -> Self {
        RefSchema {
            ref_path: ref_path.to_string(),
        }
    }

    /// Extract the definition name from a ref path like "#/definitions/User".
    fn definition_name(&self) -> &str {
        if let Some(name) = self.ref_path.strip_prefix("#/definitions/") {
            name
        } else {
            &self.ref_path
        }
    }
}

impl Schema for RefSchema {
    fn kind(&self) -> &str {
        "ref"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        let name = self.definition_name();
        match ctx.definitions.get(name) {
            Some(schema) => schema.parse_value(input, path, ctx),
            None => Err(vec![ValidationIssue {
                code: UNSUPPORTED_SCHEMA_KIND.to_string(),
                path: path.to_vec(),
                expected: format!("ref({})", self.ref_path),
                received: "unresolved reference".to_string(),
                meta: None,
            }]),
        }
    }

    fn export_node(&self) -> Value {
        json!({
            "kind": "ref",
            "ref": self.ref_path
        })
    }
}
