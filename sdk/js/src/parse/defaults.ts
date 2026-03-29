import { ABSENT, type Absent } from "../schemas/base.js";

/**
 * Apply a default value if the input is absent.
 * Returns the default value if input is absent, otherwise returns the input as-is.
 */
export function applyDefault<T>(
  input: unknown,
  defaultValue: T | Absent
): unknown {
  if ((input === undefined || input === ABSENT) && defaultValue !== ABSENT) {
    return defaultValue;
  }
  return input;
}
