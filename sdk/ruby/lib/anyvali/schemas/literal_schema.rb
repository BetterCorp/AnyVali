# frozen_string_literal: true

module AnyVali
  class LiteralSchema < Schema
    attr_reader :literal_value

    def initialize(value:, **kwargs)
      @literal_value = value
      super(kind: "literal", **kwargs)
    end

    def to_node
      node = super
      node["value"] = @literal_value
      node
    end

    protected

    def validate(value, path, issues, context)
      # Strict equality: must be same type and value
      unless value.class == @literal_value.class && value == @literal_value
        # Special case: nil comparisons
        if @literal_value.nil? && value.nil?
          return
        end
        # Handle Integer/Float comparisons: both are "number" but Integer 42 != Float 42.0 for literal
        expected_str = @literal_value.nil? ? "null" : @literal_value.to_s
        received_str = value.nil? ? "null" : value.to_s
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_LITERAL,
          path: path,
          expected: expected_str,
          received: received_str
        )
      end
    end

    def dup_with(**overrides)
      attrs = {
        value: @literal_value,
        kind: @kind,
        constraints: @constraints,
        coerce_config: @coerce_config,
        default_value: @default_value,
        has_default: @has_default,
        custom_validators: @custom_validators
      }.merge(overrides)
      self.class.new(**attrs)
    end
  end
end
