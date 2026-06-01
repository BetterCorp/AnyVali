# frozen_string_literal: true

module AnyVali
  class AnyValiDocument
    ANYVALI_VERSION = "1.0"
    SCHEMA_VERSION = "1"

    attr_reader :anyvali_version, :schema_version, :root, :definitions, :extensions

    def initialize(root:, definitions: {}, extensions: {}, anyvali_version: ANYVALI_VERSION, schema_version: SCHEMA_VERSION)
      @anyvali_version = anyvali_version
      @schema_version = schema_version
      @root = root
      @definitions = definitions
      @extensions = extensions
    end

    def to_h
      {
        "anyvaliVersion" => @anyvali_version,
        "schemaVersion" => @schema_version,
        "root" => node_to_h(@root),
        "definitions" => @definitions.transform_values { |v| node_to_h(v) },
        "extensions" => @extensions
      }
    end

    def to_json(*_args)
      require "json"
      JSON.generate(to_h)
    end

    private

    def node_to_h(node)
      case node
      when Hash then node
      when Schema then node.to_node
      else node
      end
    end
  end
end
