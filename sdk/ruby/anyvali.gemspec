# frozen_string_literal: true

Gem::Specification.new do |s|
  s.name        = "anyvali"
  s.version     = "0.0.1"
  s.summary     = "Native validation with portable schema interchange"
  s.description = "AnyVali Ruby SDK - native validation with portable schema interchange across 10 languages"
  s.authors     = ["AnyVali Contributors"]
  s.email       = ["hello@anyvali.com"]
  s.homepage    = "https://anyvali.com"
  s.metadata    = { "source_code_uri" => "https://github.com/anyvali/anyvali" }
  s.license     = "MIT"

  s.required_ruby_version = ">= 3.0.0"

  s.files = Dir["lib/**/*.rb"]

  s.add_development_dependency "minitest", "~> 5.0"
  s.add_development_dependency "rake", "~> 13.0"
end
