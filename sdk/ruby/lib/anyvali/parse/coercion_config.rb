# frozen_string_literal: true

module AnyVali
  module CoercionConfig
    PORTABLE_COERCIONS = %w[
      string->int
      string->number
      string->bool
      trim
      lower
      upper
    ].freeze

    def self.portable?(config)
      configs = config.is_a?(Array) ? config : [config]
      configs.all? { |c| PORTABLE_COERCIONS.include?(c) }
    end
  end
end
