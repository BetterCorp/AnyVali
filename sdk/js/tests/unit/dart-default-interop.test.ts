import { describe, expect, it } from "vitest";
import { importSchema, optional, nullable, string, int } from "../../src/index.js";
import type { AnyValiDocument } from "../../src/types.js";

/** Build a portable document exercising defaults at their declaring nodes. */
const document = (): AnyValiDocument => ({
  anyvaliVersion: "1.0", schemaVersion: "1.1",
  root: {
    kind: "object", required: ["text"], properties: {
      text: { kind: "nullable", schema: { kind: "string", coerce: ["trim"], default: "  fallback  " } },
      flag: { kind: "ref", ref: "#/definitions/Flag" },
      bag: { kind: "nullable", schema: { kind: "any", default: { value: "original" } } },
      nil: { kind: "nullable", schema: { kind: "null", default: null } },
      empty: { kind: "optional", schema: { kind: "nullable", schema: {
        kind: "array", items: { kind: "string" }, default: []
      } } },
    },
  },
  definitions: { Flag: { kind: "optional", schema: { kind: "bool", default: false } } },
  extensions: {},
} as AnyValiDocument);

describe("Dart portable transitive defaults", () => {
  it("materializes wrapped and referenced property defaults without re-coercing them", () => {
    const schema = importSchema(document());
    const expected = { text: "  fallback  ", flag: false, nil: null, empty: [], bag: { value: "original" } };
    const parsed = schema.parse({}) as typeof expected;
    expect(parsed).toEqual(expected);
    parsed.bag.value = "changed";
    expect(schema.parse({})).toEqual(expected);
    expect(schema.parse({ text: "  present  " })).toEqual({ ...expected, text: "present" });
    expect(importSchema(schema.export()).parse({})).toEqual(expected);
  });

  it("reports invalid nested defaults at the property path", () => {
    const doc = document();
    (doc.root as any).properties.text.schema.default = 42;
    const result = importSchema(doc).safeParse({});
    expect(result.success).toBe(false);
    if (!result.success) expect(result.issues).toContainEqual(expect.objectContaining({
      code: "default_invalid", path: ["text"],
    }));
  });

  it("bounds default lookup through cyclic references", () => {
    const doc = document();
    (doc.root as any).properties.cycle = { kind: "ref", ref: "#/definitions/Cycle" };
    doc.definitions.Cycle = { kind: "ref", ref: "#/definitions/Cycle" };
    expect(importSchema(doc).safeParse({}).success).toBe(true);
  });
});


describe("Dart reference and wrapper modifiers", () => {
  /** Define references sharing a target while retaining independent modifiers. */
  const modifiers = (): AnyValiDocument => ({
    anyvaliVersion: "1.0", schemaVersion: "1.1", extensions: {},
    root: { kind: "object", required: ["c", "d"], properties: {
      a: { kind: "ref", ref: "#/definitions/Value", default: 7 },
      b: { kind: "ref", ref: "#/definitions/Value", default: 8 },
      c: { kind: "ref", ref: "#/definitions/Value", coerce: ["string->int"] },
      d: { kind: "ref", ref: "#/definitions/Value" },
      chain: { kind: "ref", ref: "#/definitions/Upper", default: " x ", coerce: ["trim"] },
      wrapped: { kind: "optional", schema: { kind: "nullable", default: " x ", schema: {
        kind: "string", coerce: ["trim"], startsWith: "x"
      } } },
    } },
    definitions: { Value: { kind: "int" }, Upper: { kind: "string", coerce: ["upper"] } },
  } as AnyValiDocument);

  it("retains ref defaults/coercions independently before and after export", () => {
    const schema = importSchema(modifiers());
    const input = { c: "9", d: 10 };
    const expected = { a: 7, b: 8, c: 9, d: 10, chain: " X ", wrapped: "x" };
    expect(schema.parse(input)).toEqual(expected);
    expect(importSchema(schema.export()).parse(input)).toEqual(expected);
    expect(schema.parse({ ...input, chain: " y " })).toEqual({ ...expected, chain: "Y" });
    expect(schema.safeParse({ ...input, d: "10" }).success).toBe(false);
    expect(schema.safeParse({ ...input, a: "7" }).success).toBe(false);
  });

  it("keeps optional wrapping from moving a nullable default past coercion", () => {
    const schema = optional(nullable(string().coerce({ trim: true }).startsWith("x")).default(" x "));
    expect(schema.parse(undefined)).toBe("x");
    expect(importSchema(schema.export()).parse(undefined)).toBe("x");
    // A default declared on the string itself still skips that string's coercion.
    expect(optional(string().coerce({ trim: true }).default(" x ")).parse(undefined)).toBe(" x ");
  });

  it("infers coercion targets through references and transparent wrappers", () => {
    const doc = modifiers();
    doc.root = { kind: "ref", ref: "#/definitions/Value", coerce: {} } as any;
    expect(importSchema(doc).parse("7")).toBe(7);
    expect(optional(nullable(int())).coerce().parse("7")).toBe(7);
  });
});
