import { describe, it, expect } from "vitest";
import {
  union,
  intersection,
  optional,
  nullable,
  string,
  int,
  object,
  literal,
} from "../../src/index.js";

describe("UnionSchema", () => {
  it("accepts values matching any variant", () => {
    const s = union([string(), int()]);
    expect(s.parse("hello")).toBe("hello");
    expect(s.parse(42)).toBe(42);
  });

  it("rejects values matching no variant", () => {
    const s = union([string(), int()]);
    const result = s.safeParse(true);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("invalid_union");
    }
  });

  it("returns first matching variant result", () => {
    const s = union([literal("a"), string()]);
    expect(s.parse("a")).toBe("a");
    expect(s.parse("b")).toBe("b");
  });
});

describe("IntersectionSchema", () => {
  it("validates against all schemas", () => {
    const s = intersection([
      object({ name: string() }).unknownKeys("strip"),
      object({ age: int() }).unknownKeys("strip"),
    ]);
    const result = s.parse({ name: "Alice", age: 30 });
    expect(result).toEqual({ name: "Alice", age: 30 });
  });

  it("fails if any schema fails", () => {
    const s = intersection([
      object({ name: string() }).unknownKeys("strip"),
      object({ age: int() }).unknownKeys("strip"),
    ]);
    const result = s.safeParse({ name: "Alice" });
    expect(result.success).toBe(false);
  });
});

describe("OptionalSchema", () => {
  it("accepts undefined", () => {
    const s = optional(string());
    expect(s.parse(undefined)).toBe(undefined);
  });

  it("validates when present", () => {
    const s = optional(string());
    expect(s.parse("hello")).toBe("hello");
    expect(s.safeParse(42).success).toBe(false);
  });
});

describe("NullableSchema", () => {
  it("accepts null", () => {
    const s = nullable(string());
    expect(s.parse(null)).toBe(null);
  });

  it("validates when not null", () => {
    const s = nullable(string());
    expect(s.parse("hello")).toBe("hello");
    expect(s.safeParse(42).success).toBe(false);
  });
});
