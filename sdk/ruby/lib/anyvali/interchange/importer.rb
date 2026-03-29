# frozen_string_literal: true

require "json"

module AnyVali
  module Interchange
    module Importer
      module_function

      def import(doc)
        doc = JSON.parse(doc) if doc.is_a?(String)

        # Build definitions first (needed for refs)
        definitions = {}
        if doc["definitions"]
          doc["definitions"].each do |name, node|
            definitions[name] = node_to_schema(node)
          end
        end

        root_schema = node_to_schema(doc["root"])
        context = ValidationContext.new(definitions: definitions)
        [root_schema, context, definitions]
      end

      def import_schema(doc)
        schema, context, definitions = import(doc)
        schema
      end

      def node_to_schema(node)
        kind = node["kind"]

        case kind
        when "string"
          build_string(node)
        when "number", "float32", "float64"
          build_number(node, kind)
        when "int", "int8", "int16", "int32", "int64",
             "uint8", "uint16", "uint32", "uint64"
          build_int(node, kind)
        when "bool"
          build_bool(node)
        when "null"
          NullSchema.new
        when "any"
          AnySchema.new
        when "unknown"
          UnknownSchema.new
        when "never"
          NeverSchema.new
        when "literal"
          LiteralSchema.new(value: node["value"])
        when "enum"
          EnumSchema.new(values: node["values"])
        when "array"
          build_array(node)
        when "tuple"
          build_tuple(node)
        when "object"
          build_object(node)
        when "record"
          build_record(node)
        when "union"
          build_union(node)
        when "intersection"
          build_intersection(node)
        when "optional"
          OptionalSchema.new(schema: node_to_schema(node["schema"]))
        when "nullable"
          build_nullable(node)
        when "ref"
          RefSchema.new(ref: node["ref"])
        else
          raise ValidationError, [
            ValidationIssue.new(
              code: IssueCodes::UNSUPPORTED_SCHEMA_KIND,
              expected: "known schema kind",
              received: kind.to_s
            )
          ]
        end
      end

      def build_string(node)
        constraints = {}
        %w[minLength maxLength pattern startsWith endsWith includes format].each do |k|
          constraints[k] = node[k] if node.key?(k)
        end
        coerce = node["coerce"]
        default_val = node["default"]
        has_default = node.key?("default")
        StringSchema.new(
          constraints: constraints,
          coerce_config: coerce,
          default_value: default_val,
          has_default: has_default
        )
      end

      def build_number(node, kind)
        constraints = {}
        %w[min max exclusiveMin exclusiveMax multipleOf].each do |k|
          constraints[k] = node[k] if node.key?(k)
        end
        coerce = node["coerce"]
        default_val = node["default"]
        has_default = node.key?("default")
        NumberSchema.new(
          kind: kind,
          constraints: constraints,
          coerce_config: coerce,
          default_value: default_val,
          has_default: has_default
        )
      end

      def build_int(node, kind)
        constraints = {}
        %w[min max exclusiveMin exclusiveMax multipleOf].each do |k|
          constraints[k] = node[k] if node.key?(k)
        end
        coerce = node["coerce"]
        default_val = node["default"]
        has_default = node.key?("default")
        IntSchema.new(
          kind: kind,
          constraints: constraints,
          coerce_config: coerce,
          default_value: default_val,
          has_default: has_default
        )
      end

      def build_bool(node)
        coerce = node["coerce"]
        default_val = node["default"]
        has_default = node.key?("default")
        BoolSchema.new(
          coerce_config: coerce,
          default_value: default_val,
          has_default: has_default
        )
      end

      def build_array(node)
        items = node_to_schema(node["items"])
        constraints = {}
        %w[minItems maxItems].each do |k|
          constraints[k] = node[k] if node.key?(k)
        end
        ArraySchema.new(items: items, constraints: constraints)
      end

      def build_tuple(node)
        elements = node["elements"].map { |e| node_to_schema(e) }
        TupleSchema.new(elements: elements)
      end

      def build_object(node)
        props = {}
        (node["properties"] || {}).each do |k, v|
          props[k] = node_to_schema(v)
        end
        required = node["required"] || []
        unknown_keys = node["unknownKeys"] || "reject"
        ObjectSchema.new(
          properties: props,
          required: required,
          unknown_keys: unknown_keys
        )
      end

      def build_record(node)
        values = node_to_schema(node["values"])
        RecordSchema.new(values: values)
      end

      def build_union(node)
        variants = node["variants"].map { |v| node_to_schema(v) }
        UnionSchema.new(variants: variants)
      end

      def build_intersection(node)
        all_of = node["allOf"].map { |v| node_to_schema(v) }
        IntersectionSchema.new(all_of: all_of)
      end

      def build_nullable(node)
        inner = node_to_schema(node["schema"])
        default_val = node["default"]
        has_default = node.key?("default")
        NullableSchema.new(
          schema: inner,
          default_value: default_val,
          has_default: has_default
        )
      end
    end
  end
end
