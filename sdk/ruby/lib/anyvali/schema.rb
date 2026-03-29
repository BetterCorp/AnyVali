# frozen_string_literal: true

module AnyVali
  class Schema
    attr_reader :kind, :constraints, :coerce_config, :default_value, :has_default,
                :custom_validators

    def initialize(kind:, constraints: {}, coerce_config: nil, default_value: nil, has_default: false, custom_validators: [])
      @kind = kind
      @constraints = constraints.freeze
      @coerce_config = coerce_config
      @default_value = default_value
      @has_default = has_default
      @custom_validators = custom_validators.freeze
    end

    def parse(input)
      result = safe_parse(input)
      raise ValidationError, result.issues if result.failure?
      result.value
    end

    def safe_parse(input, path: [], context: nil)
      context ||= ValidationContext.new
      value = input
      issues = []

      # Step 1: Coercion (if present and configured)
      if @coerce_config && !value.nil?
        coerced = Coercion.apply(value, @coerce_config, @kind)
        if coerced[:success]
          value = coerced[:value]
        else
          issues << ValidationIssue.new(
            code: IssueCodes::COERCION_FAILED,
            path: path,
            expected: @kind,
            received: value.is_a?(String) ? value : value.to_s
          )
          return ParseResult.new(value: nil, issues: issues)
        end
      end

      # Step 2: Validate
      validate(value, path, issues, context)

      # Step 3: Custom validators
      if issues.empty? && !@custom_validators.empty?
        @custom_validators.each do |validator|
          validator_issues = validator.call(value, path)
          issues.concat(validator_issues) if validator_issues
        end
      end

      if issues.empty?
        ParseResult.new(value: value, issues: [])
      else
        ParseResult.new(value: nil, issues: issues)
      end
    end

    def default(value)
      dup_with(default_value: value, has_default: true)
    end

    def coerce(config)
      dup_with(coerce_config: config)
    end

    def refine(&block)
      dup_with(custom_validators: @custom_validators + [block])
    end

    def portable?
      @custom_validators.empty?
    end

    def export(mode: :portable)
      if mode == :portable && !portable?
        raise ValidationError, [
          ValidationIssue.new(
            code: IssueCodes::CUSTOM_VALIDATION_NOT_PORTABLE,
            expected: "portable schema",
            received: "schema with custom validators"
          )
        ]
      end
      doc = AnyValiDocument.new(root: self)
      doc.to_h
    end

    def to_node
      node = { "kind" => @kind }
      @constraints.each { |k, v| node[k] = v }
      node["coerce"] = @coerce_config if @coerce_config
      node["default"] = @default_value if @has_default
      node
    end

    protected

    def validate(value, path, issues, context)
      raise NotImplementedError, "Subclasses must implement validate"
    end

    def dup_with(**overrides)
      attrs = {
        kind: @kind,
        constraints: @constraints,
        coerce_config: @coerce_config,
        default_value: @default_value,
        has_default: @has_default,
        custom_validators: @custom_validators
      }.merge(overrides)
      self.class.new(**attrs)
    end

    # Helper to get the JSON type name for a Ruby value
    def self.type_name(value)
      case value
      when NilClass then "null"
      when TrueClass, FalseClass then "boolean"
      when Integer then "number"
      when Float then "number"
      when String then "string"
      when Array then "array"
      when Hash then "object"
      else value.class.name.downcase
      end
    end
  end
end
