use serde_json::{json, Value};
use std::collections::{HashMap, HashSet};
use std::sync::{Arc, Mutex};

use crate::types::{ParseResult, PathSegment, UnknownKeyMode, ValidationError, ValidationIssue};

/// Context passed through during parsing, carrying definitions for ref resolution.
#[derive(Debug, Clone)]
pub struct ParseContext {
    pub definitions: HashMap<String, Box<dyn Schema>>,
    pub inherited_unknown_keys: Option<UnknownKeyMode>,
    active_refs: Arc<Mutex<HashSet<String>>>,
}

impl ParseContext {
    pub fn new() -> Self {
        ParseContext {
            definitions: HashMap::new(),
            inherited_unknown_keys: None,
            active_refs: Arc::new(Mutex::new(HashSet::new())),
        }
    }

    pub fn with_definitions(definitions: HashMap<String, Box<dyn Schema>>) -> Self {
        ParseContext {
            definitions,
            inherited_unknown_keys: None,
            active_refs: Arc::new(Mutex::new(HashSet::new())),
        }
    }

    pub fn enter_ref(&self, key: String) -> bool {
        self.active_refs.lock().unwrap().insert(key)
    }

    pub fn exit_ref(&self, key: &str) {
        self.active_refs.lock().unwrap().remove(key);
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

    /// Schema-level default used by wrappers that carry their own default.
    fn default_value(&self) -> Option<&Value> {
        None
    }

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

/// Reserved metadata keys that must use describe() instead of metadata().
pub fn reserved_metadata_keys() -> HashSet<&'static str> {
    let mut keys = HashSet::new();
    keys.insert("title");
    keys.insert("description");
    keys.insert("deprecated");
    keys.insert("deprecatedMessage");
    keys.insert("notStable");
    keys.insert("since");
    keys.insert("sensitive");
    keys.insert("readonly");
    keys.insert("writeonly");
    keys.insert("examples");
    keys
}

/// Options for describe().
#[derive(Debug, Clone, Default)]
pub struct DescribeOpts {
    pub title: Option<String>,
    pub deprecated: Option<bool>,
    pub deprecated_message: Option<String>,
    pub not_stable: Option<bool>,
    pub since: Option<String>,
    pub sensitive: Option<bool>,
    pub readonly: Option<bool>,
    pub writeonly: Option<bool>,
    pub examples: Option<Vec<Value>>,
}

/// Validate and build metadata from describe() parameters.
pub fn build_describe_metadata(description: &str, opts: Option<&DescribeOpts>) -> Value {
    let mut meta = serde_json::Map::new();
    meta.insert("description".to_string(), json!(description));

    if let Some(opts) = opts {
        if let Some(title) = &opts.title {
            meta.insert("title".to_string(), json!(title));
        }
        if let Some(deprecated) = opts.deprecated {
            meta.insert("deprecated".to_string(), json!(deprecated));
        }
        if let Some(msg) = &opts.deprecated_message {
            if opts.deprecated != Some(true) {
                panic!("describe(): deprecatedMessage requires deprecated to be true");
            }
            meta.insert("deprecatedMessage".to_string(), json!(msg));
        }
        if let Some(not_stable) = opts.not_stable {
            meta.insert("notStable".to_string(), json!(not_stable));
        }
        if let Some(since) = &opts.since {
            meta.insert("since".to_string(), json!(since));
        }
        if let Some(sensitive) = opts.sensitive {
            meta.insert("sensitive".to_string(), json!(sensitive));
        }
        if let Some(readonly) = opts.readonly {
            meta.insert("readonly".to_string(), json!(readonly));
        }
        if let Some(writeonly) = opts.writeonly {
            meta.insert("writeonly".to_string(), json!(writeonly));
        }
        if opts.readonly == Some(true) && opts.writeonly == Some(true) {
            panic!("describe(): readonly and writeonly cannot both be true");
        }
        if let Some(examples) = &opts.examples {
            meta.insert("examples".to_string(), json!(examples));
        }
    }

    Value::Object(meta)
}

/// Validate metadata keys (for metadata() call - no reserved keys allowed).
pub fn validate_metadata_keys(meta: &Value) {
    let reserved = reserved_metadata_keys();
    if let Value::Object(map) = meta {
        for key in map.keys() {
            if reserved.contains(key.as_str()) {
                panic!(
                    "metadata(): \"{}\" is a reserved key. Use describe() instead.",
                    key
                );
            }
        }
    }
}

/// Merge metadata: shallow merge src into dst.
pub fn merge_metadata(dst: &mut Value, src: &Value) {
    if let (Value::Object(d), Value::Object(s)) = (dst, src) {
        for (k, v) in s {
            d.insert(k.clone(), v.clone());
        }
    }
}

/// Helper to add metadata to an export node.
pub fn add_metadata_to_node(node: &mut Value, metadata: &Option<Value>) {
    if let Some(meta) = metadata {
        if let Value::Object(map) = meta {
            if !map.is_empty() {
                if let Value::Object(obj) = node {
                    obj.insert("metadata".to_string(), meta.clone());
                }
            }
        }
    }
}
