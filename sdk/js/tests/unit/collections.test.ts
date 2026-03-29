import { describe, it, expect } from "vitest";
import { array, tuple, record, string, int, bool } from "../../src/index.js";

describe("ArraySchema", () => {
  it("accepts valid arrays", () => {
    const s = array(int());
    expect(s.parse([1, 2, 3])).toEqual([1, 2, 3]);
  });

  it("rejects non-arrays", () => {
    const s = array(int());
    expect(s.safeParse("not an array").success).toBe(false);
  });

  it("validates array items", () => {
    const s = array(int());
    const result = s.safeParse([1, "two", 3]);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].path).toEqual([1]);
    }
  });

  it("validates minItems", () => {
    const s = array(int()).minItems(2);
    expect(s.parse([1, 2])).toEqual([1, 2]);
    expect(s.safeParse([1]).success).toBe(false);
  });

  it("validates maxItems", () => {
    const s = array(int()).maxItems(2);
    expect(s.parse([1, 2])).toEqual([1, 2]);
    expect(s.safeParse([1, 2, 3]).success).toBe(false);
  });
});

describe("TupleSchema", () => {
  it("accepts valid tuples", () => {
    const s = tuple([string(), int(), bool()]);
    expect(s.parse(["hello", 42, true])).toEqual(["hello", 42, true]);
  });

  it("rejects wrong length", () => {
    const s = tuple([string(), int()]);
    expect(s.safeParse(["hello"]).success).toBe(false);
    expect(s.safeParse(["hello", 42, true]).success).toBe(false);
  });

  it("validates element types", () => {
    const s = tuple([string(), int()]);
    const result = s.safeParse([42, "hello"]);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].path).toEqual([0]);
    }
  });
});

describe("RecordSchema", () => {
  it("accepts valid records", () => {
    const s = record(int());
    expect(s.parse({ a: 1, b: 2 })).toEqual({ a: 1, b: 2 });
  });

  it("rejects non-objects", () => {
    const s = record(int());
    expect(s.safeParse("not object").success).toBe(false);
    expect(s.safeParse(null).success).toBe(false);
    expect(s.safeParse([]).success).toBe(false);
  });

  it("validates record values", () => {
    const s = record(int());
    const result = s.safeParse({ a: 1, b: "two" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].path).toEqual(["b"]);
    }
  });
});
