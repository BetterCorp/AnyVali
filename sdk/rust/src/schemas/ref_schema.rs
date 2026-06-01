use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{
    add_metadata_to_node, build_describe_metadata, merge_metadata, reserved_metadata_keys,
    validate_metadata_keys, DescribeOpts, ParseContext, Schema,
};
use crate::types::{PathSegment, ValidationIssue};

/// Schema that refers to a named definition.
#[derive(Debug, Clone)]
pub struct RefSchema {
    pub ref_path: String,
    pub metadata: Option<Value>,
}

impl RefSchema {
    pub fn new(ref_path: &str) -> Self {
        RefSchema {
            ref_path: ref_path.to_string(),
            metadata: None,
        }
    }

    pub fn describe(mut self, description: &str, opts: Option<&DescribeOpts>) -> Self {
        let new_meta = build_describe_metadata(description, opts);
        match &mut self.metadata {
            Some(existing) => merge_metadata(existing, &new_meta),
            None => self.metadata = Some(new_meta),
        }
        self
    }

    pub fn with_metadata(mut self, meta: Value, replace: bool) -> Self {
        validate_metadata_keys(&meta);
        if replace {
            let reserved = reserved_metadata_keys();
            let mut preserved = serde_json::Map::new();
            if let Some(Value::Object(existing)) = &self.metadata {
                for (k, v) in existing {
                    if reserved.contains(k.as_str()) {
                        preserved.insert(k.clone(), v.clone());
                    }
                }
            }
            if let Value::Object(new_map) = &meta {
                for (k, v) in new_map {
                    preserved.insert(k.clone(), v.clone());
                }
            }
            self.metadata = Some(Value::Object(preserved));
        } else {
            match &mut self.metadata {
                Some(existing) => merge_metadata(existing, &meta),
                None => self.metadata = Some(meta),
            }
        }
        self
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
            Some(schema) => {
                let key = format!("{}@{:p}", name, input);
                if !ctx.enter_ref(key.clone()) {
                    return Err(vec![ValidationIssue {
                        code: INVALID_TYPE.to_string(),
                        path: path.to_vec(),
                        expected: format!("ref({})", self.ref_path),
                        received: "recursive reference".to_string(),
                        meta: None,
                    }]);
                }
                let result = schema.parse_value(input, path, ctx);
                ctx.exit_ref(&key);
                result
            }
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
        let mut node = json!({
            "kind": "ref",
            "ref": self.ref_path
        });
        add_metadata_to_node(&mut node, &self.metadata);
        node
    }
}
