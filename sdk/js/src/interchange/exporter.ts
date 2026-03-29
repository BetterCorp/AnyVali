import type { AnyValiDocument, ExportMode } from "../types.js";
import { BaseSchema } from "../schemas/base.js";

/**
 * Export a schema to a portable AnyValiDocument.
 */
export function exportSchema(
  schema: BaseSchema,
  mode: ExportMode = "portable"
): AnyValiDocument {
  return schema.export(mode);
}
