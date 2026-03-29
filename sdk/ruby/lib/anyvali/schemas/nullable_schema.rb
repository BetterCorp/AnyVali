# frozen_string_literal: true

module AnyVali
  class NullableSchema < Schema
    attr_reader :inner_schema

    def initialize(schema:, **kwargs)
      @inner_schema = schema
      super(kind: "nullable", **kwargs)
    end

    def to_node
      node = super
      node["schema"] = @inner_schema.to_node
      node
    end

    protected

    def validate(value, path, issues, context)
      return if value.nil?
      result = @inner_schema.safe_parse(value, path: path, context: context)
      issues.concat(result.issues) if result.failure?
    end

    def dup_with(**overrides)
      schema = overrides.delete(:schema) || @inner_schema
      attrs = {
        schema: schema,
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
