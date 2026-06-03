# frozen_string_literal: true

require_relative "test_helper"

class TestDefaults < Minitest::Test
  def test_missing_field_gets_default
    schema = AnyVali.object(
      properties: {
        "name" => AnyVali.string,
        "role" => AnyVali.string.default("user")
      },
      required: ["name"]
    )

    result = schema.safe_parse({ "name" => "Alice" })

    assert result.success?, result.issues.map(&:to_h).inspect
    assert_equal "user", result.value["role"]
  end

  def test_present_field_is_not_overwritten
    schema = AnyVali.object(
      properties: {
        "role" => AnyVali.string.default("user")
      }
    )

    result = schema.safe_parse({ "role" => "admin" })

    assert result.success?, result.issues.map(&:to_h).inspect
    assert_equal "admin", result.value["role"]
  end

  def test_invalid_default_produces_default_invalid
    schema = AnyVali.object(
      properties: {
        "count" => AnyVali.int_.min(10).default(5)
      }
    )

    result = schema.safe_parse({})

    assert result.failure?
    assert_equal "default_invalid", result.issues.first.code
    assert_equal ["count"], result.issues.first.path
  end

  def test_null_is_not_absent_for_nullable_default
    schema = AnyVali.object(
      properties: {
        "value" => AnyVali.nullable(AnyVali.string).default("fallback")
      }
    )

    result = schema.safe_parse({ "value" => nil })

    assert result.success?, result.issues.map(&:to_h).inspect
    assert_nil result.value["value"]
  end

  def test_falsy_defaults_are_applied
    schema = AnyVali.object(
      properties: {
        "count" => AnyVali.int_.default(0),
        "name" => AnyVali.string.default(""),
        "active" => AnyVali.bool.default(false)
      }
    )

    result = schema.safe_parse({})

    assert result.success?, result.issues.map(&:to_h).inspect
    assert_equal({ "count" => 0, "name" => "", "active" => false }, result.value)
  end

  def test_nested_object_field_gets_default
    schema = AnyVali.object(
      properties: {
        "user" => AnyVali.object(
          properties: {
            "name" => AnyVali.string,
            "role" => AnyVali.string.default("guest")
          },
          required: ["name"]
        )
      },
      required: ["user"]
    )

    result = schema.safe_parse({ "user" => { "name" => "Bob" } })

    assert result.success?, result.issues.map(&:to_h).inspect
    assert_equal "guest", result.value["user"]["role"]
  end
end
