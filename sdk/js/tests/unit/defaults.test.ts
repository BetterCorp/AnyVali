import { describe, it, expect } from "vitest";
import { string, int, object, optional } from "../../src/index.js";

describe("Defaults", () => {
  it("applies default when value is absent", () => {
    const s = string().default("fallback");
    expect(s.parse(undefined)).toBe("fallback");
  });

  it("does not apply default when value is present", () => {
    const s = string().default("fallback");
    expect(s.parse("provided")).toBe("provided");
  });

  it("validates the defaulted value", () => {
    const s = string().minLength(5).default("hi");
    // default "hi" is only 2 chars, should fail validation with default_invalid
    const result = s.safeParse(undefined);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("default_invalid");
    }
  });

  it("applies defaults in object fields", () => {
    const s = object({
      name: string(),
      role: string().default("user"),
    });
    const result = s.parse({ name: "Alice" });
    expect(result.role).toBe("user");
  });

  it("does not override present object field with default", () => {
    const s = object({
      role: string().default("user"),
    });
    const result = s.parse({ role: "admin" });
    expect(result.role).toBe("admin");
  });

  it("default on optional field", () => {
    const s = object({
      count: optional(int().default(0)),
    });
    const result = s.parse({});
    expect(result).toEqual({ count: 0 });
  });
});
