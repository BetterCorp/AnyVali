import { describe, it, expect } from "vitest";
import {
  any,
  unknown,
  never,
  null_,
  literal,
  enum_,
  string,
  int,
  bool,
  array,
  tuple,
  object,
  record,
  union,
  intersection,
  optional,
  nullable,
  exportSchema,
  parse,
  safeParse,
} from "../../src/index.js";
import { RefSchema } from "../../src/schemas/ref.js";
import { NullableSchema } from "../../src/schemas/nullable.js";
import { OptionalSchema } from "../../src/schemas/optional.js";
import { StringSchema } from "../../src/schemas/string.js";
import { ABSENT } from "../../src/schemas/base.js";

describe("_toNode / export for all schema types", () => {
  it("any", () => {
    const doc = exportSchema(any());
    expect(doc.root.kind).toBe("any");
  });

  it("any with default", () => {
    const doc = exportSchema(any().default(42));
    expect(doc.root.kind).toBe("any");
    expect(doc.root.default).toBe(42);
  });

  it("unknown", () => {
    const doc = exportSchema(unknown());
    expect(doc.root.kind).toBe("unknown");
  });

  it("unknown with default", () => {
    const doc = exportSchema(unknown().default("x"));
    expect(doc.root.kind).toBe("unknown");
    expect(doc.root.default).toBe("x");
  });

  it("never", () => {
    const doc = exportSchema(never());
    expect(doc.root.kind).toBe("never");
  });

  it("null", () => {
    const doc = exportSchema(null_());
    expect(doc.root.kind).toBe("null");
  });

  it("null with default", () => {
    const doc = exportSchema(null_().default(null));
    expect(doc.root.kind).toBe("null");
    expect(doc.root.default).toBe(null);
  });

  it("literal", () => {
    const doc = exportSchema(literal("hello"));
    expect(doc.root.kind).toBe("literal");
    expect((doc.root as any).value).toBe("hello");
  });

  it("literal with default", () => {
    const doc = exportSchema(literal(42).default(42));
    expect(doc.root.default).toBe(42);
  });

  it("enum", () => {
    const doc = exportSchema(enum_(["a", "b", 1]));
    expect(doc.root.kind).toBe("enum");
    expect((doc.root as any).values).toEqual(["a", "b", 1]);
  });

  it("enum with default", () => {
    const doc = exportSchema(enum_(["x", "y"]).default("x"));
    expect(doc.root.default).toBe("x");
  });

  it("tuple", () => {
    const doc = exportSchema(tuple([string(), int()]));
    expect(doc.root.kind).toBe("tuple");
    expect((doc.root as any).elements).toHaveLength(2);
    expect((doc.root as any).elements[0].kind).toBe("string");
    expect((doc.root as any).elements[1].kind).toBe("int");
  });

  it("tuple with default", () => {
    const doc = exportSchema(tuple([string()]).default(["hi"] as any));
    expect(doc.root.default).toEqual(["hi"]);
  });

  it("record", () => {
    const doc = exportSchema(record(int()));
    expect(doc.root.kind).toBe("record");
    expect((doc.root as any).valueSchema.kind).toBe("int");
  });

  it("record with default", () => {
    const doc = exportSchema(record(int()).default({} as any));
    expect(doc.root.default).toEqual({});
  });

  it("union", () => {
    const doc = exportSchema(union([string(), int()]));
    expect(doc.root.kind).toBe("union");
    expect((doc.root as any).variants).toHaveLength(2);
  });

  it("union with default", () => {
    const doc = exportSchema(union([string(), int()]).default("x" as any));
    expect(doc.root.default).toBe("x");
  });

  it("intersection", () => {
    const doc = exportSchema(
      intersection([object({ a: string() }), object({ b: int() })])
    );
    expect(doc.root.kind).toBe("intersection");
    expect((doc.root as any).allOf).toHaveLength(2);
  });

  it("intersection with default", () => {
    const doc = exportSchema(
      intersection([object({ a: string() })]).default({ a: "x" } as any)
    );
    expect(doc.root.default).toEqual({ a: "x" });
  });

  it("optional", () => {
    const doc = exportSchema(optional(string()));
    expect(doc.root.kind).toBe("optional");
    expect((doc.root as any).inner.kind).toBe("string");
  });

  it("optional with default", () => {
    const doc = exportSchema(optional(string().default("hi")));
    expect(doc.root.kind).toBe("optional");
  });

  it("nullable", () => {
    const doc = exportSchema(nullable(string()));
    expect(doc.root.kind).toBe("nullable");
    expect((doc.root as any).inner.kind).toBe("string");
  });

  it("nullable with default", () => {
    const doc = exportSchema(nullable(string()).default(null as any));
    expect(doc.root.default).toBe(null);
  });

  it("ref", () => {
    const ref = new RefSchema("#/definitions/User");
    const node = ref._toNode();
    expect(node.kind).toBe("ref");
    expect((node as any).ref).toBe("#/definitions/User");
  });

  it("ref _validate without resolver adds issue", () => {
    const ref = new RefSchema("#/definitions/Missing");
    const result = ref.safeParse("anything");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("unsupported_schema_kind");
    }
  });

  it("ref _validate with resolver delegates", () => {
    const ref = new RefSchema("#/definitions/Name", () => string());
    expect(ref.parse("hello")).toBe("hello");
  });
});

describe("Top-level parse/safeParse functions", () => {
  it("parse delegates to schema.parse", () => {
    expect(parse(string(), "hello")).toBe("hello");
  });

  it("parse throws on failure", () => {
    expect(() => parse(string(), 42)).toThrow();
  });

  it("safeParse returns result", () => {
    const r = safeParse(string(), "hi");
    expect(r.success).toBe(true);
    if (r.success) expect(r.data).toBe("hi");
  });

  it("safeParse returns failure", () => {
    const r = safeParse(string(), 42);
    expect(r.success).toBe(false);
  });
});

describe("base.ts edge cases", () => {
  it("_getCoercionTarget returns 'unknown' by default", () => {
    const s = any();
    expect((s as any)._getCoercionTarget()).toBe("unknown");
  });

  it("export throws in portable mode when not portable", () => {
    const s = any();
    (s as any)._isPortable = false;
    expect(() => s.export("portable")).toThrow(
      "Cannot export in portable mode"
    );
  });

  it("export works in extended mode even if not portable", () => {
    const s = any();
    (s as any)._isPortable = false;
    const doc = s.export("extended");
    expect(doc.root.kind).toBe("any");
  });
});

describe("nullable._validate direct calls", () => {
  it("returns null when input is null", () => {
    const s = new NullableSchema(new StringSchema());
    const ctx = { path: [], issues: [] };
    expect(s._validate(null, ctx)).toBe(null);
    expect(ctx.issues).toHaveLength(0);
  });

  it("delegates to inner when input is not null", () => {
    const s = new NullableSchema(new StringSchema());
    const ctx = { path: [], issues: [] };
    expect(s._validate("hello", ctx)).toBe("hello");
  });

  it("inner validation failure propagates", () => {
    const s = new NullableSchema(new StringSchema());
    const ctx = { path: [], issues: [] };
    s._validate(42, ctx);
    expect(ctx.issues.length).toBeGreaterThan(0);
  });
});

describe("optional._validate direct calls", () => {
  it("returns undefined when input is undefined", () => {
    const s = new OptionalSchema(new StringSchema());
    const ctx = { path: [], issues: [] };
    expect(s._validate(undefined, ctx)).toBe(undefined);
    expect(ctx.issues).toHaveLength(0);
  });

  it("returns undefined when input is ABSENT", () => {
    const s = new OptionalSchema(new StringSchema());
    const ctx = { path: [], issues: [] };
    expect(s._validate(ABSENT, ctx)).toBe(undefined);
    expect(ctx.issues).toHaveLength(0);
  });

  it("delegates to inner when input is present", () => {
    const s = new OptionalSchema(new StringSchema());
    const ctx = { path: [], issues: [] };
    expect(s._validate("hello", ctx)).toBe("hello");
  });
});
