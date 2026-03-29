# frozen_string_literal: true

module AnyVali
  class EnumSchema < Schema
    attr_reader :values

    def initialize(values:, **kwargs)
      @values = values.freeze
      super(kind: "enum", **kwargs)
    end

    def to_node
      node = super
      node["values"] = @values
      node
    end

    protected

    def validate(value, path, issues, context)
      match = @values.any? { |v| v.class == value.class && v == value }
      # Special nil handling
      match ||= @values.include?(nil) && value.nil?

      unless match
        enum_str = @values.map(&:to_s).join(",")
        received_str = value.nil? ? "null" : value.to_s
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "enum(#{enum_str})",
          received: received_str
        )
      end
    end

    def dup_with(**overrides)
      attrs = {
        values: @values,
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
