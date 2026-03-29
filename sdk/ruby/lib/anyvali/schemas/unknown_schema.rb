# frozen_string_literal: true

module AnyVali
  class UnknownSchema < Schema
    def initialize(**kwargs)
      super(kind: "unknown", **kwargs)
    end

    protected

    def validate(value, path, issues, context)
      # unknown accepts everything
    end
  end
end
