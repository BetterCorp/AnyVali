use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Schema for array validation.
#[derive(Debug, Clone)]
pub struct ArraySchema {
    pub items: Box<dyn Schema>,
    pub min_items: Option<usize>,
    pub max_items: Option<usize>,
}

impl ArraySchema {
    pub fn new(items: Box<dyn Schema>) -> Self {
        ArraySchema {
            items,
            min_items: None,
            max_items: None,
        }
    }

    pub fn min_items(mut self, n: usize) -> Self {
        self.min_items = Some(n);
        self
    }

    pub fn max_items(mut self, n: usize) -> Self {
        self.max_items = Some(n);
        self
    }
}

impl Schema for ArraySchema {
    fn kind(&self) -> &str {
        "array"
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        let arr = match input {
            Value::Array(a) => a,
            other => {
                return Err(vec![ValidationIssue {
                    code: INVALID_TYPE.to_string(),
                    path: path.to_vec(),
                    expected: "array".to_string(),
                    received: value_type_name(other).to_string(),
                    meta: None,
                }]);
            }
        };

        let mut issues = Vec::new();

        if let Some(min) = self.min_items {
            if arr.len() < min {
                issues.push(ValidationIssue {
                    code: TOO_SMALL.to_string(),
                    path: path.to_vec(),
                    expected: min.to_string(),
                    received: arr.len().to_string(),
                    meta: None,
                });
            }
        }

        if let Some(max) = self.max_items {
            if arr.len() > max {
                issues.push(ValidationIssue {
                    code: TOO_LARGE.to_string(),
                    path: path.to_vec(),
                    expected: max.to_string(),
                    received: arr.len().to_string(),
                    meta: None,
                });
            }
        }

        if !issues.is_empty() {
            return Err(issues);
        }

        let mut result = Vec::new();
        for (i, item) in arr.iter().enumerate() {
            let mut item_path = path.to_vec();
            item_path.push(PathSegment::Index(i));
            match self.items.parse_value(item, &item_path, ctx) {
                Ok(v) => result.push(v),
                Err(mut errs) => issues.append(&mut errs),
            }
        }

        if issues.is_empty() {
            Ok(Value::Array(result))
        } else {
            Err(issues)
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({
            "kind": "array",
            "items": self.items.export_node()
        });
        let obj = node.as_object_mut().unwrap();
        if let Some(v) = self.min_items {
            obj.insert("minItems".to_string(), json!(v));
        }
        if let Some(v) = self.max_items {
            obj.insert("maxItems".to_string(), json!(v));
        }
        node
    }
}
