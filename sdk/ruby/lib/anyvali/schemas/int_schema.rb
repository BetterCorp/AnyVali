# frozen_string_literal: true

module AnyVali
  class IntSchema < NumberSchema
    def initialize(kind: "int", **kwargs)
      super(kind: kind, **kwargs)
    end
  end
end
