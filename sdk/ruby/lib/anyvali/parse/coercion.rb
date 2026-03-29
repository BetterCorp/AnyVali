# frozen_string_literal: true

module AnyVali
  module Coercion
    module_function

    def apply(value, config, kind)
      configs = config.is_a?(Array) ? config : [config]
      current = value

      configs.each do |c|
        result = apply_single(current, c, kind)
        return result unless result[:success]
        current = result[:value]
      end

      { success: true, value: current }
    end

    def apply_single(value, config, kind)
      case config
      when "string->int"
        coerce_string_to_int(value)
      when "string->number"
        coerce_string_to_number(value)
      when "string->bool"
        coerce_string_to_bool(value)
      when "trim"
        coerce_trim(value)
      when "lower"
        coerce_lower(value)
      when "upper"
        coerce_upper(value)
      else
        { success: false, value: value }
      end
    end

    def coerce_string_to_int(value)
      return { success: true, value: value } if value.is_a?(Integer)
      return { success: false, value: value } unless value.is_a?(String)

      stripped = value.strip
      if stripped.match?(/\A-?\d+\z/)
        { success: true, value: stripped.to_i }
      else
        { success: false, value: value }
      end
    end

    def coerce_string_to_number(value)
      return { success: true, value: value } if value.is_a?(Numeric)
      return { success: false, value: value } unless value.is_a?(String)

      stripped = value.strip
      begin
        f = Float(stripped)
        { success: true, value: f }
      rescue ArgumentError
        { success: false, value: value }
      end
    end

    def coerce_string_to_bool(value)
      return { success: true, value: value } if value.is_a?(TrueClass) || value.is_a?(FalseClass)
      return { success: false, value: value } unless value.is_a?(String)

      case value.strip.downcase
      when "true", "1"
        { success: true, value: true }
      when "false", "0"
        { success: true, value: false }
      else
        { success: false, value: value }
      end
    end

    def coerce_trim(value)
      return { success: true, value: value } unless value.is_a?(String)
      { success: true, value: value.strip }
    end

    def coerce_lower(value)
      return { success: true, value: value } unless value.is_a?(String)
      { success: true, value: value.downcase }
    end

    def coerce_upper(value)
      return { success: true, value: value } unless value.is_a?(String)
      { success: true, value: value.upcase }
    end
  end
end
