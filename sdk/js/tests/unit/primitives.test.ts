import { describe, it, expect } from "vitest";
import {
  bool,
  null_,
  any,
  unknown,
  never,
  literal,
  enum_,
} from "../../src/index.js";

describe("BoolSchema", () => {
  it("accepts booleans", () => {
    expect(bool().parse(true)).toBe(true);
    expect(bool().parse(false)).toBe(false);
  });

  it("rejects non-booleans", () => {
    expect(bool().safeParse("true").success).toBe(false);
    expect(bool().safeParse(1).success).toBe(false);
    expect(bool().safeParse(null).success).toBe(false);
  });

  it("coerces string to bool", () => {
    const s = bool().coerce({ from: "string" });
    expect(s.parse("true")).toBe(true);
    expect(s.parse("false")).toBe(false);
    expect(s.parse("1")).toBe(true);
    expect(s.parse("0")).toBe(false);
  });

  it("coercion failure for invalid string", () => {
    const s = bool().coerce({ from: "string" });
    const result = s.safeParse("maybe");
    expect(result.success).toBe(false);
    if (!result.success) expect(result.issues[0].code).toBe("coercion_failed");
  });
});

describe("NullSchema", () => {
  it("accepts null", () => {
    expect(null_().parse(null)).toBe(null);
  });

  it("rejects non-null", () => {
    expect(null_().safeParse(undefined).success).toBe(false);
    expect(null_().safeParse(0).success).toBe(false);
    expect(null_().safeParse("").success).toBe(false);
  });
});

describe("AnySchema", () => {
  it("accepts anything", () => {
    expect(any().parse(42)).toBe(42);
    expect(any().parse("hello")).toBe("hello");
    expect(any().parse(null)).toBe(null);
    expect(any().parse(undefined)).toBe(undefined);
  });
});

describe("UnknownSchema", () => {
  it("accepts anything", () => {
    expect(unknown().parse(42)).toBe(42);
    expect(unknown().parse("hello")).toBe("hello");
  });
});

describe("NeverSchema", () => {
  it("rejects everything", () => {
    expect(never().safeParse(42).success).toBe(false);
    expect(never().safeParse(null).success).toBe(false);
    expect(never().safeParse(undefined).success).toBe(false);
  });
});

describe("LiteralSchema", () => {
  it("accepts matching literal", () => {
    expect(literal("hello").parse("hello")).toBe("hello");
    expect(literal(42).parse(42)).toBe(42);
    expect(literal(true).parse(true)).toBe(true);
    expect(literal(null).parse(null)).toBe(null);
  });

  it("rejects non-matching values", () => {
    expect(literal("hello").safeParse("world").success).toBe(false);
    expect(literal(42).safeParse(43).success).toBe(false);
    if (!literal(42).safeParse(43).success) {
      expect((literal(42).safeParse(43) as any).issues[0].code).toBe("invalid_literal");
    }
  });
});

describe("EnumSchema", () => {
  it("accepts values in the enum", () => {
    const s = enum_(["a", "b", "c"]);
    expect(s.parse("a")).toBe("a");
    expect(s.parse("b")).toBe("b");
  });

  it("rejects values not in the enum", () => {
    const s = enum_(["a", "b", "c"]);
    expect(s.safeParse("d").success).toBe(false);
    expect(s.safeParse(1).success).toBe(false);
  });

  it("works with numeric enums", () => {
    const s = enum_([1, 2, 3]);
    expect(s.parse(1)).toBe(1);
    expect(s.safeParse(4).success).toBe(false);
  });
});
