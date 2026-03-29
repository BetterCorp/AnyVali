# frozen_string_literal: true

module AnyVali
  class NullSchema < Schema
    def initialize(**kwargs)
      super(kind: "null", **kwargs)
    end

    protected

    def validate(value, path, issues, context)
      unless value.nil?
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "null",
          received: Schema.type_name(value)
        )
      end
    end
  end
end
