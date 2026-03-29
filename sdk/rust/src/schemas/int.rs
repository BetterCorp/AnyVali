use serde_json::{json, Value};

use crate::issue_codes::*;
use crate::schema::{ParseContext, Schema};
use crate::types::{PathSegment, ValidationIssue, value_type_name};

/// Integer width specification.
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum IntWidth {
    /// int (alias for int64)
    Default,
    Int8,
    Int16,
    Int32,
    Int64,
    UInt8,
    UInt16,
    UInt32,
    UInt64,
}

impl IntWidth {
    pub fn kind_name(&self) -> &'static str {
        match self {
            IntWidth::Default => "int",
            IntWidth::Int8 => "int8",
            IntWidth::Int16 => "int16",
            IntWidth::Int32 => "int32",
            IntWidth::Int64 => "int64",
            IntWidth::UInt8 => "uint8",
            IntWidth::UInt16 => "uint16",
            IntWidth::UInt32 => "uint32",
            IntWidth::UInt64 => "uint64",
        }
    }

    pub fn min_value(&self) -> i128 {
        match self {
            IntWidth::Default | IntWidth::Int64 => i64::MIN as i128,
            IntWidth::Int8 => i8::MIN as i128,
            IntWidth::Int16 => i16::MIN as i128,
            IntWidth::Int32 => i32::MIN as i128,
            IntWidth::UInt8 | IntWidth::UInt16 | IntWidth::UInt32 | IntWidth::UInt64 => 0,
        }
    }

    pub fn max_value(&self) -> i128 {
        match self {
            IntWidth::Default | IntWidth::Int64 => i64::MAX as i128,
            IntWidth::Int8 => i8::MAX as i128,
            IntWidth::Int16 => i16::MAX as i128,
            IntWidth::Int32 => i32::MAX as i128,
            IntWidth::UInt8 => u8::MAX as i128,
            IntWidth::UInt16 => u16::MAX as i128,
            IntWidth::UInt32 => u32::MAX as i128,
            IntWidth::UInt64 => u64::MAX as i128,
        }
    }
}

/// Schema for integer validation with optional width and constraints.
#[derive(Debug, Clone)]
pub struct IntSchema {
    pub width: IntWidth,
    pub min: Option<f64>,
    pub max: Option<f64>,
    pub exclusive_min: Option<f64>,
    pub exclusive_max: Option<f64>,
    pub multiple_of: Option<f64>,
    pub coerce: Option<Vec<String>>,
    pub default_value: Option<Value>,
}

impl IntSchema {
    pub fn new() -> Self {
        IntSchema {
            width: IntWidth::Default,
            min: None,
            max: None,
            exclusive_min: None,
            exclusive_max: None,
            multiple_of: None,
            coerce: None,
            default_value: None,
        }
    }

    pub fn with_width(width: IntWidth) -> Self {
        IntSchema {
            width,
            ..Self::new()
        }
    }

    pub fn min(mut self, v: f64) -> Self {
        self.min = Some(v);
        self
    }

    pub fn max(mut self, v: f64) -> Self {
        self.max = Some(v);
        self
    }

    pub fn exclusive_min(mut self, v: f64) -> Self {
        self.exclusive_min = Some(v);
        self
    }

    pub fn exclusive_max(mut self, v: f64) -> Self {
        self.exclusive_max = Some(v);
        self
    }

    pub fn multiple_of(mut self, v: f64) -> Self {
        self.multiple_of = Some(v);
        self
    }

    pub fn coerce(mut self, c: Vec<String>) -> Self {
        self.coerce = Some(c);
        self
    }

    pub fn default(mut self, v: Value) -> Self {
        self.default_value = Some(v);
        self
    }
}

impl Default for IntSchema {
    fn default() -> Self {
        Self::new()
    }
}

/// Format a number for display, matching spec (integer display for whole numbers).
fn format_num(n: f64) -> String {
    if n.fract() == 0.0 && n.abs() < 1e18 {
        format!("{}", n as i64)
    } else {
        format!("{}", n)
    }
}

impl Schema for IntSchema {
    fn kind(&self) -> &str {
        self.width.kind_name()
    }

    fn parse_value(
        &self,
        input: &Value,
        path: &[PathSegment],
        _ctx: &ParseContext,
    ) -> Result<Value, Vec<ValidationIssue>> {
        let kind = self.width.kind_name();

        // Apply coercions
        let mut value = input.clone();
        if let Some(coercions) = &self.coerce {
            for c in coercions {
                if c == "string->int" {
                    if let Value::String(s) = &value {
                        let trimmed = s.trim();
                        match trimmed.parse::<i64>() {
                            Ok(n) => {
                                value = json!(n);
                            }
                            Err(_) => {
                                return Err(vec![ValidationIssue {
                                    code: COERCION_FAILED.to_string(),
                                    path: path.to_vec(),
                                    expected: kind.to_string(),
                                    received: s.clone(),
                                    meta: None,
                                }]);
                            }
                        }
                    }
                }
            }
        }

        // Must be a number
        let n = match &value {
            Value::Number(n) => n,
            other => {
                return Err(vec![ValidationIssue {
                    code: INVALID_TYPE.to_string(),
                    path: path.to_vec(),
                    expected: kind.to_string(),
                    received: value_type_name(other).to_string(),
                    meta: None,
                }]);
            }
        };

        let f = n.as_f64().unwrap();

        // Must be a whole number
        if f.fract() != 0.0 {
            return Err(vec![ValidationIssue {
                code: INVALID_TYPE.to_string(),
                path: path.to_vec(),
                expected: kind.to_string(),
                received: "number".to_string(),
                meta: None,
            }]);
        }

        let int_val = f as i128;

        // Width range check (for typed widths)
        match self.width {
            IntWidth::Default | IntWidth::Int64 => {
                // int64 range - checked by i64 conversion
            }
            _ => {
                let min_v = self.width.min_value();
                let max_v = self.width.max_value();
                if int_val < min_v {
                    return Err(vec![ValidationIssue {
                        code: TOO_SMALL.to_string(),
                        path: path.to_vec(),
                        expected: kind.to_string(),
                        received: format!("{}", int_val),
                        meta: None,
                    }]);
                }
                if int_val > max_v {
                    return Err(vec![ValidationIssue {
                        code: TOO_LARGE.to_string(),
                        path: path.to_vec(),
                        expected: kind.to_string(),
                        received: format!("{}", int_val),
                        meta: None,
                    }]);
                }
            }
        }

        // User constraints
        let mut issues = Vec::new();

        if let Some(min) = self.min {
            if f < min {
                issues.push(ValidationIssue {
                    code: TOO_SMALL.to_string(),
                    path: path.to_vec(),
                    expected: format_num(min),
                    received: format_num(f),
                    meta: None,
                });
            }
        }

        if let Some(max) = self.max {
            if f > max {
                issues.push(ValidationIssue {
                    code: TOO_LARGE.to_string(),
                    path: path.to_vec(),
                    expected: format_num(max),
                    received: format_num(f),
                    meta: None,
                });
            }
        }

        if let Some(emin) = self.exclusive_min {
            if f <= emin {
                issues.push(ValidationIssue {
                    code: TOO_SMALL.to_string(),
                    path: path.to_vec(),
                    expected: format_num(emin),
                    received: format_num(f),
                    meta: None,
                });
            }
        }

        if let Some(emax) = self.exclusive_max {
            if f >= emax {
                issues.push(ValidationIssue {
                    code: TOO_LARGE.to_string(),
                    path: path.to_vec(),
                    expected: format_num(emax),
                    received: format_num(f),
                    meta: None,
                });
            }
        }

        if let Some(mo) = self.multiple_of {
            let remainder = f % mo;
            if remainder.abs() > 1e-10 && (remainder - mo).abs() > 1e-10 {
                issues.push(ValidationIssue {
                    code: INVALID_NUMBER.to_string(),
                    path: path.to_vec(),
                    expected: format_num(mo),
                    received: format_num(f),
                    meta: None,
                });
            }
        }

        if issues.is_empty() {
            Ok(value)
        } else {
            Err(issues)
        }
    }

    fn export_node(&self) -> Value {
        let mut node = json!({"kind": self.width.kind_name()});
        let obj = node.as_object_mut().unwrap();
        if let Some(v) = self.min {
            obj.insert("min".to_string(), json!(v));
        }
        if let Some(v) = self.max {
            obj.insert("max".to_string(), json!(v));
        }
        if let Some(v) = self.exclusive_min {
            obj.insert("exclusiveMin".to_string(), json!(v));
        }
        if let Some(v) = self.exclusive_max {
            obj.insert("exclusiveMax".to_string(), json!(v));
        }
        if let Some(v) = self.multiple_of {
            obj.insert("multipleOf".to_string(), json!(v));
        }
        if let Some(v) = &self.coerce {
            if v.len() == 1 {
                obj.insert("coerce".to_string(), json!(v[0]));
            } else {
                obj.insert("coerce".to_string(), json!(v));
            }
        }
        if let Some(v) = &self.default_value {
            obj.insert("default".to_string(), v.clone());
        }
        node
    }
}
