use std::collections::HashMap;
use std::fmt;

use serde_json::{json, Map, Value};

use crate::interchange::importer::import_node;
use crate::issue_codes::UNSUPPORTED_SCHEMA_KIND;
use crate::{ParseResult, PathSegment, Schema, ValidationError, ValidationIssue};

const ENCRYPTED_PREFIX: &str = "encrypted:";

#[derive(Debug)]
pub enum SensitiveError {
    Validation(ValidationError),
    Transform(String),
    Schema(String),
}

impl fmt::Display for SensitiveError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Validation(error) => fmt::Display::fmt(error, f),
            Self::Transform(error) | Self::Schema(error) => f.write_str(error),
        }
    }
}

impl std::error::Error for SensitiveError {}

impl From<ValidationError> for SensitiveError {
    fn from(value: ValidationError) -> Self {
        Self::Validation(value)
    }
}

/// Validate encrypted storage data. Sensitive nodes are opaque `encrypted:*` strings.
pub fn safe_parse_encrypted(schema: &dyn Schema, data: &Value) -> ParseResult {
    match encrypted_schema(schema) {
        Ok(schema) => schema.safe_parse(data),
        Err(error) => ParseResult::err(vec![ValidationIssue {
            code: UNSUPPORTED_SCHEMA_KIND.to_string(),
            path: vec![],
            expected: "valid schema".to_string(),
            received: error.to_string(),
            meta: None,
        }]),
    }
}

/// Validate plaintext, transform every sensitive node, then validate encrypted storage data.
pub fn encrypt<F, E>(
    schema: &dyn Schema,
    data: &Value,
    transform: F,
) -> Result<Value, SensitiveError>
where
    F: Fn(&[PathSegment], &Value) -> Result<Value, E>,
    E: fmt::Display,
{
    let plain = schema.parse(data)?;
    let node = schema.export_node();
    let encrypted = transform_node(
        &node,
        &plain,
        &transform,
        true,
        &mut vec![],
        &mut HashMap::new(),
    )?;
    let checked = encrypted_schema(schema)?.parse(&encrypted)?;
    Ok(checked)
}

/// Validate encrypted storage data, transform every sensitive node, then validate plaintext.
pub fn decrypt<F, E>(
    schema: &dyn Schema,
    data: &Value,
    transform: F,
) -> Result<Value, SensitiveError>
where
    F: Fn(&[PathSegment], &Value) -> Result<Value, E>,
    E: fmt::Display,
{
    let encrypted = encrypted_schema(schema)?.parse(data)?;
    let node = schema.export_node();
    let plain = transform_node(
        &node,
        &encrypted,
        &transform,
        false,
        &mut vec![],
        &mut HashMap::new(),
    )?;
    Ok(schema.parse(&plain)?)
}

fn encrypted_schema(schema: &dyn Schema) -> Result<Box<dyn Schema>, SensitiveError> {
    import_node(&encrypted_node(&schema.export_node())).map_err(SensitiveError::Schema)
}

fn encrypted_node(node: &Value) -> Value {
    let Some(object) = node.as_object() else {
        return node.clone();
    };

    if is_sensitive(object) {
        let marker = json!({
            "kind": "string",
            "minLength": ENCRYPTED_PREFIX.len() + 1,
            "startsWith": ENCRYPTED_PREFIX,
        });
        return match object.get("kind").and_then(Value::as_str) {
            Some("optional") | Some("nullable") => json!({
                "kind": object["kind"],
                "schema": marker,
            }),
            _ => marker,
        };
    }

    let mut projected = object.clone();
    match object.get("kind").and_then(Value::as_str) {
        Some("object") => project_map(&mut projected, "properties"),
        Some("array") => project_child(&mut projected, "items"),
        Some("tuple") => project_list(&mut projected, "elements"),
        Some("record") => project_child(&mut projected, "values"),
        Some("union") => project_list(&mut projected, "variants"),
        Some("intersection") => project_list(&mut projected, "allOf"),
        Some("optional") | Some("nullable") => project_child(&mut projected, "schema"),
        _ => {}
    }
    Value::Object(projected)
}

fn project_child(node: &mut Map<String, Value>, key: &str) {
    if let Some(child) = node.get(key).cloned() {
        node.insert(key.to_string(), encrypted_node(&child));
    }
}

fn project_list(node: &mut Map<String, Value>, key: &str) {
    let projected = node
        .get(key)
        .and_then(Value::as_array)
        .map(|items| Value::Array(items.iter().map(encrypted_node).collect()));
    if let Some(projected) = projected {
        node.insert(key.to_string(), projected);
    }
}

fn project_map(node: &mut Map<String, Value>, key: &str) {
    let projected = node.get(key).and_then(Value::as_object).map(|items| {
        Value::Object(
            items
                .iter()
                .map(|(name, child)| (name.clone(), encrypted_node(child)))
                .collect(),
        )
    });
    if let Some(projected) = projected {
        node.insert(key.to_string(), projected);
    }
}

fn is_sensitive(node: &Map<String, Value>) -> bool {
    node.get("metadata")
        .and_then(Value::as_object)
        .and_then(|metadata| metadata.get("sensitive"))
        .and_then(Value::as_bool)
        == Some(true)
}

fn transform_node<F, E>(
    node: &Value,
    value: &Value,
    transform: &F,
    encrypting: bool,
    path: &mut Vec<PathSegment>,
    cache: &mut HashMap<String, Value>,
) -> Result<Value, SensitiveError>
where
    F: Fn(&[PathSegment], &Value) -> Result<Value, E>,
    E: fmt::Display,
{
    let Some(object) = node.as_object() else {
        return Ok(value.clone());
    };

    if is_sensitive(object) && !value.is_null() {
        if encrypting {
            import_node(node)
                .map_err(SensitiveError::Schema)?
                .parse(value)?;
        }
        let key = serde_json::to_string(path)
            .map_err(|error| SensitiveError::Schema(error.to_string()))?;
        if let Some(cached) = cache.get(&key) {
            return Ok(cached.clone());
        }
        let result =
            transform(path, value).map_err(|error| SensitiveError::Transform(error.to_string()))?;
        cache.insert(key, result.clone());
        return Ok(result);
    }

    match object.get("kind").and_then(Value::as_str) {
        Some("object") => {
            let Some(values) = value.as_object() else {
                return Ok(value.clone());
            };
            let properties = object.get("properties").and_then(Value::as_object);
            let mut result = values.clone();
            for (name, child) in properties.into_iter().flatten() {
                if let Some(child_value) = values.get(name) {
                    path.push(PathSegment::Key(name.clone()));
                    result.insert(
                        name.clone(),
                        transform_node(child, child_value, transform, encrypting, path, cache)?,
                    );
                    path.pop();
                }
            }
            Ok(Value::Object(result))
        }
        Some("array") | Some("tuple") => {
            let Some(values) = value.as_array() else {
                return Ok(value.clone());
            };
            let items = object.get("items");
            let elements = object.get("elements").and_then(Value::as_array);
            let mut result = Vec::with_capacity(values.len());
            for (index, child_value) in values.iter().enumerate() {
                let child = items.or_else(|| elements.and_then(|schemas| schemas.get(index)));
                path.push(PathSegment::Index(index));
                result.push(match child {
                    Some(child) => {
                        transform_node(child, child_value, transform, encrypting, path, cache)?
                    }
                    None => child_value.clone(),
                });
                path.pop();
            }
            Ok(Value::Array(result))
        }
        Some("record") => {
            let Some(values) = value.as_object() else {
                return Ok(value.clone());
            };
            let Some(child) = object.get("values") else {
                return Ok(value.clone());
            };
            let mut result = Map::new();
            for (name, child_value) in values {
                path.push(PathSegment::Key(name.clone()));
                result.insert(
                    name.clone(),
                    transform_node(child, child_value, transform, encrypting, path, cache)?,
                );
                path.pop();
            }
            Ok(Value::Object(result))
        }
        Some("optional") | Some("nullable") => object.get("schema").map_or_else(
            || Ok(value.clone()),
            |child| transform_node(child, value, transform, encrypting, path, cache),
        ),
        Some("union") => {
            let variants = object.get("variants").and_then(Value::as_array);
            for child in variants.into_iter().flatten() {
                let candidate = if encrypting {
                    child.clone()
                } else {
                    encrypted_node(child)
                };
                if import_node(&candidate)
                    .map_err(SensitiveError::Schema)?
                    .safe_parse(value)
                    .success
                {
                    return transform_node(child, value, transform, encrypting, path, cache);
                }
            }
            Ok(value.clone())
        }
        Some("intersection") => {
            let mut result = value.clone();
            for child in object
                .get("allOf")
                .and_then(Value::as_array)
                .into_iter()
                .flatten()
            {
                result = transform_node(child, &result, transform, encrypting, path, cache)?;
            }
            Ok(result)
        }
        _ => Ok(value.clone()),
    }
}
