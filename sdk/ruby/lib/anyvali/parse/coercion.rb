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
      when "string"
        # Generic, portable coercion source (spec 5.1). The only portable
        # source is "string"; infer the target from the schema kind. On a
        # string-kind schema this is a no-op (the value is already a string).
        coerce_from_string_to_kind(value, kind)
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

    # Map the portable "string" source to the concrete string->kind coercion
    # for the target schema kind. Integer family -> int; float family ->
    # number; bool -> bool; everything else (string/unknown/...) is a no-op.
    def coerce_from_string_to_kind(value, kind)
      case kind
      when "int", "int8", "int16", "int32", "int64",
           "uint8", "uint16", "uint32", "uint64"
        coerce_string_to_int(value)
      when "number", "float32", "float64"
        coerce_string_to_number(value)
      when "bool"
        coerce_string_to_bool(value)
      else
        { success: true, value: value }
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
      # Spec 5.1: parse DECIMAL floating-point only. Ruby's Float() also accepts
      # digit-group underscores ("1_000.5") and hex floats ("0x1.8p3"), which
      # diverge from the JS reference and let non-decimal strings slip through.
      return { success: false, value: value } unless stripped.match?(/\A[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?\z/)

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
