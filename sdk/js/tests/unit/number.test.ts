import { describe, it, expect } from "vitest";
import { number, float32, float64, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, object } from "../../src/index.js";

describe("NumberSchema", () => {
  it("accepts valid numbers", () => {
    expect(number().parse(42)).toBe(42);
    expect(number().parse(3.14)).toBe(3.14);
    expect(number().parse(-0.5)).toBe(-0.5);
  });

  it("rejects non-numbers", () => {
    const result = number().safeParse("42");
    expect(result.success).toBe(false);
    if (!result.success) expect(result.issues[0].code).toBe("invalid_type");
  });

  it("rejects NaN and Infinity", () => {
    expect(number().safeParse(NaN).success).toBe(false);
    expect(number().safeParse(Infinity).success).toBe(false);
    expect(number().safeParse(-Infinity).success).toBe(false);
  });

  it("validates min", () => {
    const s = number().min(5);
    expect(s.parse(5)).toBe(5);
    expect(s.safeParse(4).success).toBe(false);
  });

  it("validates max", () => {
    const s = number().max(10);
    expect(s.parse(10)).toBe(10);
    expect(s.safeParse(11).success).toBe(false);
  });

  it("validates exclusiveMin", () => {
    const s = number().exclusiveMin(5);
    expect(s.parse(6)).toBe(6);
    expect(s.safeParse(5).success).toBe(false);
  });

  it("validates exclusiveMax", () => {
    const s = number().exclusiveMax(10);
    expect(s.parse(9)).toBe(9);
    expect(s.safeParse(10).success).toBe(false);
  });

  it("validates multipleOf", () => {
    const s = number().multipleOf(3);
    expect(s.parse(9)).toBe(9);
    expect(s.safeParse(10).success).toBe(false);
  });

  it("coerces string to number", () => {
    const s = number().coerce({ from: "string" });
    expect(s.parse("3.14")).toBe(3.14);
  });

  it("coercion failure", () => {
    const s = number().coerce({ from: "string" });
    const result = s.safeParse("abc");
    expect(result.success).toBe(false);
    if (!result.success) expect(result.issues[0].code).toBe("coercion_failed");
  });

  // Bare `.coerce()` (no args) must imply string source on a numeric target.
  // Regression: previously no-op'd, so a string input failed with invalid_type.
  it("coerces string to number with no-arg coerce()", () => {
    const s = number().coerce();
    expect(s.parse("3.14")).toBe(3.14);
  });

  // Full string->number coercion matrix (no-arg form). ASCII decimal float
  // incl. exponent, trimmed. No hex/Infinity/NaN/underscores.
  describe("string->number matrix via no-arg coerce()", () => {
    const s = number().coerce();

    it.each([
      ["3.14", 3.14],
      ["-1.5e3", -1500],
      ["  2  ", 2],
      ["0", 0],
    ])("accepts %j -> %j", (input, expected) => {
      expect(s.parse(input as string)).toBe(expected);
    });

    it.each([
      "0x10",
      "Infinity",
      "NaN",
      "",
      "1_000",
      "abc",
    ])("rejects %j with coercion_failed", (input) => {
      const result = s.safeParse(input);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.issues[0].code).toBe("coercion_failed");
      }
    });
  });
});

describe("Float32Schema", () => {
  it("accepts valid numbers", () => {
    expect(float32().parse(1.5)).toBe(1.5);
  });
});

describe("Float64Schema", () => {
  it("accepts valid numbers", () => {
    expect(float64().parse(1.5)).toBe(1.5);
  });
});

describe("IntSchema", () => {
  it("accepts integers", () => {
    expect(int().parse(42)).toBe(42);
    expect(int().parse(-10)).toBe(-10);
    expect(int().parse(0)).toBe(0);
  });

  it("rejects floats", () => {
    const result = int().safeParse(3.14);
    expect(result.success).toBe(false);
    if (!result.success) expect(result.issues[0].code).toBe("invalid_type");
  });

  it("rejects non-numbers", () => {
    expect(int().safeParse("42").success).toBe(false);
  });

  it("coerces string to int", () => {
    const s = int().coerce({ from: "string" });
    expect(s.parse("42")).toBe(42);
  });

  it("rejects float strings when coercing to int", () => {
    const s = int().coerce({ from: "string" });
    const result = s.safeParse("3.14");
    expect(result.success).toBe(false);
  });

  it("coerces string to int with no-arg coerce()", () => {
    const s = int().coerce();
    expect(s.parse("42")).toBe(42);
  });

  // Full string->int coercion matrix (no-arg form). ASCII `^-?\d+$`, trimmed.
  describe("string->int matrix via no-arg coerce()", () => {
    const s = int().coerce();

    it.each([
      ["42", 42],
      ["  42  ", 42],
      ["-7", -7],
    ])("accepts %j -> %j", (input, expected) => {
      expect(s.parse(input as string)).toBe(expected);
    });

    it.each([
      "3.14",
      "0x10",
      "1_000",
      "+5",
      "Infinity",
      "",
      "abc",
    ])("rejects %j with coercion_failed", (input) => {
      const result = s.safeParse(input);
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.issues[0].code).toBe("coercion_failed");
      }
    });
  });
});

// Reproduces the reported field-level coercion failure: an object whose
// numeric fields use bare `.coerce()` must coerce string inputs, not reject
// them with invalid_type.
describe("object with no-arg coerce() on numeric fields", () => {
  it("coerces all string fields to numbers", () => {
    const schema = object({
      lumpSum: number().coerce().min(0),
      monthlyContributions: number().coerce().min(0),
      investmentTerm: number().coerce().min(1),
    });
    const result = schema.safeParse({
      lumpSum: "1000000",
      monthlyContributions: "1000",
      investmentTerm: "20",
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toEqual({
        lumpSum: 1000000,
        monthlyContributions: 1000,
        investmentTerm: 20,
      });
    }
  });
});

describe("Int8Schema", () => {
  it("accepts values in range [-128, 127]", () => {
    expect(int8().parse(0)).toBe(0);
    expect(int8().parse(-128)).toBe(-128);
    expect(int8().parse(127)).toBe(127);
  });

  it("rejects out-of-range values", () => {
    expect(int8().safeParse(128).success).toBe(false);
    expect(int8().safeParse(-129).success).toBe(false);
  });
});

describe("Int16Schema", () => {
  it("accepts values in range", () => {
    expect(int16().parse(0)).toBe(0);
    expect(int16().parse(-32768)).toBe(-32768);
    expect(int16().parse(32767)).toBe(32767);
  });

  it("rejects out-of-range values", () => {
    expect(int16().safeParse(32768).success).toBe(false);
  });
});

describe("Int32Schema", () => {
  it("accepts values in range", () => {
    expect(int32().parse(0)).toBe(0);
    expect(int32().parse(2147483647)).toBe(2147483647);
    expect(int32().parse(-2147483648)).toBe(-2147483648);
  });

  it("rejects out-of-range values", () => {
    expect(int32().safeParse(2147483648).success).toBe(false);
  });
});

describe("Int64Schema", () => {
  it("accepts safe integers", () => {
    expect(int64().parse(0)).toBe(0);
    expect(int64().parse(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
    expect(int64().parse(Number.MIN_SAFE_INTEGER)).toBe(Number.MIN_SAFE_INTEGER);
  });
});

describe("Uint8Schema", () => {
  it("accepts values in range [0, 255]", () => {
    expect(uint8().parse(0)).toBe(0);
    expect(uint8().parse(255)).toBe(255);
  });

  it("rejects negative and out-of-range", () => {
    expect(uint8().safeParse(-1).success).toBe(false);
    expect(uint8().safeParse(256).success).toBe(false);
  });
});

describe("Uint16Schema", () => {
  it("accepts values in range [0, 65535]", () => {
    expect(uint16().parse(0)).toBe(0);
    expect(uint16().parse(65535)).toBe(65535);
  });

  it("rejects out-of-range", () => {
    expect(uint16().safeParse(65536).success).toBe(false);
  });
});

describe("Uint32Schema", () => {
  it("accepts values in range [0, 4294967295]", () => {
    expect(uint32().parse(0)).toBe(0);
    expect(uint32().parse(4294967295)).toBe(4294967295);
  });

  it("rejects out-of-range", () => {
    expect(uint32().safeParse(4294967296).success).toBe(false);
  });
});

describe("Uint64Schema", () => {
  it("accepts values in range [0, MAX_SAFE_INTEGER]", () => {
    expect(uint64().parse(0)).toBe(0);
    expect(uint64().parse(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
  });

  it("rejects negative values", () => {
    expect(uint64().safeParse(-1).success).toBe(false);
  });
});
