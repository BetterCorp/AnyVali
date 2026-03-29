use serde_json::{json, Value};

use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue};

/// Schema for intersection (all schemas must pass) validation.
#[derive(Debug, Clone)]
pub struct IntersectionSchema {
    pub all_of: Vec<Box<dyn Schema>>,
}

impl IntersectionSchema {
    pub fn new(all_of: Vec<Box<dyn Schema>>) -> Self {
        IntersectionSchema { all_of }
    }
}

impl Schema for IntersectionSchema {
    fn kind(&self) -> &str {
        "intersection"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        let mut all_issues = Vec::new();
        let mut last_result = input.clone();

        for schema in &self.all_of {
            match schema.parse_value(input, path, ctx) {
                Ok(v) => {
                    // For objects, merge results
                    if let (Value::Object(existing), Value::Object(new)) =
                        (&mut last_result, &v)
                    {
                        for (k, val) in new {
                            existing.insert(k.clone(), val.clone());
                        }
                    } else {
                        last_result = v;
                    }
                }
                Err(mut errs) => all_issues.append(&mut errs),
            }
        }

        if all_issues.is_empty() {
            Ok(last_result)
        } else {
            Err(all_issues)
        }
    }

    fn export_node(&self) -> Value {
        json!({
            "kind": "intersection",
            "allOf": self.all_of.iter().map(|s| s.export_node()).collect::<Vec<_>>()
        })
    }
}
