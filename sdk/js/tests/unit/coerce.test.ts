import { describe, it, expect } from "vitest";
import {
  applyCoercion,
  normalizeCoercionConfig,
} from "../../src/parse/coerce.js";

describe("normalizeCoercionConfig", () => {
  it("passes native object config through unchanged", () => {
    const config = { from: "string", trim: true };
    expect(normalizeCoercionConfig(config)).toBe(config);
  });
  it.each([
    ["trim", " x ", "x"], ["lower", "X", "x"], ["upper", "x", "X"],
    ["string->int", "42", 42], ["string->number", "2.5", 2.5], ["string->bool", "true", true],
  ])("executes the declared %s operation", (operation, input, value) => {
    expect(applyCoercion(input, normalizeCoercionConfig(operation), "unknown")).toEqual({ success: true, value });
  });
  it("retains operation order and repeated operations", () => {
    expect(applyCoercion(" X ", normalizeCoercionConfig(["trim", "upper", "lower"]), "string"))
      .toEqual({ success: true, value: "x" });
    expect(applyCoercion("x", normalizeCoercionConfig(["upper", "lower", "upper"]), "string"))
      .toEqual({ success: true, value: "X" });
  });
  it("keeps an empty operation list as a no-op", () => {
    expect(applyCoercion("7", normalizeCoercionConfig([]), "int")).toEqual({ success: true, value: "7" });
  });
  it("rejects unknown operations rather than silently dropping them", () => {
    expect(() => normalizeCoercionConfig(["unknown-operation"])).toThrow();
  });
});

describe("applyCoercion edge cases", () => {
  it("applies trim to string input", () => {
    const result = applyCoercion("  hello  ", { trim: true }, "string");
    expect(result).toEqual({ success: true, value: "hello" });
  });

  it("applies lower to string input", () => {
    const result = applyCoercion("HELLO", { lower: true }, "string");
    expect(result).toEqual({ success: true, value: "hello" });
  });

  it("applies upper to string input", () => {
    const result = applyCoercion("hello", { upper: true }, "string");
    expect(result).toEqual({ success: true, value: "HELLO" });
  });

  it("applies trim + lower together", () => {
    const result = applyCoercion("  HELLO  ", { trim: true, lower: true }, "string");
    expect(result).toEqual({ success: true, value: "hello" });
  });

  it("fails coercing non-integer string to int", () => {
    const result = applyCoercion("3.5", { from: "string" }, "int");
    expect(result.success).toBe(false);
  });

  it("fails coercing non-finite string to int", () => {
    // A very large number that becomes Infinity
    const huge = "9".repeat(400);
    const result = applyCoercion(huge, { from: "string" }, "int");
    expect(result.success).toBe(false);
  });

  it("fails coercing empty string to float", () => {
    const result = applyCoercion("", { from: "string" }, "number");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.message).toContain("empty string");
    }
  });

  it("fails coercing empty string to float32", () => {
    const result = applyCoercion("", { from: "string" }, "float32");
    expect(result.success).toBe(false);
  });

  it("fails coercing empty string to float64", () => {
    const result = applyCoercion("", { from: "string" }, "float64");
    expect(result.success).toBe(false);
  });

  it("fails coercing non-numeric string to number", () => {
    const result = applyCoercion("abc", { from: "string" }, "number");
    expect(result.success).toBe(false);
  });

  it("successfully coerces valid float string", () => {
    const result = applyCoercion("3.14", { from: "string" }, "number");
    expect(result).toEqual({ success: true, value: 3.14 });
  });

  it("coerces string to int subtypes", () => {
    for (const kind of [
      "int8", "int16", "int32", "int64",
      "uint8", "uint16", "uint32", "uint64",
    ]) {
      const result = applyCoercion("42", { from: "string" }, kind);
      expect(result).toEqual({ success: true, value: 42 });
    }
  });

  it("coerces string to float32/float64", () => {
    for (const kind of ["float32", "float64"]) {
      const result = applyCoercion("1.5", { from: "string" }, kind);
      expect(result).toEqual({ success: true, value: 1.5 });
    }
  });

  it("does not coerce when input is not a string", () => {
    const result = applyCoercion(42, { from: "string" }, "int");
    expect(result).toEqual({ success: true, value: 42 });
  });

  it("does not apply trim/lower/upper when input is not a string", () => {
    const result = applyCoercion(42, { trim: true, lower: true, upper: true }, "number");
    expect(result).toEqual({ success: true, value: 42 });
  });
});
