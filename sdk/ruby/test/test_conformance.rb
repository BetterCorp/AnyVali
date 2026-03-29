# frozen_string_literal: true

require_relative "test_helper"
require "json"

class TestConformance < Minitest::Test
  CORPUS_DIR = File.expand_path("../../../../spec/corpus", __dir__)

  def self.load_corpus
    suites = []
    Dir.glob(File.join(CORPUS_DIR, "**", "*.json")).sort.each do |file|
      next if File.basename(file) == "README.md"
      begin
        data = JSON.parse(File.read(file))
        suites << data if data.is_a?(Hash) && data["cases"]
      rescue JSON::ParserError
        # skip non-JSON files
      end
    end
    suites
  end

  load_corpus.each do |suite|
    suite_name = suite["suite"]
    suite["cases"].each_with_index do |test_case, idx|
      desc = test_case["description"]
      method_name = "test_#{suite_name}_#{idx}_#{desc.downcase.gsub(/[^a-z0-9]+/, '_')}"

      define_method(method_name) do
        run_conformance_case(test_case)
      end
    end
  end

  private

  def run_conformance_case(test_case)
    doc = test_case["schema"]
    input = test_case["input"]
    expected_valid = test_case["valid"]
    expected_output = test_case["output"]
    expected_issues = test_case["issues"]

    schema, context, _defs = AnyVali.import(doc)
    result = schema.safe_parse(input, context: context)

    if expected_valid
      assert result.success?, "Expected success for: #{test_case['description']}, got issues: #{result.issues.map(&:to_h).inspect}"
      assert_equal_deep(expected_output, result.value, test_case["description"])
    else
      assert result.failure?, "Expected failure for: #{test_case['description']}"

      # Verify issue codes and paths match
      expected_issues.each_with_index do |expected_issue, i|
        actual_issue = result.issues[i]
        assert_not_nil actual_issue, "Expected issue at index #{i} for: #{test_case['description']}"
        assert_equal expected_issue["code"], actual_issue.code,
          "Issue code mismatch at index #{i} for: #{test_case['description']}"

        expected_path = expected_issue["path"]
        assert_equal expected_path, actual_issue.path,
          "Issue path mismatch at index #{i} for: #{test_case['description']}"

        if expected_issue["expected"]
          assert_equal expected_issue["expected"], actual_issue.expected,
            "Issue expected mismatch at index #{i} for: #{test_case['description']}"
        end

        if expected_issue["received"]
          assert_equal expected_issue["received"], actual_issue.received,
            "Issue received mismatch at index #{i} for: #{test_case['description']}"
        end
      end

      assert_equal expected_issues.length, result.issues.length,
        "Issue count mismatch for: #{test_case['description']}. Got: #{result.issues.map(&:to_h).inspect}"
    end
  end

  def assert_equal_deep(expected, actual, desc)
    case expected
    when Hash
      assert actual.is_a?(Hash), "Expected Hash for #{desc}, got #{actual.class}"
      expected.each do |k, v|
        assert_equal_deep(v, actual[k], "#{desc}.#{k}")
      end
      assert_equal expected.keys.sort, actual.keys.sort, "Key mismatch for #{desc}"
    when Array
      assert actual.is_a?(Array), "Expected Array for #{desc}, got #{actual.class}"
      assert_equal expected.length, actual.length, "Array length mismatch for #{desc}"
      expected.each_with_index do |v, i|
        assert_equal_deep(v, actual[i], "#{desc}[#{i}]")
      end
    when Float
      if actual.is_a?(Integer) && expected == actual.to_f
        # Allow integer/float equivalence for JSON round-trips
        return
      end
      assert_in_delta expected, actual, 1e-10, "Float mismatch for #{desc}"
    when Integer
      # In JSON, integers and floats can be equivalent
      if actual.is_a?(Float) && expected.to_f == actual
        return
      end
      assert_equal expected, actual, "Integer mismatch for #{desc}"
    when NilClass
      assert_nil actual, "Expected nil for #{desc}"
    else
      assert_equal expected, actual, "Value mismatch for #{desc}"
    end
  end

  def assert_not_nil(obj, msg = nil)
    refute_nil obj, msg
  end
end
