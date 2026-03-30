# frozen_string_literal: true

if ENV["COVERAGE"]
  require "simplecov"
  require "simplecov-cobertura"
  SimpleCov.start do
    add_filter "/test/"
    formatter SimpleCov::Formatter::CoberturaFormatter
  end
end

$LOAD_PATH.unshift File.expand_path("../lib", __dir__)

require "anyvali"
require "minitest/autorun"
