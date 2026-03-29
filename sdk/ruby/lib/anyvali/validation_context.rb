# frozen_string_literal: true

module AnyVali
  class ValidationContext
    attr_reader :definitions

    def initialize(definitions: {})
      @definitions = definitions
    end
  end
end
