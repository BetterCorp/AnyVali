# frozen_string_literal: true

module AnyVali
  class ArraySchema < Schema
    attr_reader :items_schema

    def initialize(items:, constraints: {}, **kwargs)
      @items_schema = items
      super(kind: "array", constraints: constraints, **kwargs)
    end

    def min_items(n)
      dup_with(constraints: @constraints.merge("minItems" => n))
    end

    def max_items(n)
      dup_with(constraints: @constraints.merge("maxItems" => n))
    end

    def to_node
      node = super
      node["items"] = @items_schema.to_node
      node
    end

    protected

    def validate(value, path, issues, context)
      unless value.is_a?(Array)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "array",
          received: Schema.type_name(value)
        )
        return
      end

      if @constraints["minItems"] && value.length < @constraints["minItems"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_SMALL,
          path: path,
          expected: @constraints["minItems"].to_s,
          received: value.length.to_s
        )
      end

      if @constraints["maxItems"] && value.length > @constraints["maxItems"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_LARGE,
          path: path,
          expected: @constraints["maxItems"].to_s,
          received: value.length.to_s
        )
      end

      value.each_with_index do |item, i|
        result = @items_schema.safe_parse(item, path: path + [i], context: context)
        issues.concat(result.issues) if result.failure?
      end
    end

    def dup_with(**overrides)
      items = overrides.delete(:items) || @items_schema
      attrs = {
        items: items,
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
