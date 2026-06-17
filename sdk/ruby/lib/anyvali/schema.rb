# frozen_string_literal: true

module AnyVali
  class Schema
    attr_reader :kind, :constraints, :coerce_config, :default_value, :has_default,
                :custom_validators, :metadata

    RESERVED_METADATA_KEYS = %w[title description deprecated deprecatedMessage notStable since sensitive readonly writeonly examples].freeze

    def initialize(kind:, constraints: {}, coerce_config: nil, default_value: nil, has_default: false, custom_validators: [], metadata: {})
      @kind = kind
      @constraints = constraints.freeze
      @coerce_config = coerce_config
      @default_value = default_value
      @has_default = has_default
      @custom_validators = custom_validators.freeze
      @metadata = metadata.freeze
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

    # Enable coercion on this schema.
    #
    # With no argument (`schema.coerce`) the only portable coercion source —
    # "string" (spec 5.1) — is implied, and the coercion target is inferred
    # from the schema kind (e.g. number().coerce coerces a string to a number).
    # An explicit config string/array (e.g. "string->int", "trim",
    # ["trim", "lower"]) is still accepted for backwards compatibility.
    def coerce(config = "string")
      dup_with(coerce_config: config)
    end

    def describe(description, **opts)
      raise ArgumentError, "describe(): description must be a string" unless description.is_a?(String)

      meta = { "description" => description }

      if opts.key?(:title)
        raise ArgumentError, "describe(): title must be a string" unless opts[:title].is_a?(String)
        meta["title"] = opts[:title]
      end
      if opts.key?(:deprecated)
        raise ArgumentError, "describe(): deprecated must be a boolean" unless [true, false].include?(opts[:deprecated])
        meta["deprecated"] = opts[:deprecated]
      end
      if opts.key?(:deprecated_message)
        raise ArgumentError, "describe(): deprecatedMessage must be a string" unless opts[:deprecated_message].is_a?(String)
        raise ArgumentError, "describe(): deprecatedMessage requires deprecated: true" unless opts[:deprecated]
        meta["deprecatedMessage"] = opts[:deprecated_message]
      end
      if opts.key?(:not_stable)
        raise ArgumentError, "describe(): notStable must be a boolean" unless [true, false].include?(opts[:not_stable])
        meta["notStable"] = opts[:not_stable]
      end
      if opts.key?(:since)
        raise ArgumentError, "describe(): since must be a string" unless opts[:since].is_a?(String)
        meta["since"] = opts[:since]
      end
      if opts.key?(:sensitive)
        raise ArgumentError, "describe(): sensitive must be a boolean" unless [true, false].include?(opts[:sensitive])
        meta["sensitive"] = opts[:sensitive]
      end
      if opts.key?(:readonly)
        raise ArgumentError, "describe(): readonly must be a boolean" unless [true, false].include?(opts[:readonly])
        meta["readonly"] = opts[:readonly]
      end
      if opts.key?(:writeonly)
        raise ArgumentError, "describe(): writeonly must be a boolean" unless [true, false].include?(opts[:writeonly])
        meta["writeonly"] = opts[:writeonly]
      end
      if opts[:readonly] && opts[:writeonly]
        raise ArgumentError, "describe(): readonly and writeonly cannot both be true"
      end
      if opts.key?(:examples)
        raise ArgumentError, "describe(): examples must be an array" unless opts[:examples].is_a?(Array)
        meta["examples"] = opts[:examples]
      end

      dup_with(metadata: @metadata.merge(meta))
    end

    def with_metadata(meta, replace: false)
      meta.each_key do |key|
        if RESERVED_METADATA_KEYS.include?(key)
          raise ArgumentError, "with_metadata(): \"#{key}\" is a reserved key. Use describe() instead."
        end
      end

      if replace
        preserved = @metadata.select { |k, _| RESERVED_METADATA_KEYS.include?(k) }
        dup_with(metadata: preserved.merge(meta))
      else
        dup_with(metadata: @metadata.merge(meta))
      end
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
      node["metadata"] = @metadata.dup unless @metadata.empty?
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
        custom_validators: @custom_validators,
        metadata: @metadata
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
