# frozen_string_literal: true

require "json"

module AnyVali
  module Interchange
    module Exporter
      module_function

      def export(schema, mode: :portable, definitions: {})
        if mode == :portable && !schema.portable?
          raise ValidationError, [
            ValidationIssue.new(
              code: IssueCodes::CUSTOM_VALIDATION_NOT_PORTABLE,
              expected: "portable schema",
              received: "schema with custom validators"
            )
          ]
        end

        doc = {
          "anyvaliVersion" => AnyValiDocument::ANYVALI_VERSION,
          "schemaVersion" => AnyValiDocument::SCHEMA_VERSION,
          "root" => schema_to_node(schema),
          "definitions" => definitions.transform_values { |v| schema_to_node(v) },
          "extensions" => {}
        }
        doc
      end

      def to_json(schema, mode: :portable, definitions: {})
        JSON.generate(export(schema, mode: mode, definitions: definitions))
      end

      def schema_to_node(schema)
        schema.to_node
      end
    end
  end
end
