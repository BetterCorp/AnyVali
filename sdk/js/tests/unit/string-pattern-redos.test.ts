import { describe, it, expect } from "vitest";
import { string, array } from "../../src/index.js";

describe("String pattern: caching + ReDoS defense-in-depth", () => {
  it("matches and rejects correctly (behavior preserved by caching)", () => {
    const s = string().pattern("^[a-z]+$");
    expect(s.safeParse("abc").success).toBe(true);
    const bad = s.safeParse("ABC");
    expect(bad.success).toBe(false);
    if (!bad.success) expect(bad.issues[0].code).toBe("invalid_string");
  });

  it("cached regex stays correct when reused across many values", () => {
    const s = array(string().pattern("^\\d{3}$"));
    const r = s.safeParse(["123", "456", "789"]);
    expect(r.success).toBe(true);
    const r2 = s.safeParse(["123", "xx", "789"]);
    expect(r2.success).toBe(false);
  });

  it("invalid pattern is reported as invalid_string", () => {
    const s = string().pattern("(");
    const r = s.safeParse("anything");
    expect(r.success).toBe(false);
    if (!r.success) {
      expect(r.issues[0].code).toBe("invalid_string");
      expect(r.issues[0].message).toContain("Invalid regex pattern");
    }
  });

  it("re-patterning resets the compiled cache", () => {
    const base = string().pattern("^a$");
    const repat = base.pattern("^b$");
    expect(base.safeParse("a").success).toBe(true);
    expect(repat.safeParse("b").success).toBe(true);
    expect(repat.safeParse("a").success).toBe(false);
  });

  it("long input with a safe (linear) pattern still validates", () => {
    // Caching must not change behavior for legitimate large inputs.
    const s = string().pattern("^x+$");
    const big = "x".repeat(1_000_000);
    const r = s.safeParse(big);
    expect(r.success).toBe(true);
  });
});
