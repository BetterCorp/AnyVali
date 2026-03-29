# frozen_string_literal: true

require_relative "test_helper"
require "json"

class TestInterchange < Minitest::Test
  def test_export_string_schema
    s = AnyVali.string.min_length(1).max_length(100)
    doc = AnyVali.export(s)
    assert_equal "1.0", doc["anyvaliVersion"]
    assert_equal "1", doc["schemaVersion"]
    assert_equal "string", doc["root"]["kind"]
    assert_equal 1, doc["root"]["minLength"]
    assert_equal 100, doc["root"]["maxLength"]
  end

  def test_export_number_schema
    s = AnyVali.number.min(0).max(100)
    doc = AnyVali.export(s)
    assert_equal "number", doc["root"]["kind"]
    assert_equal 0, doc["root"]["min"]
    assert_equal 100, doc["root"]["max"]
  end

  def test_export_object_schema
    s = AnyVali.object(
      properties: { "name" => AnyVali.string, "age" => AnyVali.int_ },
      required: ["name", "age"]
    )
    doc = AnyVali.export(s)
    assert_equal "object", doc["root"]["kind"]
    assert_equal "string", doc["root"]["properties"]["name"]["kind"]
    assert_equal "int", doc["root"]["properties"]["age"]["kind"]
    assert_equal ["name", "age"], doc["root"]["required"]
  end

  def test_export_portable_fails_with_custom
    s = AnyVali.string.refine { |v, p| nil }
    assert_raises(AnyVali::ValidationError) { AnyVali.export(s) }
  end

  def test_export_extended_with_custom
    s = AnyVali.string.refine { |v, p| nil }
    doc = AnyVali.export(s, mode: :extended)
    assert_equal "string", doc["root"]["kind"]
  end

  def test_import_string_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "string", "minLength" => 3 },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal "hello", schema.parse("hello")
    result = schema.safe_parse("ab")
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
  end

  def test_import_number_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "number", "min" => 10 },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal 15, schema.parse(15)
    result = schema.safe_parse(5)
    assert result.failure?
  end

  def test_import_int_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "int", "min" => 1, "max" => 10 },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal 5, schema.parse(5)
  end

  def test_import_bool_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "bool" },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal true, schema.parse(true)
  end

  def test_import_null_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "null" },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_nil schema.parse(nil)
  end

  def test_import_object_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "object",
        "properties" => {
          "name" => { "kind" => "string" },
          "age" => { "kind" => "int" }
        },
        "required" => ["name", "age"],
        "unknownKeys" => "reject"
      },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    result = schema.parse({ "name" => "Alice", "age" => 30 })
    assert_equal({ "name" => "Alice", "age" => 30 }, result)
  end

  def test_import_array_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "array", "items" => { "kind" => "string" } },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal ["a", "b"], schema.parse(["a", "b"])
  end

  def test_import_tuple_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "tuple",
        "elements" => [{ "kind" => "string" }, { "kind" => "int" }]
      },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal ["hello", 42], schema.parse(["hello", 42])
  end

  def test_import_union_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "union",
        "variants" => [{ "kind" => "string" }, { "kind" => "int" }]
      },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal "hello", schema.parse("hello")
    assert_equal 42, schema.parse(42)
  end

  def test_import_record_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "record", "values" => { "kind" => "int" } },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal({ "a" => 1, "b" => 2 }, schema.parse({ "a" => 1, "b" => 2 }))
  end

  def test_import_nullable_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "nullable", "schema" => { "kind" => "string" } },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_nil schema.parse(nil)
    assert_equal "hello", schema.parse("hello")
  end

  def test_import_optional_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "object",
        "properties" => {
          "name" => { "kind" => "optional", "schema" => { "kind" => "string" } }
        },
        "required" => [],
        "unknownKeys" => "reject"
      },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal({}, schema.parse({}))
    assert_equal({ "name" => "Alice" }, schema.parse({ "name" => "Alice" }))
  end

  def test_import_literal_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "literal", "value" => "hello" },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal "hello", schema.parse("hello")
  end

  def test_import_enum_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "enum", "values" => ["red", "green", "blue"] },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal "red", schema.parse("red")
  end

  def test_import_ref_schema
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "object",
        "properties" => {
          "user" => { "kind" => "ref", "ref" => "#/definitions/User" }
        },
        "required" => ["user"],
        "unknownKeys" => "reject"
      },
      "definitions" => {
        "User" => {
          "kind" => "object",
          "properties" => {
            "name" => { "kind" => "string" },
            "age" => { "kind" => "int" }
          },
          "required" => ["name", "age"],
          "unknownKeys" => "reject"
        }
      },
      "extensions" => {}
    }
    schema, context, _defs = AnyVali.import(doc)
    result = schema.safe_parse({ "user" => { "name" => "Alice", "age" => 30 } }, context: context)
    assert result.success?
    assert_equal({ "user" => { "name" => "Alice", "age" => 30 } }, result.value)
  end

  def test_import_from_json_string
    json = '{"anyvaliVersion":"1.0","schemaVersion":"1","root":{"kind":"string"},"definitions":{},"extensions":{}}'
    schema = AnyVali.import_schema(json)
    assert_equal "hello", schema.parse("hello")
  end

  def test_roundtrip_export_import
    original = AnyVali.object(
      properties: {
        "name" => AnyVali.string.min_length(1),
        "age" => AnyVali.int_.min(0)
      },
      required: ["name", "age"]
    )
    doc = AnyVali.export(original)
    reimported = AnyVali.import_schema(doc)
    assert_equal({ "name" => "Alice", "age" => 30 }, reimported.parse({ "name" => "Alice", "age" => 30 }))
  end

  def test_import_coercion
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "int", "coerce" => "string->int" },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal 42, schema.parse("42")
  end

  def test_import_chained_coercion
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "string", "coerce" => ["trim", "lower"] },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal "hello", schema.parse("  HELLO  ")
  end

  def test_import_intersection
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "intersection",
        "allOf" => [
          { "kind" => "number", "min" => 0 },
          { "kind" => "number", "max" => 100 }
        ]
      },
      "definitions" => {},
      "extensions" => {}
    }
    schema = AnyVali.import_schema(doc)
    assert_equal 50, schema.parse(50)
  end

  def test_import_unsupported_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "custom_thing" },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_width_types
    %w[int8 int16 int32 int64 uint8 uint16 uint32 uint64 float32 float64].each do |kind|
      doc = {
        "anyvaliVersion" => "1.0",
        "schemaVersion" => "1",
        "root" => { "kind" => kind },
        "definitions" => {},
        "extensions" => {}
      }
      schema = AnyVali.import_schema(doc)
      assert_equal kind, schema.kind
    end
  end

  def test_import_any_unknown_never
    %w[any unknown never].each do |kind|
      doc = {
        "anyvaliVersion" => "1.0",
        "schemaVersion" => "1",
        "root" => { "kind" => kind },
        "definitions" => {},
        "extensions" => {}
      }
      schema = AnyVali.import_schema(doc)
      assert_equal kind, schema.kind
    end
  end
end
