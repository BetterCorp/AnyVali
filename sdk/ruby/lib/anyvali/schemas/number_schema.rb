# frozen_string_literal: true

module AnyVali
  class NumberSchema < Schema
    FLOAT_KINDS = %w[number float32 float64].freeze
    INT_KINDS = %w[int int8 int16 int32 int64 uint8 uint16 uint32 uint64].freeze

    INT_RANGES = {
      "int8"   => [-128, 127],
      "int16"  => [-32_768, 32_767],
      "int32"  => [-2_147_483_648, 2_147_483_647],
      "int64"  => [-9_223_372_036_854_775_808, 9_223_372_036_854_775_807],
      "uint8"  => [0, 255],
      "uint16" => [0, 65_535],
      "uint32" => [0, 4_294_967_295],
      "uint64" => [0, 18_446_744_073_709_551_615],
      "int"    => [-9_223_372_036_854_775_808, 9_223_372_036_854_775_807]
    }.freeze

    FLOAT32_MAX = 3.4028235e+38

    def initialize(kind: "number", constraints: {}, **kwargs)
      super(kind: kind, constraints: constraints, **kwargs)
    end

    def min(n)
      dup_with(constraints: @constraints.merge("min" => n))
    end

    def max(n)
      dup_with(constraints: @constraints.merge("max" => n))
    end

    def exclusive_min(n)
      dup_with(constraints: @constraints.merge("exclusiveMin" => n))
    end

    def exclusive_max(n)
      dup_with(constraints: @constraints.merge("exclusiveMax" => n))
    end

    def multiple_of(n)
      dup_with(constraints: @constraints.merge("multipleOf" => n))
    end

    protected

    def validate(value, path, issues, context)
      int_kind = INT_KINDS.include?(@kind)
      float_kind = FLOAT_KINDS.include?(@kind)

      if int_kind
        validate_int(value, path, issues)
      elsif float_kind
        validate_float(value, path, issues)
      end
    end

    private

    def validate_int(value, path, issues)
      # Must be numeric
      unless value.is_a?(Integer) || value.is_a?(Float)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: @kind,
          received: Schema.type_name(value)
        )
        return
      end

      # Must be a whole number
      if value.is_a?(Float) && value != value.floor
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: @kind,
          received: "number"
        )
        return
      end

      int_val = value.is_a?(Float) ? value.to_i : value

      # Width range check
      if INT_RANGES.key?(@kind)
        min_val, max_val = INT_RANGES[@kind]
        if int_val < min_val
          issues << ValidationIssue.new(
            code: IssueCodes::TOO_SMALL,
            path: path,
            expected: @kind,
            received: int_val.to_s
          )
          return
        end
        if int_val > max_val
          issues << ValidationIssue.new(
            code: IssueCodes::TOO_LARGE,
            path: path,
            expected: @kind,
            received: int_val.to_s
          )
          return
        end
      end

      validate_numeric_constraints(int_val, path, issues)
    end

    def validate_float(value, path, issues)
      unless value.is_a?(Numeric) && !value.is_a?(Complex)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: @kind,
          received: Schema.type_name(value)
        )
        return
      end

      # For booleans (true/false are not Numeric in Ruby, so this is fine)
      if value == true || value == false
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: @kind,
          received: "boolean"
        )
        return
      end

      # float32 range check
      if @kind == "float32"
        fval = value.to_f
        if fval.abs > FLOAT32_MAX && !fval.infinite?
          issues << ValidationIssue.new(
            code: IssueCodes::TOO_LARGE,
            path: path,
            expected: "float32",
            received: value.to_s
          )
          return
        end
      end

      validate_numeric_constraints(value, path, issues)
    end

    def validate_numeric_constraints(value, path, issues)
      if @constraints["min"] && value < @constraints["min"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_SMALL,
          path: path,
          expected: @constraints["min"].to_s,
          received: value.to_s
        )
      end

      if @constraints["max"] && value > @constraints["max"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_LARGE,
          path: path,
          expected: @constraints["max"].to_s,
          received: value.to_s
        )
      end

      if @constraints["exclusiveMin"] && value <= @constraints["exclusiveMin"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_SMALL,
          path: path,
          expected: @constraints["exclusiveMin"].to_s,
          received: value.to_s
        )
      end

      if @constraints["exclusiveMax"] && value >= @constraints["exclusiveMax"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_LARGE,
          path: path,
          expected: @constraints["exclusiveMax"].to_s,
          received: value.to_s
        )
      end

      if @constraints["multipleOf"]
        divisor = @constraints["multipleOf"]
        remainder = value.to_f % divisor.to_f
        unless remainder.abs < 1e-10 || (divisor.to_f - remainder.abs).abs < 1e-10
          issues << ValidationIssue.new(
            code: IssueCodes::INVALID_NUMBER,
            path: path,
            expected: divisor.to_s,
            received: value.to_s
          )
        end
      end
    end

    def dup_with(**overrides)
      attrs = {
        kind: @kind,
        constraints: @constraints,
        coerce_config: @coerce_config,
        default_value: @default_value,
        has_default: @has_default,
        custom_validators: @custom_validators
      }.merge(overrides)
      self.class.new(**attrs)
    end
  end
end
