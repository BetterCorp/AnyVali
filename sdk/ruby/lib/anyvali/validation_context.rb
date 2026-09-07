# frozen_string_literal: true

module AnyVali
  class ValidationContext
    attr_reader :definitions
    attr_accessor :inherited_unknown_keys, :sensitive_mode, :sensitive_transform, :sensitive_cache

    def initialize(definitions: {})
      @definitions = definitions
      @inherited_unknown_keys = nil
      @sensitive_mode = nil
      @sensitive_transform = nil
      @sensitive_cache = nil
    end
  end
end
