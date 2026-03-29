# frozen_string_literal: true

module AnyVali
  class TupleSchema < Schema
    attr_reader :element_schemas

    def initialize(elements:, **kwargs)
      @element_schemas = elements.freeze
      super(kind: "tuple", **kwargs)
    end

    def to_node
      node = super
      node["elements"] = @element_schemas.map(&:to_node)
      node
    end

    protected

    def validate(value, path, issues, context)
      unless value.is_a?(Array)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "tuple",
          received: Schema.type_name(value)
        )
        return
      end

      expected_len = @element_schemas.length
      if value.length < expected_len
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_SMALL,
          path: path,
          expected: expected_len.to_s,
          received: value.length.to_s
        )
        return
      end

      if value.length > expected_len
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_LARGE,
          path: path,
          expected: expected_len.to_s,
          received: value.length.to_s
        )
        return
      end

      @element_schemas.each_with_index do |schema, i|
        result = schema.safe_parse(value[i], path: path + [i], context: context)
        issues.concat(result.issues) if result.failure?
      end
    end

    def dup_with(**overrides)
      elements = overrides.delete(:elements) || @element_schemas
      attrs = {
        elements: elements,
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
