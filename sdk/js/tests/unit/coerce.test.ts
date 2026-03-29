import { describe, it, expect } from "vitest";
import {
  applyCoercion,
  normalizeCoercionConfig,
} from "../../src/parse/coerce.js";

describe("normalizeCoercionConfig", () => {
  it("passes through object config as-is", () => {
    const config = { from: "string", trim: true };
    expect(normalizeCoercionConfig(config)).toBe(config);
  });

  it("converts single string 'trim' to config", () => {
    expect(normalizeCoercionConfig("trim")).toEqual({ trim: true });
  });

  it("converts single string 'lower' to config", () => {
    expect(normalizeCoercionConfig("lower")).toEqual({ lower: true });
  });

  it("converts single string 'upper' to config", () => {
    expect(normalizeCoercionConfig("upper")).toEqual({ upper: true });
  });

  it("converts 'string->int' to from config", () => {
    expect(normalizeCoercionConfig("string->int")).toEqual({ from: "string" });
  });

  it("converts 'string->number' to from config", () => {
    expect(normalizeCoercionConfig("string->number")).toEqual({
      from: "string",
    });
  });

  it("converts 'string->bool' to from config", () => {
    expect(normalizeCoercionConfig("string->bool")).toEqual({
      from: "string",
    });
  });

  it("converts array of strings", () => {
    expect(normalizeCoercionConfig(["trim", "lower"])).toEqual({
      trim: true,
      lower: true,
    });
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
