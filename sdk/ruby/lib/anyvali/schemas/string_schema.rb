# frozen_string_literal: true

module AnyVali
  class StringSchema < Schema
    def initialize(constraints: {}, **kwargs)
      super(kind: "string", constraints: constraints, **kwargs)
    end

    def min_length(n)
      dup_with(constraints: @constraints.merge("minLength" => n))
    end

    def max_length(n)
      dup_with(constraints: @constraints.merge("maxLength" => n))
    end

    def pattern(re)
      pat = re.is_a?(Regexp) ? re.source : re
      dup_with(constraints: @constraints.merge("pattern" => pat))
    end

    def starts_with(prefix)
      dup_with(constraints: @constraints.merge("startsWith" => prefix))
    end

    def ends_with(suffix)
      dup_with(constraints: @constraints.merge("endsWith" => suffix))
    end

    def includes(substr)
      dup_with(constraints: @constraints.merge("includes" => substr))
    end

    def format(fmt)
      dup_with(constraints: @constraints.merge("format" => fmt))
    end

    protected

    def validate(value, path, issues, context)
      unless value.is_a?(String)
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_TYPE,
          path: path,
          expected: "string",
          received: Schema.type_name(value)
        )
        return
      end

      if @constraints["minLength"] && value.length < @constraints["minLength"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_SMALL,
          path: path,
          expected: @constraints["minLength"].to_s,
          received: value.length.to_s
        )
      end

      if @constraints["maxLength"] && value.length > @constraints["maxLength"]
        issues << ValidationIssue.new(
          code: IssueCodes::TOO_LARGE,
          path: path,
          expected: @constraints["maxLength"].to_s,
          received: value.length.to_s
        )
      end

      if @constraints["pattern"]
        re = Regexp.new(@constraints["pattern"])
        unless re.match?(value)
          issues << ValidationIssue.new(
            code: IssueCodes::INVALID_STRING,
            path: path,
            expected: @constraints["pattern"],
            received: value
          )
        end
      end

      if @constraints["startsWith"] && !value.start_with?(@constraints["startsWith"])
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_STRING,
          path: path,
          expected: @constraints["startsWith"],
          received: value
        )
      end

      if @constraints["endsWith"] && !value.end_with?(@constraints["endsWith"])
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_STRING,
          path: path,
          expected: @constraints["endsWith"],
          received: value
        )
      end

      if @constraints["includes"] && !value.include?(@constraints["includes"])
        issues << ValidationIssue.new(
          code: IssueCodes::INVALID_STRING,
          path: path,
          expected: @constraints["includes"],
          received: value
        )
      end

      if @constraints["format"]
        Format::Validators.validate(value, @constraints["format"], path, issues)
      end
    end
  end
end
