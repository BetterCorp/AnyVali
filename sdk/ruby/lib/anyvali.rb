# frozen_string_literal: true

require_relative "anyvali/issue_codes"
require_relative "anyvali/validation_issue"
require_relative "anyvali/parse_result"
require_relative "anyvali/validation_error"
require_relative "anyvali/validation_context"
require_relative "anyvali/anyvali_document"
require_relative "anyvali/schema"

require_relative "anyvali/parse/coercion"
require_relative "anyvali/parse/coercion_config"
require_relative "anyvali/parse/defaults"

require_relative "anyvali/format/validators"

require_relative "anyvali/schemas/string_schema"
require_relative "anyvali/schemas/number_schema"
require_relative "anyvali/schemas/int_schema"
require_relative "anyvali/schemas/bool_schema"
require_relative "anyvali/schemas/null_schema"
require_relative "anyvali/schemas/any_schema"
require_relative "anyvali/schemas/unknown_schema"
require_relative "anyvali/schemas/never_schema"
require_relative "anyvali/schemas/literal_schema"
require_relative "anyvali/schemas/enum_schema"
require_relative "anyvali/schemas/array_schema"
require_relative "anyvali/schemas/tuple_schema"
require_relative "anyvali/schemas/object_schema"
require_relative "anyvali/schemas/record_schema"
require_relative "anyvali/schemas/union_schema"
require_relative "anyvali/schemas/intersection_schema"
require_relative "anyvali/schemas/optional_schema"
require_relative "anyvali/schemas/nullable_schema"
require_relative "anyvali/schemas/ref_schema"

require_relative "anyvali/interchange/exporter"
require_relative "anyvali/interchange/importer"

module AnyVali
  VERSION = "0.0.1"

  module_function

  # Builder methods

  def string
    StringSchema.new
  end

  def number
    NumberSchema.new(kind: "number")
  end

  def float32
    NumberSchema.new(kind: "float32")
  end

  def float64
    NumberSchema.new(kind: "float64")
  end

  def int_
    IntSchema.new(kind: "int")
  end

  def int8
    IntSchema.new(kind: "int8")
  end

  def int16
    IntSchema.new(kind: "int16")
  end

  def int32
    IntSchema.new(kind: "int32")
  end

  def int64
    IntSchema.new(kind: "int64")
  end

  def uint8
    IntSchema.new(kind: "uint8")
  end

  def uint16
    IntSchema.new(kind: "uint16")
  end

  def uint32
    IntSchema.new(kind: "uint32")
  end

  def uint64
    IntSchema.new(kind: "uint64")
  end

  def bool
    BoolSchema.new
  end

  def null
    NullSchema.new
  end

  def any
    AnySchema.new
  end

  def unknown
    UnknownSchema.new
  end

  def never
    NeverSchema.new
  end

  def literal(value)
    LiteralSchema.new(value: value)
  end

  def enum_(*values)
    EnumSchema.new(values: values.flatten)
  end

  def array(items)
    ArraySchema.new(items: items)
  end

  def tuple(*elements)
    TupleSchema.new(elements: elements.flatten)
  end

  def object(properties:, required: [], unknown_keys: nil)
    ObjectSchema.new(
      properties: properties,
      required: required,
      unknown_keys: unknown_keys || "strip",
      unknown_keys_explicit: !unknown_keys.nil?
    )
  end

  def record(values)
    RecordSchema.new(values: values)
  end

  def union(*variants)
    UnionSchema.new(variants: variants.flatten)
  end

  def intersection(*schemas)
    IntersectionSchema.new(all_of: schemas.flatten)
  end

  def optional(schema)
    OptionalSchema.new(schema: schema)
  end

  def nullable(schema)
    NullableSchema.new(schema: schema)
  end

  def ref(ref_path)
    RefSchema.new(ref: ref_path)
  end

  def safe_parse_encrypted(schema, data)
    context = ValidationContext.new
    context.sensitive_mode = :encrypted
    schema.safe_parse(data, context: context)
  end

  def encrypt(schema, data, transform)
    encrypted = transform_sensitive(schema, schema.parse(data), :encrypt, transform)
    result = safe_parse_encrypted(schema, encrypted)
    raise ValidationError, result.issues if result.failure?
    result.value
  end

  def decrypt(schema, data, transform)
    encrypted = safe_parse_encrypted(schema, data)
    raise ValidationError, encrypted.issues if encrypted.failure?
    schema.parse(transform_sensitive(schema, encrypted.value, :decrypt, transform))
  end

  def transform_sensitive(schema, data, mode, transform)
    context = ValidationContext.new
    context.sensitive_mode = mode
    context.sensitive_transform = transform
    context.sensitive_cache = {}
    result = schema.safe_parse(data, context: context)
    raise ValidationError, result.issues if result.failure?
    result.value
  end
  private_class_method :transform_sensitive

  # Interchange

  def export(schema, mode: :portable, definitions: {})
    Interchange::Exporter.export(schema, mode: mode, definitions: definitions)
  end

  def import(doc)
    Interchange::Importer.import(doc)
  end

  def import_schema(doc)
    Interchange::Importer.import_schema(doc)
  end
end
