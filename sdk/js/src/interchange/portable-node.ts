import type { SchemaNode, CoercionConfig } from "../types.js";
import { normalizeCoercionConfig, portableCoercionTokens } from "../parse/coerce.js";

/** Resolve coercion targets through portable/legacy wrappers and definition refs. */
export function coercionTargetKind(node: Record<string, any>, definitions: Record<string, SchemaNode>, limit: number): string {
  for (let depth = 0; depth < limit; depth++) {
    if (node.kind === "optional" || node.kind === "nullable") node = node.schema ?? node.inner;
    else if (node.kind === "ref") {
      if (typeof node.ref !== "string" || !node.ref.startsWith("#/definitions/")) return "unknown";
      node = definitions[node.ref.slice("#/definitions/".length)];
    }
    else return node.kind;
    if (!node) return "unknown";
  }
  return "unknown";
}

/** Normalize schema edges/coercions only; metadata, literals and defaults remain data. */
export function portableNode(node: SchemaNode, definitions: Record<string, SchemaNode>, limit: number, depth = 0): SchemaNode {
  if (depth >= limit) throw new Error("Maximum portable schema depth exceeded");
  const source = node as unknown as Record<string, any>;
  const result = { ...source };
  const child = (value: SchemaNode) => portableNode(value, definitions, limit, depth + 1);
  if (source.coerce !== undefined) {
    const config = normalizeCoercionConfig(source.coerce) as CoercionConfig;
    result.coerce = portableCoercionTokens(config, coercionTargetKind(source, definitions, limit));
  }
  switch (source.kind) {
    case "optional": case "nullable":
      result.schema = child(source.schema ?? source.inner); delete result.inner; break;
    case "record":
      result.values = child(source.values ?? source.valueSchema); delete result.valueSchema; break;
    case "array": result.items = child(source.items); break;
    case "tuple":
      result.elements = (source.elements ?? source.items).map(child); delete result.items; break;
    case "object":
      result.properties = Object.fromEntries(Object.entries(source.properties ?? {}).map(([key, value]) => [key, child(value as SchemaNode)])); break;
    case "union": result.variants = source.variants.map(child); break;
    case "intersection":
      result.allOf = (source.allOf ?? source.schemas).map(child); delete result.schemas; break;
  }
  return result as SchemaNode;
}
