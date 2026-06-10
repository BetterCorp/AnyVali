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
        begin
          re = Regexp.new(ecma_anchors(@constraints["pattern"]))
          unless re.match?(value)
            issues << ValidationIssue.new(
              code: IssueCodes::INVALID_STRING,
              path: path,
              expected: @constraints["pattern"],
              received: value
            )
          end
        rescue RegexpError
          # Invalid regex pattern - treat as validation failure
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

    private

    # Rewrite ECMA-262 anchors to Ruby's absolute anchors. The spec (3.1) makes
    # ECMA-262 the portable regex baseline, where "^"/"$" without the multiline
    # flag match only the start/end of the whole string. In Ruby "^"/"$" are
    # ALWAYS line anchors, so /^admin$/ matches "x\nadmin\ny" -- a line/newline
    # injection bypass that diverges from the JS reference. Translate unescaped,
    # top-level "^" -> "\A" and "$" -> "\z" (absolute start/end). Anchors inside
    # character classes and escaped "\^"/"\$" are left untouched.
    def ecma_anchors(pattern)
      out = +""
      escaped = false
      in_class = false
      pattern.each_char do |ch|
        if escaped
          out << ch
          escaped = false
        elsif ch == "\\"
          out << ch
          escaped = true
        elsif ch == "["
          in_class = true
          out << ch
        elsif ch == "]" && in_class
          in_class = false
          out << ch
        elsif ch == "^" && !in_class
          out << "\\A"
        elsif ch == "$" && !in_class
          out << "\\z"
        else
          out << ch
        end
      end
      out
    end
  end
end
