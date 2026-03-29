import type { ParseResult } from "../types.js";
import { BaseSchema } from "../schemas/base.js";

/**
 * Parse input with the given schema. Throws ValidationError on failure.
 */
export function parse<T>(schema: BaseSchema<unknown, T>, input: unknown): T {
  return schema.parse(input);
}

/**
 * Parse input with the given schema. Returns a result object.
 */
export function safeParse<T>(
  schema: BaseSchema<unknown, T>,
  input: unknown
): ParseResult<T> {
  return schema.safeParse(input);
}
