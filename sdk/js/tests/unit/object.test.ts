import { describe, it, expect } from "vitest";
import { object, string, int, optional, bool } from "../../src/index.js";

describe("ObjectSchema", () => {
  it("accepts valid objects", () => {
    const s = object({
      name: string(),
      age: int(),
    });
    const result = s.parse({ name: "Alice", age: 30 });
    expect(result).toEqual({ name: "Alice", age: 30 });
  });

  it("rejects missing required fields", () => {
    const s = object({
      name: string(),
      age: int(),
    });
    const result = s.safeParse({ name: "Alice" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("required");
      expect(result.issues[0].path).toEqual(["age"]);
    }
  });

  it("rejects unknown keys by default", () => {
    const s = object({ name: string() });
    const result = s.safeParse({ name: "Alice", extra: true });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("unknown_key");
    }
  });

  it("strips unknown keys when configured", () => {
    const s = object({ name: string() }).unknownKeys("strip");
    const result = s.parse({ name: "Alice", extra: true });
    expect(result).toEqual({ name: "Alice" });
  });

  it("allows unknown keys when configured", () => {
    const s = object({ name: string() }).unknownKeys("allow");
    const result = s.parse({ name: "Alice", extra: true });
    expect(result).toEqual({ name: "Alice", extra: true });
  });

  it("handles optional fields", () => {
    const s = object({
      name: string(),
      nickname: optional(string()),
    });
    const result = s.parse({ name: "Alice" });
    expect(result).toEqual({ name: "Alice" });
  });

  it("keeps object modifiers immutable", () => {
    const base = object({ name: string() });
    const stripped = base.unknownKeys("strip");

    expect(base.safeParse({ name: "Alice", extra: true }).success).toBe(false);
    expect(stripped.parse({ name: "Alice", extra: true })).toEqual({
      name: "Alice",
    });
  });

  it("handles defaults on fields", () => {
    const s = object({
      name: string(),
      role: string().default("user"),
    });
    const result = s.parse({ name: "Alice" });
    expect(result).toEqual({ name: "Alice", role: "user" });
  });

  it("does not overwrite present values with defaults", () => {
    const s = object({
      role: string().default("user"),
    });
    const result = s.parse({ role: "admin" });
    expect(result).toEqual({ role: "admin" });
  });

  it("validates nested objects", () => {
    const s = object({
      user: object({
        name: string(),
      }),
    });
    const result = s.safeParse({ user: { name: 42 } });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].path).toEqual(["user", "name"]);
    }
  });

  it("collects multiple nested issues with precise paths", () => {
    const s = object({
      user: object({
        name: string().minLength(2),
        age: int().min(18),
      }),
      active: bool(),
    });

    const result = s.safeParse({
      user: { name: "", age: 12 },
      active: "yes",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues.map((issue) => issue.path)).toEqual(
        expect.arrayContaining([
          ["user", "name"],
          ["user", "age"],
          ["active"],
        ])
      );
    }
  });

  it("treats inherited properties as absent", () => {
    const base = { name: "Alice" };
    const input = Object.create(base) as Record<string, unknown>;
    input.age = 30;

    const s = object({
      name: string(),
      age: int(),
    });

    const result = s.safeParse(input);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            code: "required",
            path: ["name"],
          }),
        ])
      );
    }
  });

  it("returns a detached parsed object", () => {
    const s = object({
      user: object({
        name: string(),
      }),
    });

    const input = { user: { name: "Alice" } };
    const result = s.parse(input);

    expect(result).toEqual(input);
    expect(result).not.toBe(input);
    expect(result.user).not.toBe(input.user);
  });

  it("rejects non-objects", () => {
    const s = object({ name: string() });
    expect(s.safeParse("string").success).toBe(false);
    expect(s.safeParse(null).success).toBe(false);
    expect(s.safeParse([]).success).toBe(false);
  });
});
