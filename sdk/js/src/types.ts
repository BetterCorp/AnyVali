// ---- Schema Node (interchange JSON representation) ----

export type SchemaKind =
  | "any"
  | "unknown"
  | "never"
  | "null"
  | "bool"
  | "string"
  | "number"
  | "int"
  | "float32"
  | "float64"
  | "int8"
  | "int16"
  | "int32"
  | "int64"
  | "uint8"
  | "uint16"
  | "uint32"
  | "uint64"
  | "literal"
  | "enum"
  | "array"
  | "tuple"
  | "object"
  | "record"
  | "union"
  | "intersection"
  | "optional"
  | "nullable"
  | "ref";

export interface SchemaNodeBase {
  kind: SchemaKind;
  default?: unknown;
  coerce?: CoercionConfig | string | string[];
}

export interface CoercionConfig {
  from?: string;
  trim?: boolean;
  lower?: boolean;
  upper?: boolean;
}

// String node
export interface StringSchemaNode extends SchemaNodeBase {
  kind: "string";
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  startsWith?: string;
  endsWith?: string;
  includes?: string;
  format?: StringFormat;
}

export type StringFormat =
  | "email"
  | "url"
  | "uuid"
  | "ipv4"
  | "ipv6"
  | "date"
  | "date-time";

// Numeric nodes
export interface NumericSchemaNode extends SchemaNodeBase {
  kind:
    | "number"
    | "int"
    | "float32"
    | "float64"
    | "int8"
    | "int16"
    | "int32"
    | "int64"
    | "uint8"
    | "uint16"
    | "uint32"
    | "uint64";
  min?: number;
  max?: number;
  exclusiveMin?: number;
  exclusiveMax?: number;
  multipleOf?: number;
}

// Simple nodes
export interface SimpleSchemaNode extends SchemaNodeBase {
  kind: "any" | "unknown" | "never" | "null" | "bool";
}

// Literal node
export interface LiteralSchemaNode extends SchemaNodeBase {
  kind: "literal";
  value: string | number | boolean | null;
}

// Enum node
export interface EnumSchemaNode extends SchemaNodeBase {
  kind: "enum";
  values: (string | number)[];
}

// Array node
export interface ArraySchemaNode extends SchemaNodeBase {
  kind: "array";
  items: SchemaNode;
  minItems?: number;
  maxItems?: number;
}

// Tuple node
export interface TupleSchemaNode extends SchemaNodeBase {
  kind: "tuple";
  items: SchemaNode[];
}

// Object node
export interface ObjectSchemaNode extends SchemaNodeBase {
  kind: "object";
  properties: Record<string, SchemaNode>;
  required: string[];
  unknownKeys?: UnknownKeyMode;
}

// Record node
export interface RecordSchemaNode extends SchemaNodeBase {
  kind: "record";
  valueSchema: SchemaNode;
}

// Union node
export interface UnionSchemaNode extends SchemaNodeBase {
  kind: "union";
  variants: SchemaNode[];
}

// Intersection node
export interface IntersectionSchemaNode extends SchemaNodeBase {
  kind: "intersection";
  allOf: SchemaNode[];
}

// Optional node
export interface OptionalSchemaNode extends SchemaNodeBase {
  kind: "optional";
  inner: SchemaNode;
}

// Nullable node
export interface NullableSchemaNode extends SchemaNodeBase {
  kind: "nullable";
  inner: SchemaNode;
}

// Ref node
export interface RefSchemaNode extends SchemaNodeBase {
  kind: "ref";
  ref: string;
}

export type SchemaNode =
  | StringSchemaNode
  | NumericSchemaNode
  | SimpleSchemaNode
  | LiteralSchemaNode
  | EnumSchemaNode
  | ArraySchemaNode
  | TupleSchemaNode
  | ObjectSchemaNode
  | RecordSchemaNode
  | UnionSchemaNode
  | IntersectionSchemaNode
  | OptionalSchemaNode
  | NullableSchemaNode
  | RefSchemaNode;

// ---- Parse Result ----

export type ParseResult<T> =
  | { success: true; data: T }
  | { success: false; issues: ValidationIssue[] };

export interface ValidationIssue {
  code: string;
  message: string;
  path: (string | number)[];
  expected?: string;
  received?: string;
  meta?: Record<string, unknown>;
}

// ---- Interchange Document ----

export interface AnyValiDocument {
  anyvaliVersion: string;
  schemaVersion: string;
  root: SchemaNode;
  definitions: Record<string, SchemaNode>;
  extensions: Record<string, Record<string, unknown>>;
}

// ---- Modes ----

export type ExportMode = "portable" | "extended";

export type UnknownKeyMode = "reject" | "strip" | "allow";

// ---- Parse Context (internal) ----

export interface ParseContext {
  path: (string | number)[];
  issues: ValidationIssue[];
  definitions?: Record<string, SchemaNode>;
}
