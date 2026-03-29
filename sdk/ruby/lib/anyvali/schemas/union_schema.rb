# frozen_string_literal: true

module AnyVali
  class UnionSchema < Schema
    attr_reader :variant_schemas

    def initialize(variants:, **kwargs)
      @variant_schemas = variants.freeze
      super(kind: "union", **kwargs)
    end

    def to_node
      node = super
      node["variants"] = @variant_schemas.map(&:to_node)
      node
    end

    protected

    def validate(value, path, issues, context)
      @variant_schemas.each do |schema|
        result = schema.safe_parse(value, path: path, context: context)
        return if result.success?
      end

      # No variant matched
      expected = @variant_schemas.map(&:kind).join(" | ")
      issues << ValidationIssue.new(
        code: IssueCodes::INVALID_UNION,
        path: path,
        expected: expected,
        received: Schema.type_name(value)
      )
    end

    def dup_with(**overrides)
      variants = overrides.delete(:variants) || @variant_schemas
      attrs = {
        variants: variants,
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
