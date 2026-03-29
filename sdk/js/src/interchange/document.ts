import type { AnyValiDocument } from "../types.js";

export const ANYVALI_VERSION = "1.0";
export const SCHEMA_VERSION = "1";

export function createDocument(
  root: import("../types.js").SchemaNode,
  definitions: Record<string, import("../types.js").SchemaNode> = {},
  extensions: Record<string, Record<string, unknown>> = {}
): AnyValiDocument {
  return {
    anyvaliVersion: ANYVALI_VERSION,
    schemaVersion: SCHEMA_VERSION,
    root,
    definitions,
    extensions,
  };
}
