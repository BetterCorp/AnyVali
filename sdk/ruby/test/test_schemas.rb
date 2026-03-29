# frozen_string_literal: true

require_relative "test_helper"

class TestStringSchema < Minitest::Test
  def test_accepts_simple_string
    s = AnyVali.string
    assert_equal "hello", s.parse("hello")
  end

  def test_accepts_empty_string
    s = AnyVali.string
    assert_equal "", s.parse("")
  end

  def test_accepts_unicode_string
    s = AnyVali.string
    assert_equal "\u00e9\u00e0\u00fc\u00f1\u00f6", s.parse("\u00e9\u00e0\u00fc\u00f1\u00f6")
  end

  def test_rejects_number
    s = AnyVali.string
    result = s.safe_parse(42)
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "string", result.issues.first.expected
    assert_equal "number", result.issues.first.received
  end

  def test_rejects_boolean
    s = AnyVali.string
    result = s.safe_parse(true)
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
  end

  def test_rejects_null
    s = AnyVali.string
    result = s.safe_parse(nil)
    assert result.failure?
    assert_equal "null", result.issues.first.received
  end

  def test_rejects_array
    s = AnyVali.string
    result = s.safe_parse(["a", "b"])
    assert result.failure?
    assert_equal "array", result.issues.first.received
  end

  def test_rejects_object
    s = AnyVali.string
    result = s.safe_parse({ "key" => "value" })
    assert result.failure?
    assert_equal "object", result.issues.first.received
  end

  def test_min_length_passes
    s = AnyVali.string.min_length(3)
    assert_equal "abc", s.parse("abc")
  end

  def test_min_length_fails
    s = AnyVali.string.min_length(3)
    result = s.safe_parse("ab")
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
    assert_equal "3", result.issues.first.expected
    assert_equal "2", result.issues.first.received
  end

  def test_max_length_passes
    s = AnyVali.string.max_length(5)
    assert_equal "hello", s.parse("hello")
  end

  def test_max_length_fails
    s = AnyVali.string.max_length(5)
    result = s.safe_parse("hello!")
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end

  def test_pattern_passes
    s = AnyVali.string.pattern("^[a-z]+$")
    assert_equal "abc", s.parse("abc")
  end

  def test_pattern_fails
    s = AnyVali.string.pattern("^[a-z]+$")
    result = s.safe_parse("ABC")
    assert result.failure?
    assert_equal "invalid_string", result.issues.first.code
  end

  def test_starts_with_passes
    s = AnyVali.string.starts_with("hello")
    assert_equal "hello world", s.parse("hello world")
  end

  def test_starts_with_fails
    s = AnyVali.string.starts_with("hello")
    result = s.safe_parse("world hello")
    assert result.failure?
    assert_equal "invalid_string", result.issues.first.code
  end

  def test_ends_with_passes
    s = AnyVali.string.ends_with(".json")
    assert_equal "file.json", s.parse("file.json")
  end

  def test_ends_with_fails
    s = AnyVali.string.ends_with(".json")
    result = s.safe_parse("file.xml")
    assert result.failure?
  end

  def test_includes_passes
    s = AnyVali.string.includes("world")
    assert_equal "hello world!", s.parse("hello world!")
  end

  def test_includes_fails
    s = AnyVali.string.includes("world")
    result = s.safe_parse("hello there")
    assert result.failure?
  end

  def test_parse_raises_on_invalid
    s = AnyVali.string
    assert_raises(AnyVali::ValidationError) { s.parse(42) }
  end

  def test_pattern_with_regexp
    s = AnyVali.string.pattern(/^[a-z]+$/)
    assert_equal "abc", s.parse("abc")
  end

  def test_method_chaining
    s = AnyVali.string.min_length(1).max_length(10).pattern("^[a-z]+$")
    assert_equal "hello", s.parse("hello")
  end
end

class TestNumberSchema < Minitest::Test
  def test_accepts_positive_integer
    s = AnyVali.number
    assert_equal 42, s.parse(42)
  end

  def test_accepts_zero
    s = AnyVali.number
    assert_equal 0, s.parse(0)
  end

  def test_accepts_negative_float
    s = AnyVali.number
    assert_equal(-3.14, s.parse(-3.14))
  end

  def test_accepts_large_float
    s = AnyVali.number
    assert_equal 1.7976931348623157e+308, s.parse(1.7976931348623157e+308)
  end

  def test_rejects_string
    s = AnyVali.number
    result = s.safe_parse("42")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "string", result.issues.first.received
  end

  def test_rejects_boolean
    s = AnyVali.number
    result = s.safe_parse(true)
    assert result.failure?
  end

  def test_rejects_null
    s = AnyVali.number
    result = s.safe_parse(nil)
    assert result.failure?
  end

  def test_rejects_object
    s = AnyVali.number
    result = s.safe_parse({})
    assert result.failure?
  end

  def test_rejects_array
    s = AnyVali.number
    result = s.safe_parse([1, 2, 3])
    assert result.failure?
  end

  def test_min_passes
    s = AnyVali.number.min(10)
    assert_equal 10, s.parse(10)
  end

  def test_min_fails
    s = AnyVali.number.min(10)
    result = s.safe_parse(9)
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
    assert_equal "10", result.issues.first.expected
    assert_equal "9", result.issues.first.received
  end

  def test_max_passes
    s = AnyVali.number.max(100)
    assert_equal 100, s.parse(100)
  end

  def test_max_fails
    s = AnyVali.number.max(100)
    result = s.safe_parse(101)
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end

  def test_exclusive_min_passes
    s = AnyVali.number.exclusive_min(0)
    assert_equal 0.001, s.parse(0.001)
  end

  def test_exclusive_min_fails
    s = AnyVali.number.exclusive_min(0)
    result = s.safe_parse(0)
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
  end

  def test_exclusive_max_passes
    s = AnyVali.number.exclusive_max(100)
    assert_equal 99.999, s.parse(99.999)
  end

  def test_exclusive_max_fails
    s = AnyVali.number.exclusive_max(100)
    result = s.safe_parse(100)
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end

  def test_multiple_of_passes
    s = AnyVali.number.multiple_of(3)
    assert_equal 9, s.parse(9)
  end

  def test_multiple_of_fails
    s = AnyVali.number.multiple_of(3)
    result = s.safe_parse(10)
    assert result.failure?
    assert_equal "invalid_number", result.issues.first.code
  end

  def test_multiple_of_float
    s = AnyVali.number.multiple_of(0.5)
    assert_equal 2.5, s.parse(2.5)
  end
end

class TestIntSchema < Minitest::Test
  def test_accepts_positive_integer
    s = AnyVali.int_
    assert_equal 42, s.parse(42)
  end

  def test_accepts_zero
    s = AnyVali.int_
    assert_equal 0, s.parse(0)
  end

  def test_accepts_negative_integer
    s = AnyVali.int_
    assert_equal(-100, s.parse(-100))
  end

  def test_rejects_float
    s = AnyVali.int_
    result = s.safe_parse(3.14)
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "int", result.issues.first.expected
    assert_equal "number", result.issues.first.received
  end

  def test_rejects_string
    s = AnyVali.int_
    result = s.safe_parse("42")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
  end

  def test_int_with_constraints
    s = AnyVali.int_.min(1).max(10)
    assert_equal 5, s.parse(5)
  end

  def test_large_int64
    s = AnyVali.int_
    assert_equal 9007199254740991, s.parse(9007199254740991)
  end
end

class TestIntWidths < Minitest::Test
  def test_int8_accepts_max
    assert_equal 127, AnyVali.int8.parse(127)
  end

  def test_int8_accepts_min
    assert_equal(-128, AnyVali.int8.parse(-128))
  end

  def test_int8_rejects_above
    result = AnyVali.int8.safe_parse(128)
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
    assert_equal "int8", result.issues.first.expected
  end

  def test_int8_rejects_below
    result = AnyVali.int8.safe_parse(-129)
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
  end

  def test_int16_accepts
    assert_equal 32767, AnyVali.int16.parse(32767)
  end

  def test_int16_rejects_above
    result = AnyVali.int16.safe_parse(32768)
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end

  def test_int32_accepts_max
    assert_equal 2147483647, AnyVali.int32.parse(2147483647)
  end

  def test_int32_rejects_above
    result = AnyVali.int32.safe_parse(2147483648)
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end

  def test_uint8_accepts_zero
    assert_equal 0, AnyVali.uint8.parse(0)
  end

  def test_uint8_accepts_255
    assert_equal 255, AnyVali.uint8.parse(255)
  end

  def test_uint8_rejects_negative
    result = AnyVali.uint8.safe_parse(-1)
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
  end

  def test_uint8_rejects_256
    result = AnyVali.uint8.safe_parse(256)
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end

  def test_uint16_accepts_max
    assert_equal 65535, AnyVali.uint16.parse(65535)
  end

  def test_uint32_accepts_max
    assert_equal 4294967295, AnyVali.uint32.parse(4294967295)
  end

  def test_uint64_rejects_negative
    result = AnyVali.uint64.safe_parse(-1)
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
  end

  def test_int64_accepts
    assert_equal 42, AnyVali.int64.parse(42)
  end
end

class TestFloatWidths < Minitest::Test
  def test_float64_accepts_float
    s = AnyVali.float64
    assert_equal 3.141592653589793, s.parse(3.141592653589793)
  end

  def test_float64_accepts_integer
    s = AnyVali.float64
    assert_equal 42, s.parse(42)
  end

  def test_float64_rejects_string
    result = AnyVali.float64.safe_parse("3.14")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
  end

  def test_float32_accepts_valid
    assert_equal 1.5, AnyVali.float32.parse(1.5)
  end

  def test_float32_rejects_boolean
    result = AnyVali.float32.safe_parse(true)
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
  end
end

class TestBoolSchema < Minitest::Test
  def test_accepts_true
    assert_equal true, AnyVali.bool.parse(true)
  end

  def test_accepts_false
    assert_equal false, AnyVali.bool.parse(false)
  end

  def test_rejects_number
    result = AnyVali.bool.safe_parse(1)
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "number", result.issues.first.received
  end

  def test_rejects_string
    result = AnyVali.bool.safe_parse("true")
    assert result.failure?
  end

  def test_rejects_null
    result = AnyVali.bool.safe_parse(nil)
    assert result.failure?
    assert_equal "null", result.issues.first.received
  end
end

class TestNullSchema < Minitest::Test
  def test_accepts_null
    assert_nil AnyVali.null.parse(nil)
  end

  def test_rejects_string
    result = AnyVali.null.safe_parse("null")
    assert result.failure?
    assert_equal "string", result.issues.first.received
  end

  def test_rejects_zero
    result = AnyVali.null.safe_parse(0)
    assert result.failure?
  end

  def test_rejects_false
    result = AnyVali.null.safe_parse(false)
    assert result.failure?
    assert_equal "boolean", result.issues.first.received
  end

  def test_rejects_empty_string
    result = AnyVali.null.safe_parse("")
    assert result.failure?
  end
end

class TestAnySchema < Minitest::Test
  def test_accepts_string
    assert_equal "hello", AnyVali.any.parse("hello")
  end

  def test_accepts_number
    assert_equal 42, AnyVali.any.parse(42)
  end

  def test_accepts_null
    assert_nil AnyVali.any.parse(nil)
  end

  def test_accepts_object
    assert_equal({ "key" => "value" }, AnyVali.any.parse({ "key" => "value" }))
  end

  def test_accepts_array
    assert_equal [1, "two", true], AnyVali.any.parse([1, "two", true])
  end
end

class TestUnknownSchema < Minitest::Test
  def test_accepts_string
    assert_equal "hello", AnyVali.unknown.parse("hello")
  end

  def test_accepts_number
    assert_equal 99, AnyVali.unknown.parse(99)
  end

  def test_accepts_null
    assert_nil AnyVali.unknown.parse(nil)
  end

  def test_accepts_boolean
    assert_equal false, AnyVali.unknown.parse(false)
  end

  def test_accepts_nested
    input = { "a" => [1, { "b" => true }] }
    assert_equal input, AnyVali.unknown.parse(input)
  end
end

class TestNeverSchema < Minitest::Test
  def test_rejects_string
    result = AnyVali.never.safe_parse("hello")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "never", result.issues.first.expected
    assert_equal "string", result.issues.first.received
  end

  def test_rejects_number
    result = AnyVali.never.safe_parse(0)
    assert result.failure?
  end

  def test_rejects_null
    result = AnyVali.never.safe_parse(nil)
    assert result.failure?
  end

  def test_rejects_boolean
    result = AnyVali.never.safe_parse(true)
    assert result.failure?
  end

  def test_rejects_object
    result = AnyVali.never.safe_parse({})
    assert result.failure?
  end
end

class TestLiteralSchema < Minitest::Test
  def test_accepts_matching_string
    s = AnyVali.literal("hello")
    assert_equal "hello", s.parse("hello")
  end

  def test_rejects_non_matching_string
    s = AnyVali.literal("hello")
    result = s.safe_parse("world")
    assert result.failure?
    assert_equal "invalid_literal", result.issues.first.code
  end

  def test_accepts_matching_number
    s = AnyVali.literal(42)
    assert_equal 42, s.parse(42)
  end

  def test_accepts_matching_boolean
    s = AnyVali.literal(true)
    assert_equal true, s.parse(true)
  end

  def test_rejects_wrong_type
    s = AnyVali.literal(42)
    result = s.safe_parse("42")
    assert result.failure?
    assert_equal "invalid_literal", result.issues.first.code
  end

  def test_accepts_null_literal
    s = AnyVali.literal(nil)
    assert_nil s.parse(nil)
  end
end

class TestEnumSchema < Minitest::Test
  def test_accepts_value_in_enum
    s = AnyVali.enum_("red", "green", "blue")
    assert_equal "red", s.parse("red")
  end

  def test_accepts_another_value
    s = AnyVali.enum_("red", "green", "blue")
    assert_equal "blue", s.parse("blue")
  end

  def test_rejects_value_not_in_enum
    s = AnyVali.enum_("red", "green", "blue")
    result = s.safe_parse("yellow")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "enum(red,green,blue)", result.issues.first.expected
  end

  def test_numeric_enum
    s = AnyVali.enum_(1, 2, 3)
    assert_equal 2, s.parse(2)
  end

  def test_rejects_wrong_type
    s = AnyVali.enum_(1, 2, 3)
    result = s.safe_parse("1")
    assert result.failure?
  end
end

class TestArraySchema < Minitest::Test
  def test_accepts_valid_array
    s = AnyVali.array(AnyVali.string)
    assert_equal ["a", "b", "c"], s.parse(["a", "b", "c"])
  end

  def test_accepts_empty_array
    s = AnyVali.array(AnyVali.int_)
    assert_equal [], s.parse([])
  end

  def test_rejects_non_array
    s = AnyVali.array(AnyVali.string)
    result = s.safe_parse("not an array")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "array", result.issues.first.expected
  end

  def test_rejects_invalid_element
    s = AnyVali.array(AnyVali.int_)
    result = s.safe_parse([1, 2, "three"])
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal [2], result.issues.first.path
  end

  def test_reports_multiple_invalid_elements
    s = AnyVali.array(AnyVali.bool)
    result = s.safe_parse([true, "yes", false, 1])
    assert result.failure?
    assert_equal 2, result.issues.length
    assert_equal [1], result.issues[0].path
    assert_equal [3], result.issues[1].path
  end

  def test_min_items_passes
    s = AnyVali.array(AnyVali.int_).min_items(2)
    assert_equal [1, 2], s.parse([1, 2])
  end

  def test_min_items_fails
    s = AnyVali.array(AnyVali.int_).min_items(2)
    result = s.safe_parse([1])
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
  end

  def test_max_items_passes
    s = AnyVali.array(AnyVali.string).max_items(3)
    assert_equal ["a", "b", "c"], s.parse(["a", "b", "c"])
  end

  def test_max_items_fails
    s = AnyVali.array(AnyVali.string).max_items(3)
    result = s.safe_parse(["a", "b", "c", "d"])
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end
end

class TestTupleSchema < Minitest::Test
  def test_accepts_valid_tuple
    s = AnyVali.tuple(AnyVali.string, AnyVali.int_)
    assert_equal ["hello", 42], s.parse(["hello", 42])
  end

  def test_rejects_too_few_elements
    s = AnyVali.tuple(AnyVali.string, AnyVali.int_)
    result = s.safe_parse(["hello"])
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
    assert_equal "2", result.issues.first.expected
    assert_equal "1", result.issues.first.received
  end

  def test_rejects_too_many_elements
    s = AnyVali.tuple(AnyVali.string, AnyVali.int_)
    result = s.safe_parse(["hello", 42, true])
    assert result.failure?
    assert_equal "too_large", result.issues.first.code
  end

  def test_rejects_wrong_element_types
    s = AnyVali.tuple(AnyVali.string, AnyVali.int_)
    result = s.safe_parse([42, "hello"])
    assert result.failure?
    assert_equal 2, result.issues.length
    assert_equal [0], result.issues[0].path
    assert_equal [1], result.issues[1].path
  end

  def test_rejects_non_array
    s = AnyVali.tuple(AnyVali.string)
    result = s.safe_parse("not a tuple")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "tuple", result.issues.first.expected
  end
end

class TestObjectSchema < Minitest::Test
  def test_accepts_valid_object
    s = AnyVali.object(
      properties: { "name" => AnyVali.string, "age" => AnyVali.int_ },
      required: ["name", "age"]
    )
    result = s.parse({ "name" => "Alice", "age" => 30 })
    assert_equal({ "name" => "Alice", "age" => 30 }, result)
  end

  def test_rejects_missing_required
    s = AnyVali.object(
      properties: { "name" => AnyVali.string, "age" => AnyVali.int_ },
      required: ["name", "age"]
    )
    result = s.safe_parse({ "name" => "Alice" })
    assert result.failure?
    assert_equal "required", result.issues.first.code
    assert_equal ["age"], result.issues.first.path
  end

  def test_rejects_missing_all_required
    s = AnyVali.object(
      properties: { "name" => AnyVali.string, "age" => AnyVali.int_ },
      required: ["name", "age"]
    )
    result = s.safe_parse({})
    assert result.failure?
    assert_equal 2, result.issues.length
  end

  def test_accepts_optional_absent
    s = AnyVali.object(
      properties: { "name" => AnyVali.string, "nickname" => AnyVali.string },
      required: ["name"]
    )
    result = s.parse({ "name" => "Alice" })
    assert_equal({ "name" => "Alice" }, result)
  end

  def test_rejects_non_object
    s = AnyVali.object(
      properties: { "name" => AnyVali.string },
      required: ["name"]
    )
    result = s.safe_parse("not an object")
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "object", result.issues.first.expected
  end

  def test_unknown_keys_reject_default
    s = AnyVali.object(
      properties: { "name" => AnyVali.string },
      required: ["name"]
    )
    result = s.safe_parse({ "name" => "Alice", "extra" => "value" })
    assert result.failure?
    assert_equal "unknown_key", result.issues.first.code
    assert_equal ["extra"], result.issues.first.path
  end

  def test_unknown_keys_strip
    s = AnyVali.object(
      properties: { "name" => AnyVali.string },
      required: ["name"],
      unknown_keys: "strip"
    )
    result = s.parse({ "name" => "Alice", "extra" => "value", "another" => 42 })
    assert_equal({ "name" => "Alice" }, result)
  end

  def test_unknown_keys_allow
    s = AnyVali.object(
      properties: { "name" => AnyVali.string },
      required: ["name"],
      unknown_keys: "allow"
    )
    result = s.parse({ "name" => "Alice", "extra" => "value" })
    assert_equal({ "name" => "Alice", "extra" => "value" }, result)
  end

  def test_unknown_keys_multiple_rejected
    s = AnyVali.object(
      properties: { "id" => AnyVali.int_ },
      required: ["id"],
      unknown_keys: "reject"
    )
    result = s.safe_parse({ "id" => 1, "foo" => "bar", "baz" => true })
    assert result.failure?
    assert_equal 2, result.issues.length
    codes = result.issues.map(&:code)
    assert codes.all? { |c| c == "unknown_key" }
  end
end

class TestRecordSchema < Minitest::Test
  def test_accepts_valid_record
    s = AnyVali.record(AnyVali.int_)
    assert_equal({ "a" => 1, "b" => 2, "c" => 3 }, s.parse({ "a" => 1, "b" => 2, "c" => 3 }))
  end

  def test_accepts_empty_record
    s = AnyVali.record(AnyVali.string)
    assert_equal({}, s.parse({}))
  end

  def test_rejects_invalid_value
    s = AnyVali.record(AnyVali.int_)
    result = s.safe_parse({ "a" => 1, "b" => "two" })
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal ["b"], result.issues.first.path
  end

  def test_rejects_non_object
    s = AnyVali.record(AnyVali.string)
    result = s.safe_parse([1, 2, 3])
    assert result.failure?
    assert_equal "record", result.issues.first.expected
  end
end

class TestUnionSchema < Minitest::Test
  def test_accepts_first_variant
    s = AnyVali.union(AnyVali.string, AnyVali.int_)
    assert_equal "hello", s.parse("hello")
  end

  def test_accepts_second_variant
    s = AnyVali.union(AnyVali.string, AnyVali.int_)
    assert_equal 42, s.parse(42)
  end

  def test_rejects_no_match
    s = AnyVali.union(AnyVali.string, AnyVali.int_)
    result = s.safe_parse(true)
    assert result.failure?
    assert_equal "invalid_union", result.issues.first.code
    assert_equal "string | int", result.issues.first.expected
    assert_equal "boolean", result.issues.first.received
  end

  def test_first_matching_wins
    s = AnyVali.union(AnyVali.number, AnyVali.int_)
    assert_equal 5, s.parse(5)
  end

  def test_union_with_null
    s = AnyVali.union(AnyVali.string, AnyVali.null)
    assert_nil s.parse(nil)
  end
end

class TestIntersectionSchema < Minitest::Test
  def test_accepts_satisfying_all
    s = AnyVali.intersection(
      AnyVali.object(
        properties: { "name" => AnyVali.string },
        required: ["name"],
        unknown_keys: "allow"
      ),
      AnyVali.object(
        properties: { "age" => AnyVali.int_ },
        required: ["age"],
        unknown_keys: "allow"
      )
    )
    result = s.parse({ "name" => "Alice", "age" => 30 })
    assert_equal({ "name" => "Alice", "age" => 30 }, result)
  end

  def test_rejects_missing_from_second
    s = AnyVali.intersection(
      AnyVali.object(
        properties: { "name" => AnyVali.string },
        required: ["name"],
        unknown_keys: "allow"
      ),
      AnyVali.object(
        properties: { "age" => AnyVali.int_ },
        required: ["age"],
        unknown_keys: "allow"
      )
    )
    result = s.safe_parse({ "name" => "Alice" })
    assert result.failure?
    assert_equal "required", result.issues.first.code
  end

  def test_numeric_intersection_passes
    s = AnyVali.intersection(
      AnyVali.number.min(0),
      AnyVali.number.max(100)
    )
    assert_equal 50, s.parse(50)
  end

  def test_numeric_intersection_fails
    s = AnyVali.intersection(
      AnyVali.number.min(0),
      AnyVali.number.max(100)
    )
    result = s.safe_parse(-5)
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
  end
end

class TestOptionalSchema < Minitest::Test
  def test_accepts_present_value
    s = AnyVali.object(
      properties: { "name" => AnyVali.optional(AnyVali.string) },
      required: []
    )
    result = s.parse({ "name" => "Alice" })
    assert_equal({ "name" => "Alice" }, result)
  end

  def test_accepts_absent
    s = AnyVali.object(
      properties: { "name" => AnyVali.optional(AnyVali.string) },
      required: []
    )
    result = s.parse({})
    assert_equal({}, result)
  end

  def test_rejects_invalid_present
    s = AnyVali.object(
      properties: { "name" => AnyVali.optional(AnyVali.string) },
      required: []
    )
    result = s.safe_parse({ "name" => 123 })
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal ["name"], result.issues.first.path
  end

  def test_null_not_treated_as_absent
    s = AnyVali.object(
      properties: { "name" => AnyVali.optional(AnyVali.string) },
      required: []
    )
    result = s.safe_parse({ "name" => nil })
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "null", result.issues.first.received
  end
end

class TestNullableSchema < Minitest::Test
  def test_accepts_null
    s = AnyVali.nullable(AnyVali.string)
    assert_nil s.parse(nil)
  end

  def test_accepts_valid_non_null
    s = AnyVali.nullable(AnyVali.string)
    assert_equal "hello", s.parse("hello")
  end

  def test_rejects_invalid_non_null
    s = AnyVali.nullable(AnyVali.string)
    result = s.safe_parse(42)
    assert result.failure?
    assert_equal "invalid_type", result.issues.first.code
    assert_equal "number", result.issues.first.received
  end

  def test_nullable_int_accepts_null
    s = AnyVali.nullable(AnyVali.int_)
    assert_nil s.parse(nil)
  end

  def test_nullable_int_accepts_valid
    s = AnyVali.nullable(AnyVali.int_)
    assert_equal 99, s.parse(99)
  end
end

class TestCustomValidators < Minitest::Test
  def test_refine_adds_custom_validation
    s = AnyVali.string.refine { |v, p| [AnyVali::ValidationIssue.new(code: "custom", path: p, expected: "x", received: v)] if v == "bad" }
    assert_equal "good", s.parse("good")
    result = s.safe_parse("bad")
    assert result.failure?
  end

  def test_not_portable
    s = AnyVali.string.refine { |v, p| nil }
    refute s.portable?
  end

  def test_portable_export_fails_with_custom
    s = AnyVali.string.refine { |v, p| nil }
    assert_raises(AnyVali::ValidationError) { s.export(mode: :portable) }
  end
end
