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

  # --- Ergonomic no-arg / generic-"string" coercion -------------------------
  # The only portable coercion source is "string" (spec 5.1). Enabling coercion
  # on a numeric/bool schema with no explicit typed target (`.coerce` no-arg, or
  # the generic "string" source) infers the target from the schema kind. A bare
  # `.coerce` defaults the source to "string", so `number.coerce` and
  # `number.coerce("string")` are equivalent.

  def test_no_arg_coerce_equivalent_to_string
    no_arg = AnyVali.number.coerce
    explicit = AnyVali.number.coerce("string")
    assert_equal "string", no_arg.coerce_config
    assert_equal "string", explicit.coerce_config
    assert_equal 3.14, no_arg.parse("3.14")
  end

  def test_default_coerce_number_from_string
    s = AnyVali.number.coerce
    assert_equal 3.14, s.parse("3.14")
  end

  def test_default_coerce_int_from_string
    s = AnyVali.int_.coerce
    assert_equal 42, s.parse("42")
  end

  def test_default_coerce_bool_true_from_string
    s = AnyVali.bool.coerce
    assert_equal true, s.parse("true")
  end

  def test_default_coerce_bool_false_from_string
    s = AnyVali.bool.coerce
    assert_equal false, s.parse("false")
  end

  def test_default_coerce_object_numeric_fields_from_string
    s = AnyVali.object(
      properties: {
        "lumpSum" => AnyVali.number.coerce,
        "monthlyContributions" => AnyVali.number.coerce,
        "investmentTerm" => AnyVali.number.coerce
      },
      required: %w[lumpSum monthlyContributions investmentTerm]
    )

    result = s.safe_parse(
      "lumpSum" => "1000000",
      "monthlyContributions" => "1000",
      "investmentTerm" => "20"
    )

    assert result.success?, "object with default-coerce numeric fields must parse"
    assert_equal 1_000_000.0, result.value["lumpSum"]
    assert_equal 1_000.0, result.value["monthlyContributions"]
    assert_equal 20.0, result.value["investmentTerm"]
  end

  def test_no_arg_coerce_is_portable
    assert AnyVali::CoercionConfig.portable?("string")
    assert AnyVali.number.coerce.portable?
  end

  # --- Canonical coercion matrix (all FROM STRING, no-arg ergonomic) --------
  # Targets inferred from schema kind. Every ACCEPT/REJECT row from the spec.

  INT_ACCEPT = { "42" => 42, "  42  " => 42, "-7" => -7 }.freeze
  INT_REJECT = ["3.14", "0x10", "1_000", "+5", "Infinity", "", "abc"].freeze

  def test_matrix_string_to_int_accept
    s = AnyVali.int_.coerce
    INT_ACCEPT.each do |input, expected|
      result = s.safe_parse(input)
      assert result.success?, "int coerce must accept #{input.inspect}"
      assert_equal expected, result.value
    end
  end

  def test_matrix_string_to_int_reject
    s = AnyVali.int_.coerce
    INT_REJECT.each do |input|
      result = s.safe_parse(input)
      assert result.failure?, "int coerce must reject #{input.inspect}"
      assert_equal "coercion_failed", result.issues.first.code
    end
  end

  NUMBER_ACCEPT = { "3.14" => 3.14, "-1.5e3" => -1500.0, "  2  " => 2.0, "0" => 0.0 }.freeze
  NUMBER_REJECT = ["0x10", "Infinity", "NaN", "", "1_000", "abc"].freeze

  def test_matrix_string_to_number_accept
    s = AnyVali.number.coerce
    NUMBER_ACCEPT.each do |input, expected|
      result = s.safe_parse(input)
      assert result.success?, "number coerce must accept #{input.inspect}"
      assert_equal expected, result.value
    end
  end

  def test_matrix_string_to_number_reject
    s = AnyVali.number.coerce
    NUMBER_REJECT.each do |input|
      result = s.safe_parse(input)
      assert result.failure?, "number coerce must reject #{input.inspect}"
      assert_equal "coercion_failed", result.issues.first.code
    end
  end

  BOOL_TRUE = %w[true TRUE 1].freeze
  BOOL_FALSE = %w[false 0].freeze
  BOOL_REJECT = ["yes", "no", "on", "off", "t", "f", "2", ""].freeze

  def test_matrix_string_to_bool_true
    s = AnyVali.bool.coerce
    BOOL_TRUE.each do |input|
      result = s.safe_parse(input)
      assert result.success?, "bool coerce must accept #{input.inspect}"
      assert_equal true, result.value
    end
  end

  def test_matrix_string_to_bool_false
    s = AnyVali.bool.coerce
    BOOL_FALSE.each do |input|
      result = s.safe_parse(input)
      assert result.success?, "bool coerce must accept #{input.inspect}"
      assert_equal false, result.value
    end
  end

  def test_matrix_string_to_bool_reject
    s = AnyVali.bool.coerce
    BOOL_REJECT.each do |input|
      result = s.safe_parse(input)
      assert result.failure?, "bool coerce must reject #{input.inspect}"
      assert_equal "coercion_failed", result.issues.first.code
    end
  end

  # --- String-kind transforms (string source is a no-op; transforms apply) ---

  def test_string_kind_no_arg_coerce_is_noop_passthrough
    s = AnyVali.string.coerce
    assert_equal "hello", s.parse("hello")
  end

  def test_string_transforms_trim
    s = AnyVali.string.coerce("trim")
    assert_equal "hello", s.parse("  hello  ")
  end

  def test_string_transforms_chainable
    s = AnyVali.string.coerce(%w[trim lower upper])
    assert_equal "HELLO", s.parse("  Hello  ")
  end

  # CWE-20 / spec 5.1: non-portable coercion bypass. Ruby's Float() accepts
  # digit-group underscores ("1_000.5") and hex floats, which diverge from the
  # JS reference. string->number must accept ASCII decimals only.
  def test_string_to_number_rejects_non_decimal
    s = AnyVali.number.coerce("string->number")
    ["1_000.5", "0x1.8p3", "0x10"].each do |bad|
      assert s.safe_parse(bad).failure?, "string->number must reject #{bad.inspect}"
    end
    [["3.14", 3.14], ["+5", 5.0], [".5", 0.5], ["1e3", 1000.0]].each do |good, expected|
      result = s.safe_parse(good)
      assert result.success?, "string->number must accept #{good.inspect}"
      assert_equal expected, result.value
    end
  end
end
