import { describe, it, expect } from "vitest";
import { ValidationError } from "../../src/errors.js";
import type { ValidationIssue } from "../../src/types.js";

describe("ValidationError", () => {
  it("constructs with issues and formats message", () => {
    const issues: ValidationIssue[] = [
      {
        code: "invalid_type",
        message: "Expected string, received number",
        path: ["name"],
        expected: "string",
        received: "number",
      },
    ];
    const err = new ValidationError(issues);
    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe("ValidationError");
    expect(err.issues).toBe(issues);
    expect(err.message).toBe("[invalid_type] name: Expected string, received number");
  });

  it("formats message for root path (empty path)", () => {
    const issues: ValidationIssue[] = [
      {
        code: "invalid_type",
        message: "Expected string",
        path: [],
      },
    ];
    const err = new ValidationError(issues);
    expect(err.message).toBe("[invalid_type] Expected string");
  });

  it("formats multiple issues joined by newline", () => {
    const issues: ValidationIssue[] = [
      { code: "invalid_type", message: "bad type", path: ["a"] },
      { code: "too_small", message: "too short", path: ["b", "c"] },
    ];
    const err = new ValidationError(issues);
    expect(err.message).toBe(
      "[invalid_type] a: bad type\n[too_small] b.c: too short"
    );
  });

  it("formats nested path with dot notation", () => {
    const issues: ValidationIssue[] = [
      { code: "required", message: "required", path: ["x", 0, "y"] },
    ];
    const err = new ValidationError(issues);
    expect(err.message).toBe("[required] x.0.y: required");
  });
});
