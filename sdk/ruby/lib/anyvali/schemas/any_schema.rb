# frozen_string_literal: true

module AnyVali
  class AnySchema < Schema
    def initialize(**kwargs)
      super(kind: "any", **kwargs)
    end

    protected

    def validate(value, path, issues, context)
      # any accepts everything
    end
  end
end
