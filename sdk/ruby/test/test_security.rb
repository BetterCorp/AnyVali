# frozen_string_literal: true

require_relative "test_helper"

# --------------------------------------------------------------------------
# CVE-2016-4055: ReDoS -- Catastrophic backtracking patterns
# --------------------------------------------------------------------------
class TestReDoS < Minitest::Test
  # A pattern known to cause exponential backtracking on non-matching input.
  # The validator must either reject the pattern or finish quickly (<5 s).
  EVIL_PATTERN = "^(a+)+$"

  def test_redos_pattern_does_not_hang
    s = AnyVali.string.pattern(EVIL_PATTERN)
    payload = "a" * 30 + "!"

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(payload)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    refute result.success?
    assert elapsed < 5, "ReDoS: pattern took #{elapsed}s (limit 5s)"
  end

  def test_nested_quantifier_redos
    s = AnyVali.string.pattern("^(a|a)+$")
    payload = "a" * 30 + "!"

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(payload)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    refute result.success?
    assert elapsed < 5, "ReDoS: nested quantifier took #{elapsed}s (limit 5s)"
  end

  def test_polynomial_backtracking_redos
    s = AnyVali.string.pattern("^([a-zA-Z0-9._-]+)*@")
    payload = "a" * 50 + "!"

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(payload)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    refute result.success?
    assert elapsed < 5, "ReDoS: polynomial backtracking took #{elapsed}s (limit 5s)"
  end
end

# --------------------------------------------------------------------------
# CVE-2003-1564: Recursive $ref -- self-referencing / deeply nested schemas
# --------------------------------------------------------------------------
class TestRecursiveRef < Minitest::Test
  def test_self_referencing_ref_does_not_stack_overflow
    # A definition that references itself: Node -> Node -> Node -> ...
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "ref", "ref" => "#/definitions/Node" },
      "definitions" => {
        "Node" => {
          "kind" => "object",
          "properties" => {
            "value" => { "kind" => "string" },
            "child" => { "kind" => "ref", "ref" => "#/definitions/Node" }
          },
          "required" => ["value"],
          "unknownKeys" => "allow"
        }
      },
      "extensions" => {}
    }

    schema, context, _defs = AnyVali.import(doc)

    # Build a cyclic object that refers to itself
    input = { "value" => "root" }
    input["child"] = input

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = schema.safe_parse(input, context: context)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    # It must either fail or succeed, but not hang or crash
    assert elapsed < 5, "Recursive ref took #{elapsed}s (limit 5s)"
    assert [true, false].include?(result.success?)
  end

  def test_deeply_nested_ref_chain_does_not_stack_overflow
    # A -> B -> A  (mutual recursion)
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "ref", "ref" => "#/definitions/A" },
      "definitions" => {
        "A" => {
          "kind" => "object",
          "properties" => {
            "b" => { "kind" => "ref", "ref" => "#/definitions/B" }
          },
          "required" => [],
          "unknownKeys" => "allow"
        },
        "B" => {
          "kind" => "object",
          "properties" => {
            "a" => { "kind" => "ref", "ref" => "#/definitions/A" }
          },
          "required" => [],
          "unknownKeys" => "allow"
        }
      },
      "extensions" => {}
    }

    schema, context, _defs = AnyVali.import(doc)

    # Deeply nested but acyclic data
    input = { "b" => { "a" => { "b" => { "a" => {} } } } }

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = schema.safe_parse(input, context: context)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    assert elapsed < 5, "Deep ref chain took #{elapsed}s (limit 5s)"
    assert [true, false].include?(result.success?)
  end

  def test_unresolved_ref_fails_gracefully
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "ref", "ref" => "#/definitions/DoesNotExist" },
      "definitions" => {},
      "extensions" => {}
    }

    schema, context, _defs = AnyVali.import(doc)
    result = schema.safe_parse("anything", context: context)
    refute result.success?
  end
end

# --------------------------------------------------------------------------
# CWE-190: Integer overflow -- Int width boundary enforcement
# --------------------------------------------------------------------------
class TestIntegerOverflow < Minitest::Test
  INT_BOUNDS = {
    int8:   [-128, 127],
    int16:  [-32_768, 32_767],
    int32:  [-2_147_483_648, 2_147_483_647],
    int64:  [-9_223_372_036_854_775_808, 9_223_372_036_854_775_807],
    uint8:  [0, 255],
    uint16: [0, 65_535],
    uint32: [0, 4_294_967_295],
    uint64: [0, 18_446_744_073_709_551_615]
  }.freeze

  INT_BOUNDS.each do |kind, (lo, hi)|
    define_method(:"test_#{kind}_accepts_lower_bound") do
      s = AnyVali.public_send(kind)
      assert_equal lo, s.parse(lo)
    end

    define_method(:"test_#{kind}_accepts_upper_bound") do
      s = AnyVali.public_send(kind)
      assert_equal hi, s.parse(hi)
    end

    define_method(:"test_#{kind}_rejects_above_upper_bound") do
      s = AnyVali.public_send(kind)
      result = s.safe_parse(hi + 1)
      refute result.success?, "#{kind} must reject #{hi + 1}"
    end

    define_method(:"test_#{kind}_rejects_below_lower_bound") do
      s = AnyVali.public_send(kind)
      result = s.safe_parse(lo - 1)
      refute result.success?, "#{kind} must reject #{lo - 1}"
    end
  end

  def test_int_rejects_float
    s = AnyVali.int_
    result = s.safe_parse(3.14)
    refute result.success?
    assert_equal "invalid_type", result.issues.first.code
  end

  def test_int_rejects_very_large_number
    s = AnyVali.int_
    result = s.safe_parse(10**100)
    # Should either accept (Ruby has bignum) or reject, but not crash
    assert [true, false].include?(result.success?)
  end
end

# --------------------------------------------------------------------------
# CWE-20: NaN / Infinity rejection
# --------------------------------------------------------------------------
class TestNaNInfinity < Minitest::Test
  def test_number_rejects_nan
    s = AnyVali.number
    result = s.safe_parse(Float::NAN)
    refute result.success?, "Number schema must reject NaN"
  end

  def test_number_rejects_positive_infinity
    s = AnyVali.number
    result = s.safe_parse(Float::INFINITY)
    refute result.success?, "Number schema must reject +Infinity"
  end

  def test_number_rejects_negative_infinity
    s = AnyVali.number
    result = s.safe_parse(-Float::INFINITY)
    refute result.success?, "Number schema must reject -Infinity"
  end

  def test_float64_rejects_nan
    s = AnyVali.float64
    result = s.safe_parse(Float::NAN)
    refute result.success?, "Float64 schema must reject NaN"
  end

  def test_float64_rejects_positive_infinity
    s = AnyVali.float64
    result = s.safe_parse(Float::INFINITY)
    refute result.success?, "Float64 schema must reject +Infinity"
  end

  def test_float64_rejects_negative_infinity
    s = AnyVali.float64
    result = s.safe_parse(-Float::INFINITY)
    refute result.success?, "Float64 schema must reject -Infinity"
  end

  def test_float32_rejects_nan
    s = AnyVali.float32
    result = s.safe_parse(Float::NAN)
    refute result.success?, "Float32 schema must reject NaN"
  end

  def test_float32_rejects_positive_infinity
    s = AnyVali.float32
    result = s.safe_parse(Float::INFINITY)
    refute result.success?, "Float32 schema must reject +Infinity"
  end

  def test_float32_rejects_negative_infinity
    s = AnyVali.float32
    result = s.safe_parse(-Float::INFINITY)
    refute result.success?, "Float32 schema must reject -Infinity"
  end

  def test_int_rejects_nan
    s = AnyVali.int_
    result = s.safe_parse(Float::NAN)
    refute result.success?, "Int schema must reject NaN"
  end

  def test_int_rejects_infinity
    s = AnyVali.int_
    result = s.safe_parse(Float::INFINITY)
    refute result.success?, "Int schema must reject Infinity"
  end

  def test_nan_in_array_element_rejected
    s = AnyVali.array(AnyVali.number)
    result = s.safe_parse([1.0, Float::NAN, 3.0])
    refute result.success?, "Array of numbers must reject NaN element"
  end

  def test_infinity_in_object_property_rejected
    s = AnyVali.object(
      properties: { "value" => AnyVali.number },
      required: ["value"]
    )
    result = s.safe_parse({ "value" => Float::INFINITY })
    refute result.success?, "Object with number property must reject Infinity"
  end
end

# --------------------------------------------------------------------------
# CWE-20: Format bypass -- edge cases in email, url, ipv4 validation
# --------------------------------------------------------------------------
class TestFormatBypass < Minitest::Test
  # Email edge cases
  def test_tampered_email_format_name_not_silently_ignored
    s = AnyVali.string.format("email\0")
    result = s.safe_parse("not-an-email")
    refute result.success?, "tampered format name bypassed email validation"
  end

  def test_imported_tampered_email_format_name_not_unconstrained
    schema = AnyVali.import_schema(
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "string", "format" => "email\0" },
      "definitions" => {},
      "extensions" => {}
    )
    result = schema.safe_parse("not-an-email")
    refute result.success?, "imported tampered format name bypassed email validation"
  end

  def test_email_rejects_missing_at
    s = AnyVali.string.format("email")
    result = s.safe_parse("userexample.com")
    refute result.success?
  end

  def test_email_rejects_missing_domain
    s = AnyVali.string.format("email")
    result = s.safe_parse("user@")
    refute result.success?
  end

  def test_email_rejects_missing_local_part
    s = AnyVali.string.format("email")
    result = s.safe_parse("@example.com")
    refute result.success?
  end

  def test_email_rejects_double_at
    s = AnyVali.string.format("email")
    result = s.safe_parse("user@@example.com")
    refute result.success?
  end

  def test_email_rejects_spaces
    s = AnyVali.string.format("email")
    result = s.safe_parse("user @example.com")
    refute result.success?
  end

  def test_email_rejects_no_tld
    s = AnyVali.string.format("email")
    result = s.safe_parse("user@localhost")
    refute result.success?
  end

  def test_email_rejects_empty_string
    s = AnyVali.string.format("email")
    result = s.safe_parse("")
    refute result.success?
  end

  def test_email_accepts_valid_with_dots
    s = AnyVali.string.format("email")
    result = s.safe_parse("first.last@example.com")
    assert result.success?
  end

  def test_email_accepts_valid_with_plus
    s = AnyVali.string.format("email")
    result = s.safe_parse("user+tag@example.com")
    assert result.success?
  end

  # URL edge cases
  def test_url_rejects_empty_string
    s = AnyVali.string.format("url")
    result = s.safe_parse("")
    refute result.success?
  end

  def test_url_rejects_plain_text
    s = AnyVali.string.format("url")
    result = s.safe_parse("not a url")
    refute result.success?
  end

  def test_url_rejects_javascript_scheme
    s = AnyVali.string.format("url")
    result = s.safe_parse("javascript:alert(1)")
    refute result.success?
  end

  def test_url_rejects_data_scheme
    s = AnyVali.string.format("url")
    result = s.safe_parse("data:text/html,<script>alert(1)</script>")
    refute result.success?
  end

  def test_url_rejects_ftp_scheme
    s = AnyVali.string.format("url")
    result = s.safe_parse("ftp://files.example.com")
    refute result.success?
  end

  def test_url_accepts_https
    s = AnyVali.string.format("url")
    result = s.safe_parse("https://example.com")
    assert result.success?
  end

  def test_url_accepts_http_with_path
    s = AnyVali.string.format("url")
    result = s.safe_parse("http://example.com/path?query=1&other=2#frag")
    assert result.success?
  end

  # IPv4 edge cases
  def test_ipv4_rejects_empty_string
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("")
    refute result.success?
  end

  def test_ipv4_rejects_leading_zeros
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("192.168.01.1")
    refute result.success?
  end

  def test_ipv4_rejects_five_octets
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("1.2.3.4.5")
    refute result.success?
  end

  def test_ipv4_rejects_three_octets
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("1.2.3")
    refute result.success?
  end

  def test_ipv4_rejects_256_in_octet
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("256.1.1.1")
    refute result.success?
  end

  def test_ipv4_rejects_negative_octet
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("-1.0.0.0")
    refute result.success?
  end

  def test_ipv4_rejects_hex_notation
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("0xC0.0xA8.0x01.0x01")
    refute result.success?
  end

  def test_ipv4_accepts_valid_address
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("192.168.1.1")
    assert result.success?
  end

  def test_ipv4_accepts_all_zeros
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("0.0.0.0")
    assert result.success?
  end

  def test_ipv4_accepts_broadcast
    s = AnyVali.string.format("ipv4")
    result = s.safe_parse("255.255.255.255")
    assert result.success?
  end
end

# --------------------------------------------------------------------------
# Unicode length constraints -- code points, not bytes/code units
# --------------------------------------------------------------------------
class TestUnicodeLength < Minitest::Test
  def test_astral_code_point_counts_as_one_character
    emoji = "😀"
    assert AnyVali.string.max_length(1).safe_parse(emoji).success?
    refute AnyVali.string.min_length(2).safe_parse(emoji).success?
  end

  def test_imported_max_length_uses_code_points
    schema = AnyVali.import_schema(
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "string", "maxLength" => 1 },
      "definitions" => {},
      "extensions" => {}
    )
    assert schema.safe_parse("😀").success?
  end
end

# --------------------------------------------------------------------------
# CWE-400: Large inputs -- resource exhaustion
# --------------------------------------------------------------------------
class TestLargeInputs < Minitest::Test
  def test_large_string_validates_within_time_limit
    s = AnyVali.string
    payload = "x" * 1_000_000

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(payload)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    assert result.success?
    assert elapsed < 5, "Large string validation took #{elapsed}s (limit 5s)"
  end

  def test_large_string_with_pattern_validates_within_time_limit
    s = AnyVali.string.pattern("^[a-z]+$")
    payload = "a" * 100_000

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(payload)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    assert result.success?
    assert elapsed < 5, "Large string pattern validation took #{elapsed}s (limit 5s)"
  end

  def test_large_array_validates_within_time_limit
    s = AnyVali.array(AnyVali.int_)
    payload = Array.new(100_000) { |i| i }

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(payload)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    assert result.success?
    assert elapsed < 5, "Large array validation took #{elapsed}s (limit 5s)"
  end

  def test_large_object_validates_within_time_limit
    s = AnyVali.record(AnyVali.int_)
    payload = {}
    10_000.times { |i| payload["key_#{i}"] = i }

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(payload)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    assert result.success?
    assert elapsed < 5, "Large object validation took #{elapsed}s (limit 5s)"
  end

  def test_deeply_nested_object_does_not_crash
    s = AnyVali.object(
      properties: {
        "child" => AnyVali.object(
          properties: {
            "child" => AnyVali.object(
              properties: {
                "child" => AnyVali.object(
                  properties: { "value" => AnyVali.string },
                  required: ["value"],
                  unknown_keys: "allow"
                )
              },
              required: [],
              unknown_keys: "allow"
            )
          },
          required: [],
          unknown_keys: "allow"
        )
      },
      required: [],
      unknown_keys: "allow"
    )

    input = { "child" => { "child" => { "child" => { "value" => "deep" } } } }

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    result = s.safe_parse(input)
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    assert result.success?
    assert elapsed < 5, "Deeply nested object took #{elapsed}s (limit 5s)"
  end

  def test_large_union_validates_within_time_limit
    # Union with many variants - worst case for union matching
    variants = (1..100).map { |i| AnyVali.literal(i) }
    s = AnyVali.union(*variants)

    start = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    # Use a value that does NOT match any variant (worst case)
    result = s.safe_parse("not_a_number")
    elapsed = Process.clock_gettime(Process::CLOCK_MONOTONIC) - start

    refute result.success?
    assert elapsed < 5, "Large union validation took #{elapsed}s (limit 5s)"
  end
end

# --------------------------------------------------------------------------
# Schema import injection -- unknown / malicious kinds rejected
# --------------------------------------------------------------------------
class TestSchemaImportInjection < Minitest::Test
  def test_import_rejects_unknown_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "exec_system_command" },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_empty_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "" },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(StandardError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_nil_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => nil },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(StandardError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_numeric_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => 42 },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(StandardError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_sql_injection_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "'; DROP TABLE users; --" },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_prototype_pollution_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "__proto__" },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_constructor_kind
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "constructor" },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_missing_root
    assert_raises(StandardError) { AnyVali.import_schema({}) }
    assert_raises(StandardError) { AnyVali.import_schema({ "root" => nil }) }
  end

  def test_import_rejects_unknown_kind_in_nested_property
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "object",
        "properties" => {
          "evil" => { "kind" => "backdoor" }
        },
        "required" => ["evil"],
        "unknownKeys" => "reject"
      },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_unknown_kind_in_array_items
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "array",
        "items" => { "kind" => "malicious" }
      },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_unknown_kind_in_union_variant
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => {
        "kind" => "union",
        "variants" => [
          { "kind" => "string" },
          { "kind" => "exploit" }
        ]
      },
      "definitions" => {},
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end

  def test_import_rejects_unknown_kind_in_definition
    doc = {
      "anyvaliVersion" => "1.0",
      "schemaVersion" => "1",
      "root" => { "kind" => "string" },
      "definitions" => {
        "Evil" => { "kind" => "shell_exec" }
      },
      "extensions" => {}
    }
    assert_raises(AnyVali::ValidationError) { AnyVali.import_schema(doc) }
  end
end
