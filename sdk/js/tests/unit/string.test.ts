import { describe, it, expect } from "vitest";
import { string } from "../../src/index.js";

describe("StringSchema", () => {
  it("accepts valid strings", () => {
    const s = string();
    expect(s.parse("hello")).toBe("hello");
    expect(s.parse("")).toBe("");
  });

  it("rejects non-strings", () => {
    const s = string();
    const result = s.safeParse(42);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("invalid_type");
    }
  });

  it("validates minLength", () => {
    const s = string().minLength(3);
    expect(s.parse("abc")).toBe("abc");
    const result = s.safeParse("ab");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("too_small");
    }
  });

  it("validates maxLength", () => {
    const s = string().maxLength(3);
    expect(s.parse("abc")).toBe("abc");
    const result = s.safeParse("abcd");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("too_large");
    }
  });

  it("validates pattern", () => {
    const s = string().pattern("^[a-z]+$");
    expect(s.parse("abc")).toBe("abc");
    const result = s.safeParse("ABC");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("invalid_string");
    }
  });

  it("validates startsWith", () => {
    const s = string().startsWith("hello");
    expect(s.parse("hello world")).toBe("hello world");
    const result = s.safeParse("world hello");
    expect(result.success).toBe(false);
  });

  it("validates endsWith", () => {
    const s = string().endsWith("world");
    expect(s.parse("hello world")).toBe("hello world");
    const result = s.safeParse("world hello");
    expect(result.success).toBe(false);
  });

  it("validates includes", () => {
    const s = string().includes("mid");
    expect(s.parse("a mid b")).toBe("a mid b");
    const result = s.safeParse("no match");
    expect(result.success).toBe(false);
  });

  it("validates email format", () => {
    const s = string().format("email");
    expect(s.parse("user@example.com")).toBe("user@example.com");
    const result = s.safeParse("notanemail");
    expect(result.success).toBe(false);
  });

  it("validates uuid format", () => {
    const s = string().format("uuid");
    expect(s.parse("550e8400-e29b-41d4-a716-446655440000")).toBe(
      "550e8400-e29b-41d4-a716-446655440000"
    );
    const result = s.safeParse("not-a-uuid");
    expect(result.success).toBe(false);
  });

  it("validates url format", () => {
    const s = string().format("url");
    expect(s.parse("https://example.com")).toBe("https://example.com");
    const result = s.safeParse("not a url");
    expect(result.success).toBe(false);
  });

  it("validates ipv4 format", () => {
    const s = string().format("ipv4");
    expect(s.parse("192.168.1.1")).toBe("192.168.1.1");
    const result = s.safeParse("999.999.999.999");
    expect(result.success).toBe(false);
  });

  it("validates date format", () => {
    const s = string().format("date");
    expect(s.parse("2024-01-15")).toBe("2024-01-15");
    const result = s.safeParse("2024-13-01");
    expect(result.success).toBe(false);
  });

  it("validates date-time format", () => {
    const s = string().format("date-time");
    expect(s.parse("2024-01-15T10:30:00Z")).toBe("2024-01-15T10:30:00Z");
    const result = s.safeParse("not-a-date-time");
    expect(result.success).toBe(false);
  });

  it("chains constraints immutably", () => {
    const base = string();
    const constrained = base.minLength(1).maxLength(10);
    // base should not be affected
    expect(base.safeParse("").success).toBe(true);
    expect(constrained.safeParse("").success).toBe(false);
  });

  it("applies coercion: trim", () => {
    const s = string().coerce({ trim: true });
    expect(s.parse("  hello  ")).toBe("hello");
  });

  it("applies coercion: lower", () => {
    const s = string().coerce({ lower: true });
    expect(s.parse("HELLO")).toBe("hello");
  });

  it("applies coercion: upper", () => {
    const s = string().coerce({ upper: true });
    expect(s.parse("hello")).toBe("HELLO");
  });
});
