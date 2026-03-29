# frozen_string_literal: true

module AnyVali
  class ParseResult
    attr_reader :value, :issues

    def initialize(value:, issues:)
      @value = value
      @issues = issues.freeze
      freeze
    end

    def success?
      @issues.empty?
    end

    def failure?
      !success?
    end
  end
end
