/**
 * Describe the type of a value for error messages, matching the corpus expectations.
 * null -> "null", array -> "array", otherwise typeof.
 */
export function describeType(value: unknown): string {
  if (value === null) return "null";
  if (Array.isArray(value)) return "array";
  return typeof value;
}
