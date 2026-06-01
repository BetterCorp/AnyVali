# frozen_string_literal: true

module AnyVali
  class RefSchema < Schema
    attr_reader :ref

    def initialize(ref:, **kwargs)
      @ref = ref
      super(kind: "ref", **kwargs)
    end

    def to_node
      { "kind" => "ref", "ref" => @ref }
    end

    def safe_parse(input, path: [], context: nil)
      context ||= ValidationContext.new
      resolved = resolve(context)
      if resolved.nil?
        issues = [ValidationIssue.new(
          code: IssueCodes::UNSUPPORTED_SCHEMA_KIND,
          path: path,
          expected: @ref,
          received: "unresolved ref"
        )]
        return ParseResult.new(value: nil, issues: issues)
      end
      context.instance_variable_set(:@active_refs, {}) unless context.instance_variable_defined?(:@active_refs)
      active_refs = context.instance_variable_get(:@active_refs)
      key = "#{@ref}:#{input.object_id}"
      if active_refs[key]
        return ParseResult.new(value: nil, issues: [
          ValidationIssue.new(
            code: IssueCodes::INVALID_TYPE,
            path: path,
            expected: @ref,
            received: "recursive ref"
          )
        ])
      end
      active_refs[key] = true
      begin
        resolved.safe_parse(input, path: path, context: context)
      ensure
        active_refs.delete(key)
      end
    end

    protected

    def validate(value, path, issues, context)
      # Handled in safe_parse override
    end

    def dup_with(**overrides)
      attrs = {
        ref: @ref,
        kind: @kind,
        constraints: @constraints,
        coerce_config: @coerce_config,
        default_value: @default_value,
        has_default: @has_default,
        custom_validators: @custom_validators
      }.merge(overrides)
      self.class.new(**attrs)
    end

    private

    def resolve(context)
      # ref format: "#/definitions/Name"
      name = @ref.sub("#/definitions/", "")
      context.definitions[name]
    end
  end
end
