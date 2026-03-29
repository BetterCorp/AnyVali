# frozen_string_literal: true

module AnyVali
  module Defaults
    module_function

    # Check if a default value is JSON-serializable (portable)
    def portable?(value)
      case value
      when NilClass, TrueClass, FalseClass, Integer, Float, String
        true
      when Array
        value.all? { |v| portable?(v) }
      when Hash
        value.all? { |k, v| k.is_a?(String) && portable?(v) }
      else
        false
      end
    end
  end
end
