# frozen_string_literal: true

module AnyVali
  class RecordSchema < Schema
    attr_reader :values_schema

    def initialize(values:, **kwargs)
      @values_schema = values
      super(kind: "record", **kwargs)
    end

    def to_node
      node = super
      node["values"] = @values_schema.to_node
      node
    end

    protected

    def validate(value, path, issues, context)
      unless value.is_a?(Hash)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "record",
          received: Schema.type_name(value)
        )
        return
      end

      value.each do |k, v|
        result = @values_schema.safe_parse(v, path: path + [k], context: context)
        issues.concat(result.issues) if result.failure?
      end
    end

    def dup_with(**overrides)
      values = overrides.delete(:values) || @values_schema
      attrs = {
        values: values,
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
