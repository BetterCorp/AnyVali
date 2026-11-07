import type { CoercionConfig } from "../types.js";

// Decimal floating-point grammar: optional sign, digits with optional
// fraction (or bare fraction), optional decimal exponent. Excludes hex/octal/
// binary literals, Infinity and NaN that JS `Number()` would otherwise accept.
const DECIMAL_FLOAT_RE = /^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$/;

export type CoercionResult =
  | { success: true; value: unknown }
  | { success: false; message: string };

/** Private sequence retained when a schema clones its coercion configuration. */
const ORDERED_COERCIONS = Symbol("anyvali.orderedCoercions");
type OrderedConfig = CoercionConfig & { [ORDERED_COERCIONS]?: string[] };
const ALIASES: Record<string, string> = { int: "string->int", number: "string->number", bool: "string->bool" };

/** Retain interchange operation order while keeping native flag configs supported. */
export function normalizeCoercionConfig(raw: unknown): CoercionConfig {
  if (typeof raw === "object" && raw !== null && !Array.isArray(raw)) return raw as CoercionConfig;
  const items = (Array.isArray(raw) ? raw : [raw]).map(item => ALIASES[String(item)] ?? String(item));
  if (items.some(item => !["string->int", "string->number", "string->bool", "trim", "lower", "upper"].includes(item))) {
    throw new Error("Unknown coercion identifier");
  }
  return { [ORDERED_COERCIONS]: items } as OrderedConfig;
}

/** Copy retained operations without exposing the private configuration symbol. */
export function orderedCoercions(config: CoercionConfig): string[] | undefined {
  return (config as OrderedConfig)[ORDERED_COERCIONS]?.slice();
}

/** Encode native flag configs in the order used by applyCoercion. */
export function portableCoercionTokens(config: CoercionConfig, targetType: string): string[] {
  const ordered = orderedCoercions(config);
  const conversions: Record<string, string> = {
    int: "string->int", int8: "string->int", int16: "string->int", int32: "string->int", int64: "string->int",
    uint8: "string->int", uint16: "string->int", uint32: "string->int", uint64: "string->int",
    number: "string->number", float32: "string->number", float64: "string->number", bool: "string->bool",
  };
  if (ordered) {
    const firstConversion = ordered.findIndex(operation => operation.startsWith("string->"));
    if (ordered.some(operation => operation.startsWith("string->") && operation !== conversions[targetType])) {
      throw new Error("Coercion conversion does not match its target kind");
    }
    if (targetType !== "string" && ordered.some((operation, index) =>
      ["trim", "lower", "upper"].includes(operation) && (firstConversion < 0 || index > firstConversion))) {
      throw new Error("String preprocessing must precede a typed conversion");
    }
    return ordered;
  }
  if (Object.keys(config).some(key => !["from", "trim", "lower", "upper"].includes(key))) throw new Error("Unknown portable coercion flag");
  if (config.from !== undefined && config.from !== "string") throw new Error("Unsupported portable coercion source");
  const tokens: string[] = [];
  for (const flag of ["trim", "lower", "upper"] as const) {
    if (config[flag] !== undefined && typeof config[flag] !== "boolean") throw new Error("Invalid coercion flag");
    if (config[flag]) tokens.push(flag);
  }
  if (tokens.length > 0 && targetType !== "string" && !conversions[targetType]) {
    throw new Error("String preprocessing requires a string target or a typed conversion");
  }
  if (conversions[targetType]) tokens.push(conversions[targetType]);
  return tokens;
}

/** Apply ordered portable operations or the native flag configuration. */
export function applyCoercion(
  input: unknown,
  config: CoercionConfig,
  targetType: string
): CoercionResult {
  let value = input;
  const ordered = orderedCoercions(config);
  if (ordered) {
    for (const operation of ordered) {
      const target = operation === "string->int" ? "int" : operation === "string->number" ? "number" : "bool";
      const step = operation.startsWith("string->")
        ? applyCoercion(value, { from: "string" }, target)
        : applyCoercion(value, { [operation]: true }, "string");
      if (!step.success) return step;
      value = step.value;
    }
    return { success: true, value };
  }

  // String transformations (trim, lower, upper) apply when input is a string
  if (typeof value === "string") {
    if (config.trim) {
      value = value.trim();
    }
    if (config.lower) {
      value = (value as string).toLowerCase();
    }
    if (config.upper) {
      value = (value as string).toUpperCase();
    }
  }

  // Type coercion from string to target.
  // The only portable coercion source is "string" (spec 5.1), so enabling
  // coercion on a non-string target (e.g. `number().coerce()`) implies a
  // string source even when `from` is omitted. Without this, a bare
  // `.coerce()` on a numeric/bool schema would silently no-op and the raw
  // string would fail validation with invalid_type.
  const isTypeTarget = targetType !== "string" && targetType !== "unknown";
  const fromString =
    config.from === "string" || (config.from === undefined && isTypeTarget);
  if (fromString && typeof value === "string") {
    switch (targetType) {
      case "int":
      case "int8":
      case "int16":
      case "int32":
      case "int64":
      case "uint8":
      case "uint16":
      case "uint32":
      case "uint64": {
        const trimmed = value.trim();
        if (trimmed === "" || !/^-?\d+$/.test(trimmed)) {
          return {
            success: false,
            message: `Cannot coerce "${value}" to ${targetType}`,
          };
        }
        const num = Number(trimmed);
        if (!Number.isFinite(num) || !Number.isInteger(num)) {
          return {
            success: false,
            message: `Cannot coerce "${value}" to ${targetType}`,
          };
        }
        value = num;
        break;
      }

      case "number":
      case "float32":
      case "float64": {
        const trimmed = value.trim();
        if (trimmed === "") {
          return {
            success: false,
            message: `Cannot coerce empty string to ${targetType}`,
          };
        }
        // Spec 5.1: parse as DECIMAL floating-point. JS `Number()` also accepts
        // hex (0x), octal (0o) and binary (0b) literals, which would let
        // "0x10" slip through as 16 and bypass the decimal-only contract.
        // Restrict to a decimal float grammar before parsing.
        if (!DECIMAL_FLOAT_RE.test(trimmed)) {
          return {
            success: false,
            message: `Cannot coerce "${value}" to ${targetType}`,
          };
        }
        const num = Number(trimmed);
        if (!Number.isFinite(num)) {
          return {
            success: false,
            message: `Cannot coerce "${value}" to ${targetType}`,
          };
        }
        value = num;
        break;
      }

      case "bool": {
        const lower = value.trim().toLowerCase();
        if (lower === "true" || lower === "1") {
          value = true;
        } else if (lower === "false" || lower === "0") {
          value = false;
        } else {
          return {
            success: false,
            message: `Cannot coerce "${value}" to bool`,
          };
        }
        break;
      }
    }
  }

  return { success: true, value };
}
