import { describe, expect, it } from "vitest";
import { importSchema } from "../../src/index.js";
import type { AnyValiDocument } from "../../src/types.js";

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
