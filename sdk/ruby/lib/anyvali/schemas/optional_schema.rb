# frozen_string_literal: true

module AnyVali
  class OptionalSchema < Schema
    attr_reader :inner_schema

    def initialize(schema:, **kwargs)
      @inner_schema = schema
      super(kind: "optional", **kwargs)
    end

    def to_node
      node = super
      node["schema"] = @inner_schema.to_node
      node
    end

    def safe_parse(input, path: [], context: nil)
      # When used in objects, absence is handled by ObjectSchema
      # If called directly with a value, delegate to inner schema
      context ||= ValidationContext.new
      @inner_schema.safe_parse(input, path: path, context: context)
    end

    protected

    def validate(value, path, issues, context)
      # Delegate to inner schema
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
