# frozen_string_literal: true

require_relative "test_helper"

class TestCoercion < Minitest::Test
  def test_string_to_int
    s = AnyVali.int_.coerce("string->int")
    assert_equal 42, s.parse("42")
  end

  def test_string_to_int_trims_whitespace
    s = AnyVali.int_.coerce("string->int")
    assert_equal 42, s.parse("  42  ")
  end

  def test_string_to_number
    s = AnyVali.number.coerce("string->number")
    assert_equal 3.14, s.parse("3.14")
  end

  def test_string_to_bool_true
    s = AnyVali.bool.coerce("string->bool")
    assert_equal true, s.parse("true")
  end

  def test_string_to_bool_false
    s = AnyVali.bool.coerce("string->bool")
    assert_equal false, s.parse("false")
  end

  def test_string_to_bool_1
    s = AnyVali.bool.coerce("string->bool")
    assert_equal true, s.parse("1")
  end

  def test_string_to_bool_0
    s = AnyVali.bool.coerce("string->bool")
    assert_equal false, s.parse("0")
  end

  def test_string_to_bool_case_insensitive
    s = AnyVali.bool.coerce("string->bool")
    assert_equal true, s.parse("TRUE")
  end

  def test_trim
    s = AnyVali.string.coerce("trim")
    assert_equal "hello", s.parse("  hello  ")
  end

  def test_lower
    s = AnyVali.string.coerce("lower")
    assert_equal "hello world", s.parse("HELLO World")
  end

  def test_upper
    s = AnyVali.string.coerce("upper")
    assert_equal "HELLO WORLD", s.parse("hello world")
  end

  def test_coercion_failure
    s = AnyVali.int_.coerce("string->int")
    result = s.safe_parse("not-a-number")
    assert result.failure?
    assert_equal "coercion_failed", result.issues.first.code
    assert_equal "int", result.issues.first.expected
    assert_equal "not-a-number", result.issues.first.received
  end

  def test_coercion_then_validation_fails
    s = AnyVali.int_.min(10).coerce("string->int")
    result = s.safe_parse("5")
    assert result.failure?
    assert_equal "too_small", result.issues.first.code
    assert_equal "10", result.issues.first.expected
    assert_equal "5", result.issues.first.received
  end

  def test_coercion_then_validation_success
    s = AnyVali.int_.min(1).max(100).coerce("string->int")
    assert_equal 50, s.parse("50")
  end

  def test_chained_coercions
    s = AnyVali.string.coerce(["trim", "lower"])
    assert_equal "hello", s.parse("  HELLO  ")
  end

  def test_int_passthrough_no_coerce
    s = AnyVali.int_.coerce("string->int")
    assert_equal 42, s.parse(42)
  end

  def test_coercion_config_portable
    assert AnyVali::CoercionConfig.portable?("trim")
    assert AnyVali::CoercionConfig.portable?("string->int")
    assert AnyVali::CoercionConfig.portable?(["trim", "lower"])
  end
end
