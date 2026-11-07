import fs from "node:fs";
import { describe, expect, it } from "vitest";
import { any, importSchema, object, optional, nullable, record, string, tuple } from "../../src/index.js";
import type { AnyValiDocument, CoercionConfig } from "../../src/types.js";

/** Build the native JS producer for the fixture consumed by Go and Dart tests. */
function producer() {
  const definitions: AnyValiDocument["definitions"] = {
    Text: { kind: "string", coerce: { trim: true } }, Value: { kind: "int" }, Flag: { kind: "bool" },
  };
  const ref = (name: string, coerce: CoercionConfig) => importSchema({
    anyvaliVersion: "1.0", schemaVersion: "1.1", root: { kind: "ref", ref: `#/definitions/${name}`, coerce },
    definitions, extensions: {},
  });
  return object({
    name: ref("Text", { trim: true, upper: true }), count: ref("Value", {}), flag: ref("Flag", { upper: true }),
    optional: optional(nullable(string().coerce({ trim: true }).startsWith("x")).default(" x ")),
    record: record(ref("Text", { trim: true })),
    tuple: tuple([nullable(ref("Value", {})), string().coerce({ lower: true })]),
    data: any().default({ kind: "ref", coerce: { trim: true }, schema: { coerce: { lower: true } } }),
  }).metadata({ wireData: { coerce: { lower: true } } });
}

const fixture = JSON.parse(fs.readFileSync(new URL("../../../go/testdata/js-ref-coercions.json", import.meta.url), "utf8"));

describe("canonical portable interoperability", () => {
  it("produces the same shared document that Go and Dart consume", () => {
    const schema = producer();
    expect(schema.export()).toEqual(fixture.schema);
    expect(schema.parse(fixture.input)).toEqual(fixture.output);
    expect(importSchema(schema.export()).parse(fixture.input)).toEqual(fixture.output);
    expect(schema.parse({ ...fixture.input, count: 42, flag: true })).toEqual({ ...fixture.output, flag: true });
  });
  it("keeps native flags and legacy child keys in extended output", () => {
    const document = producer().export("extended");
    const root = document.root as any;
    expect(root.properties.name.coerce).toEqual({ trim: true, upper: true });
    expect(root.properties.optional.inner).toBeDefined();
    expect(root.properties.record.valueSchema).toBeDefined();
    expect(document.definitions.Text.coerce).toEqual({ trim: true });
  });
  it("rejects unsupported portable coercions without changing extended exports", () => {
    const unsupported = any().coerce({ trim: true });
    expect(() => unsupported.export()).toThrow("String preprocessing");
    expect(unsupported.export("extended").root.coerce).toEqual({ trim: true });
    expect(() => string().coerce({ from: "number" }).export()).toThrow("Unsupported portable coercion source");
    expect(() => string().coerce({ unknown: true } as any).export()).toThrow("Unknown portable coercion flag");
    const invalid = importSchema({ anyvaliVersion: "1.0", schemaVersion: "1.1",
      root: { kind: "bool", coerce: ["string->int"] }, definitions: {}, extensions: {} });
    expect(() => invalid.export()).toThrow("does not match");
  });
  it("retains imported coercion operation order through export and reimport", () => {
    const operations = ["upper", "lower", "upper", "lower"];
    const schema = importSchema({ anyvaliVersion: "1.0", schemaVersion: "1.1",
      root: { kind: "string", coerce: operations }, definitions: {}, extensions: {} });
    expect(schema.parse("MiXeD")).toBe("mixed");
    expect(schema.export().root.coerce).toEqual(operations);
    expect(importSchema(schema.export()).parse("MiXeD")).toBe("mixed");
  });
});
