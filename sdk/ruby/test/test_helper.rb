# frozen_string_literal: true

if ENV["COVERAGE"]
  require "simplecov"
  require "simplecov-cobertura"
  # Wrap the Cobertura formatter so a DTD / XML generation error
  # does not turn a passing test run into a CI failure.
  safe_formatter = Class.new(SimpleCov::Formatter::CoberturaFormatter) do
    def format(result)
      super
    rescue StandardError => e
      $stderr.puts "[simplecov] Cobertura XML generation failed: #{e.message}"
    end
  end

  SimpleCov.start do
    add_filter "/test/"
    add_filter "/sig/"
    formatter safe_formatter
  end
end

$LOAD_PATH.unshift File.expand_path("../lib", __dir__)

require "anyvali"
require "minitest/autorun"
