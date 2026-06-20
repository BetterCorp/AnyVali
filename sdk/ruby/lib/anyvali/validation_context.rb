# frozen_string_literal: true

module AnyVali
  class ValidationContext
    attr_reader :definitions
    attr_accessor :inherited_unknown_keys

    def initialize(definitions: {})
      @definitions = definitions
      @inherited_unknown_keys = nil
    end
  end
end
