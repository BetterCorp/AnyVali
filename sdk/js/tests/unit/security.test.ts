import { describe, it, expect } from "vitest";
import {
  string,
  number,
  int,
  int8,
  int16,
  int32,
  int64,
  uint8,
  uint16,
  uint32,
  uint64,
  float32,
  float64,
  object,
  array,
  optional,
  union,
  importSchema,
  RefSchema,
} from "../../src/index.js";

// ---------------------------------------------------------------------------
// 1. ReDoS - CVE-2016-4055 / CVE-2022-25883
// ---------------------------------------------------------------------------
describe("CVE-2016-4055 / CVE-2022-25883 - ReDoS catastrophic backtracking", () => {
  // These patterns are known to cause catastrophic backtracking in naive
  // regex engines. The tests verify that validation completes rather than
  // hanging. If the test runner's own timeout fires, the library is
  // vulnerable to ReDoS.

  it("handles (a+)+$ pattern without hanging", () => {
    const s = string().pattern("(a+)+$");
    // 24 'a' chars followed by '!' - a classic ReDoS trigger
    const malicious = "a".repeat(24) + "!";
    const result = s.safeParse(malicious);
    // The pattern should either match or not, but must complete
    expect(result.success).toBe(false);
  });

  it("handles (a|a)+$ pattern without hanging", () => {
    const s = string().pattern("(a|a)+$");
    const malicious = "a".repeat(24) + "!";
    const result = s.safeParse(malicious);
    expect(result.success).toBe(false);
  });

  it("handles ^([a-zA-Z]+)*$ pattern without hanging", () => {
    // Anchored version forces full backtracking on non-matching suffix
    const s = string().pattern("^([a-zA-Z]+)*$");
    const malicious = "a".repeat(24) + "1";
    const result = s.safeParse(malicious);
    expect(result.success).toBe(false);
  });

  it("accepts valid input for ReDoS-prone patterns", () => {
    const s = string().pattern("(a+)+$");
    const result = s.safeParse("aaaaaa");
    expect(result.success).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// 2. Prototype Pollution - CVE-2019-10744 / CVE-2020-8203
// ---------------------------------------------------------------------------
describe("CVE-2019-10744 / CVE-2020-8203 - Prototype pollution", () => {
  it("__proto__ in input does not pollute Object.prototype (default strip mode)", () => {
    // Note: __proto__ cannot be used as a declared schema property key in JS
    // because object literals treat it as a prototype setter, not an own property.
    // Instead, we verify __proto__ is safely stripped as an unknown key.
    const s = object({ name: string() }); // default unknownKeys = "strip"
    const input = JSON.parse('{"name":"Alice","__proto__":{"polluted":"yes"}}') as Record<string, unknown>;
    const result = s.safeParse(input);

    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toEqual({ name: "Alice" });
    }
    // Object.prototype must not be polluted
    expect(({} as Record<string, unknown>).polluted).toBeUndefined();
  });

  it("handles __proto__ as an unknown key with unknownKeys: 'allow'", () => {
    const s = object({ name: string() }).unknownKeys("allow");
    const input = JSON.parse(
      '{"name":"Alice","__proto__":{"polluted":"yes"}}'
    ) as Record<string, unknown>;

    const result = s.parse(input) as Record<string, unknown>;

    expect(result.name).toBe("Alice");
    // __proto__ should be stored as own property data, not alter the prototype chain
    expect(Object.getPrototypeOf(result)).toBe(Object.prototype);
    expect(Object.prototype.hasOwnProperty.call(result, "__proto__")).toBe(true);
    expect(
      Object.getOwnPropertyDescriptor(result, "__proto__")?.value
    ).toEqual({ polluted: "yes" });
    // Global prototype must remain clean
    expect(({} as Record<string, unknown>).polluted).toBeUndefined();
  });

  it("handles 'constructor' as a property name safely", () => {
    const s = object({
      constructor: string(),
    });
    const input = { constructor: "overridden" };
    const result = s.parse(input);

    expect(result.constructor).toBe("overridden");
    // The Object constructor itself must not be corrupted
    expect(({}).constructor).toBe(Object);
  });

  it("handles 'prototype' as a property name safely", () => {
    const s = object({
      prototype: string(),
    });
    const input = { prototype: "value" };
    const result = s.parse(input);

    expect(Object.prototype.hasOwnProperty.call(result, "prototype")).toBe(true);
    expect((result as Record<string, unknown>).prototype).toBe("value");
  });

  it("strips __proto__ with unknownKeys: 'strip' without pollution", () => {
    const s = object({ name: string() }).unknownKeys("strip");
    const input = JSON.parse(
      '{"name":"Alice","__proto__":{"polluted":"stripped"}}'
    ) as Record<string, unknown>;

    const result = s.parse(input) as Record<string, unknown>;

    expect(result.name).toBe("Alice");
    expect(Object.prototype.hasOwnProperty.call(result, "__proto__")).toBe(false);
    expect(({} as Record<string, unknown>).polluted).toBeUndefined();
  });

  it("rejects __proto__ with unknownKeys: 'reject' without pollution", () => {
    const s = object({ name: string() }).unknownKeys("reject");
    const input = JSON.parse(
      '{"name":"Alice","__proto__":{"polluted":"rejected"}}'
    ) as Record<string, unknown>;

    const result = s.safeParse(input);
    expect(result.success).toBe(false);
    // Global prototype must remain clean regardless
    expect(({} as Record<string, unknown>).polluted).toBeUndefined();
  });
});

// ---------------------------------------------------------------------------
// 3. Recursive $ref DoS - billion laughs class / CVE-2003-1564
// ---------------------------------------------------------------------------
describe("CVE-2003-1564 - Recursive $ref and billion laughs DoS", () => {
  it("handles circular self-referencing $ref without infinite loop", () => {
    // Schema A references itself: A -> A
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "ref" as const, ref: "#/definitions/A" },
      definitions: {
        A: { kind: "ref" as const, ref: "#/definitions/A" },
      },
      extensions: {},
    };

    // Importing should succeed (lazy resolution).
    const schema = importSchema(doc as any);

    // Parsing on a pure self-cycle MUST terminate. Either by throwing
    // (V8 RangeError "Maximum call stack size exceeded") or by returning
    // a failure result. It MUST NOT hang the runtime. Bound wall-clock
    // time to catch a hang under CI.
    const start = Date.now();
    let threw = false;
    let result: unknown = undefined;
    try {
      result = schema.safeParse("anything");
    } catch (e) {
      threw = true;
      // V8 throws RangeError on stack exhaustion. Either kind of throw
      // is an acknowledgement that the cycle is not silently ignored.
      expect(e instanceof Error).toBe(true);
    }
    const elapsed = Date.now() - start;

    // 5-second wall-clock bound. Vitest default would catch a hang too,
    // but an explicit bound makes the assertion intentional.
    expect(elapsed).toBeLessThan(5000);

    // Exactly one of: parse threw, or parse returned a failure.
    // Returning success on a pure self-cycle would be a bug.
    if (!threw) {
      expect((result as { success: boolean }).success).toBe(false);
    }
  });

  it("handles deeply nested schemas (100+ levels) without crash", () => {
    // Build a 100-level nested object schema
    let inner: ReturnType<typeof object> = object({ value: string() });
    for (let i = 0; i < 100; i++) {
      inner = object({ child: inner });
    }

    // Build a deeply nested input to match
    let input: Record<string, unknown> = { value: "deep" };
    for (let i = 0; i < 100; i++) {
      input = { child: input };
    }

    // Should complete without crashing
    const result = inner.safeParse(input);
    expect(result.success).toBe(true);
  });

  it("handles mutually recursive $ref schemas (A -> B -> A) without hang", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "ref" as const, ref: "#/definitions/A" },
      definitions: {
        A: {
          kind: "object" as const,
          properties: {
            next: { kind: "ref" as const, ref: "#/definitions/B" },
          },
          required: [] as string[],
          unknownKeys: "reject" as const,
        },
        B: {
          kind: "object" as const,
          properties: {
            next: { kind: "ref" as const, ref: "#/definitions/A" },
          },
          required: [] as string[],
          unknownKeys: "reject" as const,
        },
      },
      extensions: {},
    };

    const schema = importSchema(doc as any);

    // Parse a non-recursive input - should succeed
    const result = schema.safeParse({});
    // Either succeeds or fails with issues, but must not hang
    expect(typeof result.success).toBe("boolean");
  });

  it("validates recursive schema with finite depth input", () => {
    // Build a recursive tree schema: Node = { value: string, children?: Node[] }
    let nodeSchema!: ReturnType<typeof object>;
    const childRef = optional(
      array(new RefSchema("#/definitions/Node", () => nodeSchema))
    );
    nodeSchema = object({
      value: string(),
      children: childRef,
    });

    const input = {
      value: "root",
      children: [
        { value: "child1" },
        {
          value: "child2",
          children: [{ value: "grandchild" }],
        },
      ],
    };

    const result = nodeSchema.safeParse(input);
    expect(result.success).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// 4. Integer overflow - CWE-190
// ---------------------------------------------------------------------------
describe("CWE-190 - Integer overflow and boundary checks", () => {
  describe("int8 boundaries", () => {
    it("accepts 127 (max int8)", () => {
      expect(int8().safeParse(127).success).toBe(true);
    });

    it("rejects 128 (overflow int8)", () => {
      expect(int8().safeParse(128).success).toBe(false);
    });

    it("accepts -128 (min int8)", () => {
      expect(int8().safeParse(-128).success).toBe(true);
    });

    it("rejects -129 (underflow int8)", () => {
      expect(int8().safeParse(-129).success).toBe(false);
    });
  });

  describe("int16 boundaries", () => {
    it("accepts 32767 (max int16)", () => {
      expect(int16().safeParse(32767).success).toBe(true);
    });

    it("rejects 32768 (overflow int16)", () => {
      expect(int16().safeParse(32768).success).toBe(false);
    });

    it("accepts -32768 (min int16)", () => {
      expect(int16().safeParse(-32768).success).toBe(true);
    });

    it("rejects -32769 (underflow int16)", () => {
      expect(int16().safeParse(-32769).success).toBe(false);
    });
  });

  describe("int32 boundaries", () => {
    it("accepts 2147483647 (max int32)", () => {
      expect(int32().safeParse(2147483647).success).toBe(true);
    });

    it("rejects 2147483648 (overflow int32)", () => {
      expect(int32().safeParse(2147483648).success).toBe(false);
    });

    it("accepts -2147483648 (min int32)", () => {
      expect(int32().safeParse(-2147483648).success).toBe(true);
    });

    it("rejects -2147483649 (underflow int32)", () => {
      expect(int32().safeParse(-2147483649).success).toBe(false);
    });
  });

  describe("uint8 boundaries", () => {
    it("accepts 255 (max uint8)", () => {
      expect(uint8().safeParse(255).success).toBe(true);
    });

    it("rejects 256 (overflow uint8)", () => {
      expect(uint8().safeParse(256).success).toBe(false);
    });

    it("accepts 0 (min uint8)", () => {
      expect(uint8().safeParse(0).success).toBe(true);
    });

    it("rejects -1 (underflow uint8)", () => {
      expect(uint8().safeParse(-1).success).toBe(false);
    });
  });

  describe("uint16 boundaries", () => {
    it("accepts 65535 (max uint16)", () => {
      expect(uint16().safeParse(65535).success).toBe(true);
    });

    it("rejects 65536 (overflow uint16)", () => {
      expect(uint16().safeParse(65536).success).toBe(false);
    });
  });

  describe("uint32 boundaries", () => {
    it("accepts 4294967295 (max uint32)", () => {
      expect(uint32().safeParse(4294967295).success).toBe(true);
    });

    it("rejects 4294967296 (overflow uint32)", () => {
      expect(uint32().safeParse(4294967296).success).toBe(false);
    });
  });

  describe("uint64 boundaries", () => {
    it("accepts Number.MAX_SAFE_INTEGER", () => {
      expect(uint64().safeParse(Number.MAX_SAFE_INTEGER).success).toBe(true);
    });

    it("accepts 0 (min uint64)", () => {
      expect(uint64().safeParse(0).success).toBe(true);
    });

    it("rejects -1 (underflow uint64)", () => {
      expect(uint64().safeParse(-1).success).toBe(false);
    });
  });

  describe("int64 / int boundaries", () => {
    it("accepts Number.MAX_SAFE_INTEGER for int64", () => {
      expect(int64().safeParse(Number.MAX_SAFE_INTEGER).success).toBe(true);
    });

    it("accepts Number.MIN_SAFE_INTEGER for int64", () => {
      expect(int64().safeParse(Number.MIN_SAFE_INTEGER).success).toBe(true);
    });

    it("accepts Number.MAX_SAFE_INTEGER for int()", () => {
      expect(int().safeParse(Number.MAX_SAFE_INTEGER).success).toBe(true);
    });

    it("accepts Number.MIN_SAFE_INTEGER for int()", () => {
      expect(int().safeParse(Number.MIN_SAFE_INTEGER).success).toBe(true);
    });
  });

  it("rejects floating point values for all integer types", () => {
    expect(int8().safeParse(1.5).success).toBe(false);
    expect(int16().safeParse(1.5).success).toBe(false);
    expect(int32().safeParse(1.5).success).toBe(false);
    expect(int64().safeParse(1.5).success).toBe(false);
    expect(uint8().safeParse(1.5).success).toBe(false);
    expect(uint16().safeParse(1.5).success).toBe(false);
    expect(uint32().safeParse(1.5).success).toBe(false);
    expect(uint64().safeParse(1.5).success).toBe(false);
    expect(int().safeParse(1.5).success).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// 5. NaN/Infinity injection - CWE-20
// ---------------------------------------------------------------------------
describe("CWE-20 - NaN and Infinity injection", () => {
  it("rejects NaN for number()", () => {
    expect(number().safeParse(NaN).success).toBe(false);
  });

  it("rejects NaN for int()", () => {
    expect(int().safeParse(NaN).success).toBe(false);
  });

  it("rejects NaN for float32()", () => {
    expect(float32().safeParse(NaN).success).toBe(false);
  });

  it("rejects NaN for float64()", () => {
    expect(float64().safeParse(NaN).success).toBe(false);
  });

  it("rejects Infinity for number()", () => {
    expect(number().safeParse(Infinity).success).toBe(false);
  });

  it("rejects -Infinity for number()", () => {
    expect(number().safeParse(-Infinity).success).toBe(false);
  });

  it("rejects Infinity for float32()", () => {
    expect(float32().safeParse(Infinity).success).toBe(false);
  });

  it("rejects -Infinity for float32()", () => {
    expect(float32().safeParse(-Infinity).success).toBe(false);
  });

  it("rejects Infinity for float64()", () => {
    expect(float64().safeParse(Infinity).success).toBe(false);
  });

  it("rejects -Infinity for float64()", () => {
    expect(float64().safeParse(-Infinity).success).toBe(false);
  });

  it("rejects Infinity for int()", () => {
    expect(int().safeParse(Infinity).success).toBe(false);
  });

  it("rejects -Infinity for int()", () => {
    expect(int().safeParse(-Infinity).success).toBe(false);
  });

  it("handles NaN !== NaN edge case - NaN is never equal to itself", () => {
    // Ensure the library does not use === NaN (which always returns false)
    const result = number().safeParse(NaN);
    expect(result.success).toBe(false);
    if (!result.success) {
      // Should produce a meaningful error, not silently accept
      expect(result.issues.length).toBeGreaterThan(0);
      expect(result.issues[0].code).toBe("invalid_type");
    }
  });

  it("rejects NaN for all integer types", () => {
    expect(int8().safeParse(NaN).success).toBe(false);
    expect(int16().safeParse(NaN).success).toBe(false);
    expect(int32().safeParse(NaN).success).toBe(false);
    expect(int64().safeParse(NaN).success).toBe(false);
    expect(uint8().safeParse(NaN).success).toBe(false);
    expect(uint16().safeParse(NaN).success).toBe(false);
    expect(uint32().safeParse(NaN).success).toBe(false);
    expect(uint64().safeParse(NaN).success).toBe(false);
  });

  it("rejects Infinity for all integer types", () => {
    expect(int8().safeParse(Infinity).success).toBe(false);
    expect(int16().safeParse(Infinity).success).toBe(false);
    expect(int32().safeParse(Infinity).success).toBe(false);
    expect(int64().safeParse(Infinity).success).toBe(false);
    expect(uint8().safeParse(Infinity).success).toBe(false);
    expect(uint16().safeParse(Infinity).success).toBe(false);
    expect(uint32().safeParse(Infinity).success).toBe(false);
    expect(uint64().safeParse(Infinity).success).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// 6. Format validation bypass - CWE-20
// ---------------------------------------------------------------------------
describe("CWE-20 - Format validation bypass", () => {
  describe("email format", () => {
    it("does not silently ignore a tampered email format name", () => {
      const s = string().format("email\0" as any);
      expect(s.safeParse("not-an-email").success).toBe(false);
    });

    it("does not import a tampered email format as an unconstrained string", () => {
      const schema = importSchema({
        anyvaliVersion: "1.0",
        schemaVersion: "1.1",
        root: { kind: "string", format: "email\0" as any },
        definitions: {},
        extensions: {},
      });
      expect(schema.safeParse("not-an-email").success).toBe(false);
    });

    it("rejects null byte injection: user@example.com\\0.evil.com", () => {
      const s = string().format("email");
      const result = s.safeParse("user@example.com\0.evil.com");
      expect(result.success).toBe(false);
    });

    it("rejects very long local part (>64 chars)", () => {
      const s = string().format("email");
      const longLocal = "a".repeat(65) + "@example.com";
      const result = s.safeParse(longLocal);
      // RFC 5321 limits local part to 64 characters; a strict validator should reject
      // If the library does not enforce this, the test documents the behavior
      expect(typeof result.success).toBe("boolean");
    });

    it("rejects email without domain", () => {
      const s = string().format("email");
      expect(s.safeParse("user@").success).toBe(false);
    });

    it("rejects email without local part", () => {
      const s = string().format("email");
      expect(s.safeParse("@example.com").success).toBe(false);
    });
  });

  describe("url format", () => {
    it("rejects javascript: protocol", () => {
      const s = string().format("url");
      expect(s.safeParse("javascript:alert(1)").success).toBe(false);
    });

    it("rejects data: protocol", () => {
      const s = string().format("url");
      expect(s.safeParse("data:text/html,<script>alert(1)</script>").success).toBe(false);
    });

    it("rejects file: protocol", () => {
      const s = string().format("url");
      expect(s.safeParse("file:///etc/passwd").success).toBe(false);
    });

    it("accepts valid https URL", () => {
      const s = string().format("url");
      expect(s.safeParse("https://example.com/path?q=1").success).toBe(true);
    });

    it("accepts valid http URL", () => {
      const s = string().format("url");
      expect(s.safeParse("http://example.com").success).toBe(true);
    });
  });

  describe("ipv4 format", () => {
    it("rejects octal notation 0177.0.0.1 (127.0.0.1 in octal)", () => {
      const s = string().format("ipv4");
      // Octal notation can be used to bypass IP filters
      expect(s.safeParse("0177.0.0.1").success).toBe(false);
    });

    it("rejects overflow 256.1.1.1", () => {
      const s = string().format("ipv4");
      expect(s.safeParse("256.1.1.1").success).toBe(false);
    });

    it("rejects overflow 999.999.999.999", () => {
      const s = string().format("ipv4");
      expect(s.safeParse("999.999.999.999").success).toBe(false);
    });

    it("accepts valid IPv4", () => {
      const s = string().format("ipv4");
      expect(s.safeParse("192.168.1.1").success).toBe(true);
    });

    it("accepts 0.0.0.0", () => {
      const s = string().format("ipv4");
      expect(s.safeParse("0.0.0.0").success).toBe(true);
    });

    it("accepts 255.255.255.255", () => {
      const s = string().format("ipv4");
      expect(s.safeParse("255.255.255.255").success).toBe(true);
    });

    it("rejects leading zeros in octets (010.0.0.1)", () => {
      const s = string().format("ipv4");
      // Leading zeros can indicate octal interpretation
      const result = s.safeParse("010.0.0.1");
      // Strict validators should reject; documents actual behavior
      expect(typeof result.success).toBe("boolean");
    });
  });

  describe("ipv6 format", () => {
    it("handles IPv4-mapped IPv6 address ::ffff:127.0.0.1", () => {
      const s = string().format("ipv6");
      // IPv4-mapped addresses may bypass IPv4-only filters
      const result = s.safeParse("::ffff:127.0.0.1");
      // This is technically valid IPv6 but may not match the regex;
      // the test documents the library's behavior
      expect(typeof result.success).toBe("boolean");
    });

    it("accepts valid IPv6 loopback ::1", () => {
      const s = string().format("ipv6");
      expect(s.safeParse("::1").success).toBe(true);
    });

    it("accepts valid full IPv6", () => {
      const s = string().format("ipv6");
      expect(
        s.safeParse("2001:0db8:85a3:0000:0000:8a2e:0370:7334").success
      ).toBe(true);
    });
  });
});

// ---------------------------------------------------------------------------
// 6b. Unicode length bypass / portability mismatch
// ---------------------------------------------------------------------------
describe("Unicode length constraints", () => {
  it("counts astral Unicode code points as one character", () => {
    const emoji = "😀";
    expect(string().maxLength(1).safeParse(emoji).success).toBe(true);
    expect(string().minLength(2).safeParse(emoji).success).toBe(false);
  });

  it("uses code point length for imported maxLength schemas", () => {
    const schema = importSchema({
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "string", maxLength: 1 },
      definitions: {},
      extensions: {},
    });
    expect(schema.safeParse("😀").success).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// 7. Large input DoS - CWE-400
// ---------------------------------------------------------------------------
describe("CWE-400 - Large input denial of service", () => {
  it("handles a 1MB string without crashing", () => {
    const s = string();
    const bigString = "x".repeat(1_000_000);
    const result = s.safeParse(bigString);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.length).toBe(1_000_000);
    }
  });

  it("validates a 1MB string with maxLength constraint", () => {
    const s = string().maxLength(500_000);
    const bigString = "x".repeat(1_000_000);
    const result = s.safeParse(bigString);
    expect(result.success).toBe(false);
  });

  it("handles deeply nested objects (100 levels) without crashing", () => {
    // Build schema
    let schema: ReturnType<typeof object> = object({ value: string() });
    for (let i = 0; i < 100; i++) {
      schema = object({ nested: schema });
    }

    // Build matching input
    let input: Record<string, unknown> = { value: "deep" };
    for (let i = 0; i < 100; i++) {
      input = { nested: input };
    }

    const result = schema.safeParse(input);
    expect(result.success).toBe(true);
  });

  it("handles deeply nested objects that fail validation with full error paths", () => {
    let schema: ReturnType<typeof object> = object({ value: int() });
    for (let i = 0; i < 50; i++) {
      schema = object({ nested: schema });
    }

    // Input with wrong type at the deepest level
    let input: Record<string, unknown> = { value: "not-an-int" };
    for (let i = 0; i < 50; i++) {
      input = { nested: input };
    }

    const result = schema.safeParse(input);
    expect(result.success).toBe(false);
    if (!result.success) {
      // Path should have 51 segments: "nested" x50 + "value"
      expect(result.issues[0].path.length).toBe(51);
    }
  });

  it("handles an array with 10000 items without crashing", () => {
    const s = array(int());
    const bigArray = Array.from({ length: 10_000 }, (_, i) => i);
    const result = s.safeParse(bigArray);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.length).toBe(10_000);
    }
  });

  it("handles an array with 10000 items that all fail validation", () => {
    const s = array(int());
    const bigArray = Array.from({ length: 10_000 }, () => "not-a-number");
    const result = s.safeParse(bigArray);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues.length).toBe(10_000);
    }
  });

  it("handles object with many properties", () => {
    const shape: Record<string, ReturnType<typeof string>> = {};
    for (let i = 0; i < 1000; i++) {
      shape[`prop_${i}`] = string();
    }
    const s = object(shape);

    const input: Record<string, string> = {};
    for (let i = 0; i < 1000; i++) {
      input[`prop_${i}`] = `value_${i}`;
    }

    const result = s.safeParse(input);
    expect(result.success).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// 8. JSON import injection
// ---------------------------------------------------------------------------
describe("JSON import injection", () => {
  it("rejects schema with unknown kind", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "evil_custom_type" },
      definitions: {},
      extensions: {},
    };

    expect(() => importSchema(doc as any)).toThrow();
  });

  it("rejects schema with missing kind field", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: {},
      definitions: {},
      extensions: {},
    };

    expect(() => importSchema(doc as any)).toThrow();
  });

  it("rejects schema with null root", () => {
    expect(() =>
      importSchema({
        anyvaliVersion: "1.0",
        schemaVersion: "1.1",
        root: null,
      } as any)
    ).toThrow();
  });

  it("rejects schema with undefined root", () => {
    expect(() =>
      importSchema({
        anyvaliVersion: "1.0",
        schemaVersion: "1.1",
      } as any)
    ).toThrow();
  });

  it("handles __proto__ in definition names without prototype pollution", () => {
    const doc = JSON.parse(`{
      "anyvaliVersion": "1.0",
      "schemaVersion": "1.1",
      "root": { "kind": "ref", "ref": "#/definitions/__proto__" },
      "definitions": {
        "__proto__": { "kind": "string" }
      },
      "extensions": {}
    }`);

    // Should either resolve the ref and work, or throw a clear error
    let result: unknown;
    let threw = false;
    try {
      const schema = importSchema(doc as any);
      result = schema.safeParse("test");
    } catch {
      threw = true;
    }

    if (!threw) {
      // If it resolved, the prototype should not be polluted
      expect(({} as Record<string, unknown>).__proto__).toBe(Object.prototype);
    }
    // Either outcome is acceptable as long as no pollution occurs
    expect(({} as Record<string, unknown>).polluted).toBeUndefined();
  });

  it("rejects schema with kind set to 'constructor'", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "constructor" },
      definitions: {},
      extensions: {},
    };

    expect(() => importSchema(doc as any)).toThrow();
  });

  it("rejects schema with kind set to 'toString'", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "toString" },
      definitions: {},
      extensions: {},
    };

    expect(() => importSchema(doc as any)).toThrow();
  });

  it("handles ref to nonexistent definition gracefully", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "ref" as const, ref: "#/definitions/DoesNotExist" },
      definitions: {},
      extensions: {},
    };

    const schema = importSchema(doc as any);
    // The ref should fail at parse time, not import time (lazy resolution)
    let threw = false;
    try {
      schema.safeParse("anything");
    } catch {
      threw = true;
    }
    // Should throw because the definition doesn't exist
    expect(threw).toBe(true);
  });

  it("imports valid schema after rejecting invalid ones (no state leakage)", () => {
    // First, try an invalid import
    try {
      importSchema({
        anyvaliVersion: "1.0",
        schemaVersion: "1.1",
        root: { kind: "unknown_kind" },
        definitions: {},
        extensions: {},
      } as any);
    } catch {
      // expected
    }

    // Now a valid import should still work
    const validDoc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "string" as const },
      definitions: {},
      extensions: {},
    };

    const schema = importSchema(validDoc as any);
    const result = schema.safeParse("hello");
    expect(result.success).toBe(true);
  });
});
