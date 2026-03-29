use serde_json::Value;
use std::collections::HashMap;

use crate::types::{ParseResult, PathSegment, ValidationError, ValidationIssue};

/// Context passed through during parsing, carrying definitions for ref resolution.
#[derive(Debug, Clone)]
pub struct ParseContext {
    pub definitions: HashMap<String, Box<dyn Schema>>,
}

impl ParseContext {
    pub fn new() -> Self {
        ParseContext {
            definitions: HashMap::new(),
        }
    }
}

impl Default for ParseContext {
    fn default() -> Self {
        Self::new()
    }
}

/// The core Schema trait. All schema types implement this.
pub trait Schema: SchemaClone + std::fmt::Debug + Send + Sync {
    /// The kind string for this schema (e.g., "string", "int", "object").
    fn kind(&self) -> &str;

    /// Validate and parse the input value. Returns the (possibly transformed) value or issues.
    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>>;

    /// Throwing parse: returns Ok(Value) or Err(ValidationError).
    fn parse(&self, input: &Value) -> Result<Value, ValidationError> {
        self.parse_with_context(input, &ParseContext::new())
    }

    /// Throwing parse with context (for refs).
    fn parse_with_context(
        &self,
        input: &Value,
        ctx: &ParseContext,
    ) -> Result<Value, ValidationError> {
        match self.parse_value(input, &[], ctx) {
            Ok(v) => Ok(v),
            Err(issues) => Err(ValidationError { issues }),
        }
    }

    /// Non-throwing parse: returns ParseResult with success/failure info.
    fn safe_parse(&self, input: &Value) -> ParseResult {
        self.safe_parse_with_context(input, &ParseContext::new())
    }

    /// Non-throwing parse with context.
    fn safe_parse_with_context(&self, input: &Value, ctx: &ParseContext) -> ParseResult {
        match self.parse_value(input, &[], ctx) {
            Ok(v) => ParseResult::ok(v),
            Err(issues) => ParseResult::err(issues),
        }
    }

    /// Export this schema to a JSON node representation.
    fn export_node(&self) -> Value;

    /// Whether this schema has non-portable features (custom validators).
    fn has_custom_validators(&self) -> bool {
        false
    }
}

/// Trait for cloning boxed schemas.
pub trait SchemaClone {
    fn clone_box(&self) -> Box<dyn Schema>;
}

impl<T: Schema + Clone + 'static> SchemaClone for T {
    fn clone_box(&self) -> Box<dyn Schema> {
        Box::new(self.clone())
    }
}

impl Clone for Box<dyn Schema> {
    fn clone(&self) -> Box<dyn Schema> {
        self.clone_box()
    }
}
