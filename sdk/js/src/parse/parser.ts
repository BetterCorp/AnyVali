import type { ParseOptions, ParseResult } from "../types.js";
import { BaseSchema } from "../schemas/base.js";

/**
 * Parse input with the given schema. Throws ValidationError on failure.
 */
export function parse<T>(
  schema: BaseSchema<unknown, T>,
  input: unknown,
  options?: ParseOptions
): T {
  return schema.parse(input, options);
}

/**
 * Parse input with the given schema. Returns a result object.
 */
export function safeParse<T>(
  schema: BaseSchema<unknown, T>,
  input: unknown,
  options?: ParseOptions
): ParseResult<T> {
  return schema.safeParse(input, options);
}
