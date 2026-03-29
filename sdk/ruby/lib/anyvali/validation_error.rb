# frozen_string_literal: true

module AnyVali
  class ValidationError < StandardError
    attr_reader :issues

    def initialize(issues)
      @issues = issues
      messages = issues.map { |i| "#{i.code} at #{i.path.inspect}: expected #{i.expected}, received #{i.received}" }
      super("Validation failed: #{messages.join('; ')}")
    end
  end
end
