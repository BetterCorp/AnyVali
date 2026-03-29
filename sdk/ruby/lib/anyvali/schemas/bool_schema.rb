# frozen_string_literal: true

module AnyVali
  class BoolSchema < Schema
    def initialize(**kwargs)
      super(kind: "bool", **kwargs)
    end

    protected

    def validate(value, path, issues, context)
      unless value.is_a?(TrueClass) || value.is_a?(FalseClass)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "bool",
          received: Schema.type_name(value)
        )
      end
    end
  end
end
