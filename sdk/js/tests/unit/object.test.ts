import { describe, it, expect } from "vitest";
import { object, string, int, optional, bool, RefSchema, parse } from "../../src/index.js";

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

  it("strips unknown keys by default", () => {
    const s = object({ name: string() });
    const result = s.parse({ name: "Alice", extra: true });
    expect(result).toEqual({ name: "Alice" });
  });

  it("rejects unknown keys when configured", () => {
    const s = object({ name: string() }).unknownKeys("reject");
    const result = s.safeParse({ name: "Alice", extra: true });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("unknown_key");
    }
  });

  it("allows unknown keys when configured", () => {
    const s = object({ name: string() }).unknownKeys("allow");
    const result = s.parse({ name: "Alice", extra: true });
    expect(result).toEqual({ name: "Alice", extra: true });
  });

  it("parse-level strip overrides object-level allow recursively", () => {
    const s = object({
      user: object({ name: string() }).unknownKeys("allow"),
    }).unknownKeys("allow");

    expect(
      s.parse(
        { user: { name: "Alice", role: "admin" }, rootExtra: true },
        { unknownKeys: "strip" }
      )
    ).toEqual({ user: { name: "Alice" } });
  });

  it("top-level parse unknownKeys overrides schema unknownKeys", () => {
    const s = object({ name: string() }).unknownKeys("strip");

    expect(() =>
      parse(s, { name: "Alice", extra: true }, { unknownKeys: "reject" })
    ).toThrow("unknown_key");
  });

  it("parent strip overrides nested allow", () => {
    const s = object({
      user: object({ name: string() }).unknownKeys("allow"),
    }).unknownKeys("strip");

    expect(s.parse({ user: { name: "Alice", role: "admin" } })).toEqual({
      user: { name: "Alice" },
    });
  });

  it("parent reject overrides nested allow", () => {
    const s = object({
      user: object({ name: string() }).unknownKeys("allow"),
    }).unknownKeys("reject");

    const result = s.safeParse({ user: { name: "Alice", role: "admin" } });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0]).toEqual(
        expect.objectContaining({
          code: "unknown_key",
          path: ["user", "role"],
        })
      );
    }
  });

  it("parent allow leaves nested unknown key policy alone", () => {
    const s = object({
      user: object({ name: string() }).unknownKeys("strip"),
    }).unknownKeys("allow");

    expect(
      s.parse({ user: { name: "Alice", role: "admin" }, rootExtra: true })
    ).toEqual({ user: { name: "Alice" }, rootExtra: true });
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
    const rejected = base.unknownKeys("reject");

    expect(base.parse({ name: "Alice", extra: true })).toEqual({
      name: "Alice",
    });
    expect(rejected.safeParse({ name: "Alice", extra: true }).success).toBe(
      false
    );
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

  it("preserves __proto__ as data when unknown keys are allowed", () => {
    const s = object({ name: string() }).unknownKeys("allow");
    const input = JSON.parse(
      '{"name":"Alice","__proto__":{"polluted":"yes"}}'
    ) as Record<string, unknown>;

    const result = s.parse(input) as Record<string, unknown>;

    expect(result.name).toBe("Alice");
    expect(Object.getPrototypeOf(result)).toBe(Object.prototype);
    expect(Object.prototype.hasOwnProperty.call(result, "__proto__")).toBe(true);
    expect(
      Object.getOwnPropertyDescriptor(result, "__proto__")?.value
    ).toEqual({ polluted: "yes" });
    expect(({} as Record<string, unknown>).polluted).toBeUndefined();
  });

  it("fails closed on cyclic input for recursive schemas", () => {
    let node!: ReturnType<typeof object>;
    const next = optional(new RefSchema("#/definitions/Node", () => node));
    node = object({
      value: string(),
      next,
    });

    const input: Record<string, unknown> = { value: "root" };
    input.next = input;

    let result:
      | ReturnType<typeof node.safeParse>
      | undefined;

    expect(() => {
      result = node.safeParse(input);
    }).not.toThrow();
    expect(result?.success).toBe(false);
  });

  it("rejects non-objects", () => {
    const s = object({ name: string() });
    expect(s.safeParse("string").success).toBe(false);
    expect(s.safeParse(null).success).toBe(false);
    expect(s.safeParse([]).success).toBe(false);
  });
});
