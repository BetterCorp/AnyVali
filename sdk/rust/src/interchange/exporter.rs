use serde_json::Value;
use std::collections::HashMap;

use crate::schema::Schema;
use crate::types::{AnyValiDocument, ExportMode};

/// Export a schema to an AnyValiDocument.
pub fn export_schema(
    schema: &dyn Schema,
    mode: ExportMode,
    definitions: &HashMap<String, Box<dyn Schema>>,
) -> Result<AnyValiDocument, String> {
    if mode == ExportMode::Portable && schema.has_custom_validators() {
        return Err("Cannot export schema with custom validators in portable mode".to_string());
    }

    let root = schema.export_node();

    let mut defs = serde_json::Map::new();
    for (name, def_schema) in definitions {
        if mode == ExportMode::Portable && def_schema.has_custom_validators() {
            return Err(format!(
                "Cannot export definition '{}' with custom validators in portable mode",
                name
            ));
        }
        defs.insert(name.clone(), def_schema.export_node());
    }

    Ok(AnyValiDocument {
        anyvali_version: "1.0".to_string(),
        schema_version: "1".to_string(),
        root,
        definitions: defs,
        extensions: serde_json::Map::new(),
    })
}

/// Serialize an AnyValiDocument to a JSON string.
pub fn document_to_json(doc: &AnyValiDocument) -> Result<String, String> {
    serde_json::to_string_pretty(doc).map_err(|e| format!("Serialization error: {}", e))
}

/// Serialize an AnyValiDocument to a serde_json::Value.
pub fn document_to_value(doc: &AnyValiDocument) -> Result<Value, String> {
    serde_json::to_value(doc).map_err(|e| format!("Serialization error: {}", e))
}
