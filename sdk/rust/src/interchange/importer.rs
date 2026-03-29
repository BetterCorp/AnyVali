use serde_json::Value;
use std::collections::HashMap;

use crate::parse::coerce::parse_coerce_config;
use crate::schema::{ParseContext, Schema};
use crate::schemas::*;
use crate::types::{AnyValiDocument, UnknownKeyMode};

/// Import an AnyValiDocument from a JSON string.
pub fn import_from_json(json_str: &str) -> Result<(Box<dyn Schema>, ParseContext), String> {
    let doc: AnyValiDocument =
        serde_json::from_str(json_str).map_err(|e| format!("Parse error: {}", e))?;
    import_document(&doc)
}

/// Import an AnyValiDocument from a serde_json::Value.
pub fn import_from_value(value: &Value) -> Result<(Box<dyn Schema>, ParseContext), String> {
    let doc: AnyValiDocument =
        serde_json::from_value(value.clone()).map_err(|e| format!("Parse error: {}", e))?;
    import_document(&doc)
}

/// Import an AnyValiDocument, returning the root schema and a context with definitions.
pub fn import_document(doc: &AnyValiDocument) -> Result<(Box<dyn Schema>, ParseContext), String> {
    // First pass: import all definitions
    let mut definitions: HashMap<String, Box<dyn Schema>> = HashMap::new();
    for (name, node) in &doc.definitions {
        let schema = import_node(node)?;
        definitions.insert(name.clone(), schema);
    }

    // Import root
    let root = import_node(&doc.root)?;

    let ctx = ParseContext { definitions };
    Ok((root, ctx))
}

/// Import a single schema node from its JSON representation.
pub fn import_node(node: &Value) -> Result<Box<dyn Schema>, String> {
    let obj = node
        .as_object()
        .ok_or_else(|| "Schema node must be an object".to_string())?;
    let kind = obj
        .get("kind")
        .and_then(|v| v.as_str())
        .ok_or_else(|| "Schema node must have a 'kind' field".to_string())?;

    match kind {
        "string" => {
            let mut schema = StringSchema::new();
            if let Some(v) = obj.get("minLength").and_then(|v| v.as_u64()) {
                schema.min_length = Some(v as usize);
            }
            if let Some(v) = obj.get("maxLength").and_then(|v| v.as_u64()) {
                schema.max_length = Some(v as usize);
            }
            if let Some(v) = obj.get("pattern").and_then(|v| v.as_str()) {
                schema.pattern = Some(v.to_string());
            }
            if let Some(v) = obj.get("startsWith").and_then(|v| v.as_str()) {
                schema.starts_with = Some(v.to_string());
            }
            if let Some(v) = obj.get("endsWith").and_then(|v| v.as_str()) {
                schema.ends_with = Some(v.to_string());
            }
            if let Some(v) = obj.get("includes").and_then(|v| v.as_str()) {
                schema.includes = Some(v.to_string());
            }
            if let Some(v) = obj.get("format").and_then(|v| v.as_str()) {
                schema.format = Some(v.to_string());
            }
            if let Some(v) = obj.get("coerce") {
                schema.coerce = Some(parse_coerce_config(v));
            }
            if let Some(v) = obj.get("default") {
                schema.default_value = Some(v.clone());
            }
            Ok(Box::new(schema))
        }

        "number" | "float64" => {
            let mut schema = if kind == "float64" {
                NumberSchema::float64()
            } else {
                NumberSchema::new()
            };
            import_numeric_constraints(obj, &mut schema);
            if let Some(v) = obj.get("coerce") {
                schema.coerce = Some(parse_coerce_config(v));
            }
            if let Some(v) = obj.get("default") {
                schema.default_value = Some(v.clone());
            }
            Ok(Box::new(schema))
        }

        "float32" => {
            let mut schema = NumberSchema::float32();
            import_numeric_constraints(obj, &mut schema);
            if let Some(v) = obj.get("coerce") {
                schema.coerce = Some(parse_coerce_config(v));
            }
            if let Some(v) = obj.get("default") {
                schema.default_value = Some(v.clone());
            }
            Ok(Box::new(schema))
        }

        "int" | "int64" => {
            let width = if kind == "int64" {
                IntWidth::Int64
            } else {
                IntWidth::Default
            };
            let mut schema = IntSchema::with_width(width);
            import_int_constraints(obj, &mut schema);
            if let Some(v) = obj.get("coerce") {
                schema.coerce = Some(parse_coerce_config(v));
            }
            if let Some(v) = obj.get("default") {
                schema.default_value = Some(v.clone());
            }
            Ok(Box::new(schema))
        }

        "int8" => {
            let mut schema = IntSchema::with_width(IntWidth::Int8);
            import_int_constraints(obj, &mut schema);
            Ok(Box::new(schema))
        }
        "int16" => {
            let mut schema = IntSchema::with_width(IntWidth::Int16);
            import_int_constraints(obj, &mut schema);
            Ok(Box::new(schema))
        }
        "int32" => {
            let mut schema = IntSchema::with_width(IntWidth::Int32);
            import_int_constraints(obj, &mut schema);
            Ok(Box::new(schema))
        }
        "uint8" => {
            let mut schema = IntSchema::with_width(IntWidth::UInt8);
            import_int_constraints(obj, &mut schema);
            Ok(Box::new(schema))
        }
        "uint16" => {
            let mut schema = IntSchema::with_width(IntWidth::UInt16);
            import_int_constraints(obj, &mut schema);
            Ok(Box::new(schema))
        }
        "uint32" => {
            let mut schema = IntSchema::with_width(IntWidth::UInt32);
            import_int_constraints(obj, &mut schema);
            Ok(Box::new(schema))
        }
        "uint64" => {
            let mut schema = IntSchema::with_width(IntWidth::UInt64);
            import_int_constraints(obj, &mut schema);
            Ok(Box::new(schema))
        }

        "bool" => {
            let mut schema = BoolSchema::new();
            if let Some(v) = obj.get("coerce") {
                schema.coerce = Some(parse_coerce_config(v));
            }
            if let Some(v) = obj.get("default") {
                schema.default_value = Some(v.clone());
            }
            Ok(Box::new(schema))
        }

        "null" => Ok(Box::new(NullSchema::new())),

        "any" => Ok(Box::new(AnySchema::new())),

        "unknown" => Ok(Box::new(UnknownSchema::new())),

        "never" => Ok(Box::new(NeverSchema::new())),

        "literal" => {
            let value = obj
                .get("value")
                .ok_or_else(|| "Literal schema must have a 'value' field".to_string())?;
            Ok(Box::new(LiteralSchema::new(value.clone())))
        }

        "enum" => {
            let values = obj
                .get("values")
                .and_then(|v| v.as_array())
                .ok_or_else(|| "Enum schema must have a 'values' array".to_string())?;
            Ok(Box::new(EnumSchema::new(values.clone())))
        }

        "array" => {
            let items_node = obj
                .get("items")
                .ok_or_else(|| "Array schema must have an 'items' field".to_string())?;
            let items = import_node(items_node)?;
            let mut schema = ArraySchema::new(items);
            if let Some(v) = obj.get("minItems").and_then(|v| v.as_u64()) {
                schema.min_items = Some(v as usize);
            }
            if let Some(v) = obj.get("maxItems").and_then(|v| v.as_u64()) {
                schema.max_items = Some(v as usize);
            }
            Ok(Box::new(schema))
        }

        "tuple" => {
            let elements_arr = obj
                .get("elements")
                .and_then(|v| v.as_array())
                .ok_or_else(|| "Tuple schema must have an 'elements' array".to_string())?;
            let elements: Result<Vec<Box<dyn Schema>>, String> =
                elements_arr.iter().map(|e| import_node(e)).collect();
            Ok(Box::new(TupleSchema::new(elements?)))
        }

        "object" => {
            let mut schema = ObjectSchema::new();

            if let Some(props) = obj.get("properties").and_then(|v| v.as_object()) {
                for (name, prop_node) in props {
                    let prop_obj = prop_node.as_object();
                    let default_val = prop_obj.and_then(|o| o.get("default")).cloned();
                    let prop_schema = import_node(prop_node)?;

                    schema.properties.insert(
                        name.clone(),
                        ObjectField {
                            schema: prop_schema,
                            default_value: default_val,
                        },
                    );
                }
            }

            if let Some(required) = obj.get("required").and_then(|v| v.as_array()) {
                schema.required = required
                    .iter()
                    .filter_map(|v| v.as_str().map(|s| s.to_string()))
                    .collect();
            }

            if let Some(uk) = obj.get("unknownKeys").and_then(|v| v.as_str()) {
                schema.unknown_keys = match uk {
                    "strip" => UnknownKeyMode::Strip,
                    "allow" => UnknownKeyMode::Allow,
                    _ => UnknownKeyMode::Reject,
                };
            }

            Ok(Box::new(schema))
        }

        "record" => {
            let values_node = obj
                .get("values")
                .ok_or_else(|| "Record schema must have a 'values' field".to_string())?;
            let values = import_node(values_node)?;
            Ok(Box::new(RecordSchema::new(values)))
        }

        "union" => {
            let variants_arr = obj
                .get("variants")
                .and_then(|v| v.as_array())
                .ok_or_else(|| "Union schema must have a 'variants' array".to_string())?;
            let variants: Result<Vec<Box<dyn Schema>>, String> =
                variants_arr.iter().map(|v| import_node(v)).collect();
            Ok(Box::new(UnionSchema::new(variants?)))
        }

        "intersection" => {
            let all_of_arr = obj
                .get("allOf")
                .and_then(|v| v.as_array())
                .ok_or_else(|| "Intersection schema must have an 'allOf' array".to_string())?;
            let all_of: Result<Vec<Box<dyn Schema>>, String> =
                all_of_arr.iter().map(|v| import_node(v)).collect();
            Ok(Box::new(IntersectionSchema::new(all_of?)))
        }

        "optional" => {
            let inner_node = obj
                .get("schema")
                .ok_or_else(|| "Optional schema must have a 'schema' field".to_string())?;
            let inner = import_node(inner_node)?;
            let mut schema = OptionalSchema::new(inner);
            if let Some(v) = obj.get("default") {
                schema.default_value = Some(v.clone());
            }
            Ok(Box::new(schema))
        }

        "nullable" => {
            let inner_node = obj
                .get("schema")
                .ok_or_else(|| "Nullable schema must have a 'schema' field".to_string())?;
            let inner = import_node(inner_node)?;
            let mut schema = NullableSchema::new(inner);
            if let Some(v) = obj.get("default") {
                schema.default_value = Some(v.clone());
            }
            Ok(Box::new(schema))
        }

        "ref" => {
            let ref_path = obj
                .get("ref")
                .and_then(|v| v.as_str())
                .ok_or_else(|| "Ref schema must have a 'ref' field".to_string())?;
            Ok(Box::new(RefSchema::new(ref_path)))
        }

        other => Err(format!("Unsupported schema kind: {}", other)),
    }
}

fn import_numeric_constraints(
    obj: &serde_json::Map<String, Value>,
    schema: &mut NumberSchema,
) {
    if let Some(v) = obj.get("min").and_then(|v| v.as_f64()) {
        schema.min = Some(v);
    }
    if let Some(v) = obj.get("max").and_then(|v| v.as_f64()) {
        schema.max = Some(v);
    }
    if let Some(v) = obj.get("exclusiveMin").and_then(|v| v.as_f64()) {
        schema.exclusive_min = Some(v);
    }
    if let Some(v) = obj.get("exclusiveMax").and_then(|v| v.as_f64()) {
        schema.exclusive_max = Some(v);
    }
    if let Some(v) = obj.get("multipleOf").and_then(|v| v.as_f64()) {
        schema.multiple_of = Some(v);
    }
}

fn import_int_constraints(obj: &serde_json::Map<String, Value>, schema: &mut IntSchema) {
    if let Some(v) = obj.get("min").and_then(|v| v.as_f64()) {
        schema.min = Some(v);
    }
    if let Some(v) = obj.get("max").and_then(|v| v.as_f64()) {
        schema.max = Some(v);
    }
    if let Some(v) = obj.get("exclusiveMin").and_then(|v| v.as_f64()) {
        schema.exclusive_min = Some(v);
    }
    if let Some(v) = obj.get("exclusiveMax").and_then(|v| v.as_f64()) {
        schema.exclusive_max = Some(v);
    }
    if let Some(v) = obj.get("multipleOf").and_then(|v| v.as_f64()) {
        schema.multiple_of = Some(v);
    }
}
