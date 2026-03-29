# frozen_string_literal: true

module AnyVali
  class ObjectSchema < Schema
    attr_reader :properties, :required_keys, :unknown_keys

    def initialize(properties:, required: [], unknown_keys: "reject", **kwargs)
      @properties = properties.freeze
      @required_keys = required.freeze
      @unknown_keys = unknown_keys
      super(kind: "object", **kwargs)
    end

    def strict
      dup_with(unknown_keys: "reject")
    end

    def strip_unknown
      dup_with(unknown_keys: "strip")
    end

    def allow_unknown
      dup_with(unknown_keys: "allow")
    end

    def to_node
      node = super
      props = {}
      @properties.each do |k, v|
        props[k] = v.to_node
      end
      node["properties"] = props
      node["required"] = @required_keys
      node["unknownKeys"] = @unknown_keys
      node
    end

    def safe_parse(input, path: [], context: nil)
      context ||= ValidationContext.new
      issues = []

      unless input.is_a?(Hash)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "object",
          received: Schema.type_name(input)
        )
        return ParseResult.new(value: nil, issues: issues)
      end

      output = {}

      # Check required fields
      @required_keys.each do |key|
        unless input.key?(key)
          prop_schema = @properties[key]
          expected = prop_schema ? prop_schema.kind : "unknown"
          issues << ValidationIssue.new(
            code: IssueCodes::REQUIRED,
            path: path + [key],
            expected: expected,
            received: "undefined"
          )
        end
      end

      # Validate known properties
      @properties.each do |key, prop_schema|
        if input.key?(key)
          # Value is present
          val = input[key]

          # Handle default on property schema (value is present, don't apply default)
          result = prop_schema.safe_parse(val, path: path + [key], context: context)
          if result.success?
            output[key] = result.value
          else
            issues.concat(result.issues)
          end
        elsif prop_schema.has_default
          # Apply default
          default_val = prop_schema.default_value
          # Validate the default value
          result = prop_schema.safe_parse(default_val, path: path + [key], context: context)
          if result.success?
            output[key] = result.value
          else
            # Default is invalid
            issues << ValidationIssue.new(
              code: IssueCodes::DEFAULT_INVALID,
              path: path + [key],
              expected: result.issues.first&.expected || prop_schema.kind,
              received: default_val.to_s
            )
          end
        elsif prop_schema.is_a?(OptionalSchema)
          # Optional field absent - OK
        elsif !@required_keys.include?(key)
          # Non-required field absent without default - OK
        end
      end

      # Handle unknown keys
      unknown = input.keys - @properties.keys
      case @unknown_keys
      when "reject"
        unknown.each do |key|
          issues << ValidationIssue.new(
            code: IssueCodes::UNKNOWN_KEY,
            path: path + [key],
            expected: "undefined",
            received: key
          )
        end
      when "strip"
        # Just don't include them
      when "allow"
        unknown.each { |key| output[key] = input[key] }
      end

      if issues.empty?
        ParseResult.new(value: output, issues: [])
      else
        ParseResult.new(value: nil, issues: issues)
      end
    end

    protected

    def validate(value, path, issues, context)
      # Handled in safe_parse override
    end

    def dup_with(**overrides)
      attrs = {
        properties: @properties,
        required: @required_keys,
        unknown_keys: @unknown_keys,
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
