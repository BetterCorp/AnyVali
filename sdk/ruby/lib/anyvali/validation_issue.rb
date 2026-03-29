# frozen_string_literal: true

module AnyVali
  class ValidationIssue
    attr_reader :code, :message, :path, :expected, :received, :meta

    def initialize(code:, message: nil, path: [], expected: nil, received: nil, meta: nil)
      @code = code
      @message = message || default_message(code)
      @path = path.freeze
      @expected = expected
      @received = received
      @meta = meta
      freeze
    end

    def to_h
      h = { "code" => @code, "path" => @path.dup }
      h["expected"] = @expected unless @expected.nil?
      h["received"] = @received unless @received.nil?
      h["message"] = @message if @message
      h["meta"] = @meta if @meta
      h
    end

    private

    def default_message(code)
      case code
      when IssueCodes::INVALID_TYPE then "Invalid type"
      when IssueCodes::REQUIRED then "Required field missing"
      when IssueCodes::UNKNOWN_KEY then "Unknown key"
      when IssueCodes::TOO_SMALL then "Value too small"
      when IssueCodes::TOO_LARGE then "Value too large"
      when IssueCodes::INVALID_STRING then "Invalid string"
      when IssueCodes::INVALID_NUMBER then "Invalid number"
      when IssueCodes::INVALID_LITERAL then "Invalid literal"
      when IssueCodes::INVALID_UNION then "Invalid union"
      when IssueCodes::COERCION_FAILED then "Coercion failed"
      when IssueCodes::DEFAULT_INVALID then "Default value is invalid"
      else "Validation failed"
      end
    end
  end
end
