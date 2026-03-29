# frozen_string_literal: true

module AnyVali
  class IntersectionSchema < Schema
    attr_reader :all_of_schemas

    def initialize(all_of:, **kwargs)
      @all_of_schemas = all_of.freeze
      super(kind: "intersection", **kwargs)
    end

    def to_node
      node = super
      node["allOf"] = @all_of_schemas.map(&:to_node)
      node
    end

    def safe_parse(input, path: [], context: nil)
      context ||= ValidationContext.new
      all_issues = []
      merged_output = nil

      @all_of_schemas.each do |schema|
        result = schema.safe_parse(input, path: path, context: context)
        if result.failure?
          all_issues.concat(result.issues)
        else
          if merged_output.nil?
            merged_output = result.value
          elsif merged_output.is_a?(Hash) && result.value.is_a?(Hash)
            merged_output = merged_output.merge(result.value)
          else
            merged_output = result.value
          end
        end
      end

      if all_issues.empty?
        ParseResult.new(value: merged_output, issues: [])
      else
        ParseResult.new(value: nil, issues: all_issues)
      end
    end

    protected

    def validate(value, path, issues, context)
      # Handled in safe_parse override
    end

    def dup_with(**overrides)
      all_of = overrides.delete(:all_of) || @all_of_schemas
      attrs = {
        all_of: all_of,
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
