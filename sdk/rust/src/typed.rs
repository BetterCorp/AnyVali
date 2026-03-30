use serde::de::DeserializeOwned;
use serde_json::Value;

use crate::schema::Schema;
use crate::schemas::{
    AnySchema, ArraySchema, BoolSchema, IntSchema, NullSchema, NumberSchema, StringSchema,
    UnknownSchema,
};
use crate::types::{ValidationError, ValidationIssue};

/// Trait for schemas that know their output Rust type.
/// Provides a typed parse method that deserializes the validated result.
pub trait TypedSchema: Schema {
    type Output: DeserializeOwned;

    /// Parse and validate the input, then deserialize the result to the associated Output type.
    fn parse_typed(&self, input: &Value) -> Result<Self::Output, ValidationError> {
        let value = self.parse(input)?;
        serde_json::from_value(value).map_err(|e| ValidationError {
            issues: vec![ValidationIssue {
                code: "type_conversion".to_string(),
                path: vec![],
                expected: std::any::type_name::<Self::Output>().to_string(),
                received: format!("{}", e),
                meta: None,
            }],
        })
    }
}

/// Parse input using any schema and deserialize the result to a specific type.
///
/// This is useful when working with `dyn Schema` trait objects where the
/// concrete type (and thus `TypedSchema`) is not known at compile time.
pub fn parse_as<T: DeserializeOwned>(
    schema: &dyn Schema,
    input: &Value,
) -> Result<T, ValidationError> {
    let value = schema.parse(input)?;
    serde_json::from_value(value).map_err(|e| ValidationError {
        issues: vec![ValidationIssue {
            code: "type_conversion".to_string(),
            path: vec![],
            expected: std::any::type_name::<T>().to_string(),
            received: format!("{}", e),
            meta: None,
        }],
    })
}

// --- TypedSchema implementations for concrete schema types ---

impl TypedSchema for StringSchema {
    type Output = String;
}

impl TypedSchema for BoolSchema {
    type Output = bool;
}

impl TypedSchema for NumberSchema {
    type Output = f64;
}

impl TypedSchema for IntSchema {
    type Output = i64;
}

impl TypedSchema for NullSchema {
    /// Null parses to `()` since `serde_json::Value::Null` deserializes to `()`.
    type Output = ();
}

impl TypedSchema for AnySchema {
    type Output = Value;
}

impl TypedSchema for UnknownSchema {
    type Output = Value;
}

impl TypedSchema for ArraySchema {
    type Output = Vec<Value>;
}
