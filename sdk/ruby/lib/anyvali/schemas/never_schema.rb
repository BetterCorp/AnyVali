# frozen_string_literal: true

module AnyVali
  class NeverSchema < Schema
    def initialize(**kwargs)
      super(kind: "never", **kwargs)
    end

    protected

    def validate(value, path, issues, context)
      issues << ValidationIssue.new(
        code: IssueCodes::INVALID_TYPE,
        path: path,
        expected: "never",
        received: Schema.type_name(value)
      )
    end
  end
end
