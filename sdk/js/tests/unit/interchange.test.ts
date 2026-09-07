import { describe, it, expect } from "vitest";
import type { AnyValiDocument } from "../../src/types.js";
import {
  string,
  int,
  object,
  array,
  optional,
  nullable,
  bool,
  exportSchema,
  importSchema,
  record,
  tuple,
  union,
  intersection,
} from "../../src/index.js";

describe("Document extensions", () => {
  const document = (extensions: AnyValiDocument["extensions"]): AnyValiDocument => ({
    anyvaliVersion: "1.0", schemaVersion: "1.1", root: { kind: "string" },
    definitions: {}, extensions,
  });

  it.each(["vendor", "js", "default"])("rejects unsupported semantic namespace %s", namespace => {
    const doc = document({
      default: { _criticality: "informational", hint: "not an implementation" },
      [namespace]: { _criticality: "semantic", feature: true },
    });
    expect(() => importSchema(doc)).toThrow("unsupported_extension");
  });

  it.each([null, [], "invalid", { vendor: null }, { vendor: [] }, { vendor: true },
    { vendor: { _criticality: "unknown" } }, { vendor: { _criticality: null } }])(
    "rejects malformed extensions %j", extensions => {
      expect(() => importSchema(document(extensions as any))).toThrow();
    },
  );

  it("retains informational namespaces without sharing mutable input or output", () => {
    const extensions = {
      vendor: { _criticality: "informational", hint: { values: ["keep"] } },
      default: { hint: "implicit informational" },
    };
    const original = structuredClone(extensions);
    const schema = importSchema(document(extensions)).default("fallback");
    extensions.vendor.hint.values.push("changed input");
    expect(schema.parse("ok")).toBe("ok");
    const exported = exportSchema(schema, "extended");
    expect(exported.extensions).toEqual(original);
    (exported.extensions.vendor.hint as { values: string[] }).values.push("changed output");
    expect(schema.export("extended").extensions).toEqual(original);
    expect(importSchema(schema.export("extended")).export("extended").extensions).toEqual(original);
    expect(exportSchema(schema, "portable").extensions).toEqual({});
  });

  it("collects composed namespaces and rejects conflicting payloads only in extended mode", () => {
    const child = importSchema(document({ vendor: { a: 1, b: 2 } }));
    for (const parent of [object({ value: child }), array(child), record(child), tuple([child]),
      optional(child), nullable(child), union([child, bool()]), intersection([child, string()])]) {
      expect(parent.export("extended").extensions).toEqual({ vendor: { a: 1, b: 2 } });
    }
    const equal = importSchema(document({ vendor: { b: 2, a: 1 }, other: { hint: true } }));
    expect(object({ child, equal }).export("extended").extensions)
      .toEqual({ vendor: { a: 1, b: 2 }, other: { hint: true } });
    const conflicting = importSchema(document({ vendor: { a: 3 } }));
    const parent = object({ child, conflicting });
    expect(() => parent.export("extended")).toThrow("Conflicting extension namespace: vendor");
    expect(parent.export("portable").extensions).toEqual({});
    const special = JSON.parse('{"__proto__":{"hint":"keep"}}');
    expect(importSchema(document(special)).export("extended").extensions).toEqual(special);
  });
});

describe("Export", () => {
  it("merges identical definitions with reordered Unicode keys", () => {
    const child = (custom: Record<string, number>) => importSchema({
      anyvaliVersion: "1.0", schemaVersion: "1.1",
      root: { kind: "ref", ref: "#/definitions/Value" },
      definitions: { Value: { kind: "string", metadata: { custom } } }, extensions: {},
    } as any);
    const parent = object({
      first: child({ "\u00e9": 1, "e\u0301": 2 }),
      second: child({ "e\u0301": 2, "\u00e9": 1 }),
    });
    expect(importSchema(parent.export()).parse({ first: "a", second: "b" }))
      .toEqual({ first: "a", second: "b" });
  });

  it("collects child definitions through every composite and rejects conflicting names", () => {
    const doc = { anyvaliVersion: "1.0", schemaVersion: "1.1",
      root: { kind: "ref", ref: "#/definitions/Value" },
      definitions: { Value: { kind: "string", minLength: 1 } }, extensions: {} };
    const child = importSchema(doc as any);
    const cases = [
      [object({ value: child }), { value: "ok" }],
      [array(child), ["ok"]], [record(child), { key: "ok" }],
      [tuple([child]), ["ok"]], [optional(child), "ok"], [nullable(child), "ok"],
      [union([child, bool()]), "ok"], [intersection([child, string()]), "ok"],
    ] as const;
    for (const [parent, input] of cases) {
      expect(parent.parse(input)).toEqual(input);
      const exported = parent.export();
      expect(exported.definitions).toEqual(doc.definitions);
      expect(importSchema(exported).parse(input)).toEqual(input);
    }
    const reordered = importSchema({ ...doc, definitions: { Value: { minLength: 1, kind: "string" } } } as any);
    expect(() => object({ first: child, second: child, third: reordered }).export()).not.toThrow();
    const conflicting = importSchema({ ...doc, definitions: { Value: { kind: "bool" } } } as any);
    const parent = object({ first: child, second: conflicting });
    expect(parent.parse({ first: "ok", second: true })).toEqual({ first: "ok", second: true });
    expect(() => exportSchema(parent)).toThrow("Conflicting definition: Value");
  });

  it("exports a simple string schema", () => {
    const doc = exportSchema(string().minLength(1).maxLength(100));
    expect(doc.anyvaliVersion).toBe("1.0");
    expect(doc.schemaVersion).toBe("1.1");
    expect(doc.root.kind).toBe("string");
    expect((doc.root as any).minLength).toBe(1);
    expect((doc.root as any).maxLength).toBe(100);
  });

  it("exports an object schema", () => {
    const s = object({
      name: string(),
      age: optional(int()),
    });
    const doc = exportSchema(s);
    expect(doc.root.kind).toBe("object");
    const root = doc.root as any;
    expect(root.properties.name.kind).toBe("string");
    expect(root.required).toContain("name");
    expect(root.required).not.toContain("age");
  });

  it("exports defaults", () => {
    const s = string().default("hello");
    const doc = exportSchema(s);
    expect(doc.root.default).toBe("hello");
  });

  it("exports coercion config", () => {
    const s = int().coerce({ from: "string" });
    const doc = exportSchema(s);
    expect(doc.root.coerce).toEqual({ from: "string" });
  });
});

describe("Import", () => {
  it("preserves reserved and custom metadata on every imported node", () => {
    const metadata = { description: "Private", sensitive: true, custom: { owner: "ports" } };
    for (const root of [
      { kind: "string", metadata },
      { kind: "optional", inner: { kind: "string" }, metadata },
      { kind: "nullable", inner: { kind: "string" }, metadata },
      { kind: "ref", ref: "#/definitions/Secret", metadata },
    ]) {
      const doc = { anyvaliVersion: "1.0", schemaVersion: "1.1", root, definitions: {
        Secret: { kind: "string" },
      }, extensions: {} };
      const schema = importSchema(doc as any);
      expect(schema.export().root.metadata).toEqual(metadata);
      expect(exportSchema(importSchema(schema.export())).root.metadata).toEqual(metadata);
    }
    expect(() => importSchema({ ...exportSchema(string()), root: { kind: "string", metadata: null } } as any))
      .toThrow("Schema metadata must be an object");
  });

  it("round-trips a string schema", () => {
    const original = string().minLength(1).maxLength(100);
    const doc = exportSchema(original);
    const imported = importSchema(doc);
    expect(imported.parse("hello")).toBe("hello");
    expect(imported.safeParse("").success).toBe(false);
  });

  it("round-trips an object schema", () => {
    const original = object({
      name: string(),
      active: bool().default(true),
    });
    const doc = exportSchema(original);
    const imported = importSchema(doc);
    const result = imported.parse({ name: "Alice" });
    expect(result).toEqual({ name: "Alice", active: true });
  });

  it("round-trips an array schema", () => {
    const original = array(int()).minItems(1);
    const doc = exportSchema(original);
    const imported = importSchema(doc);
    expect(imported.parse([1, 2, 3])).toEqual([1, 2, 3]);
    expect(imported.safeParse([]).success).toBe(false);
  });

  it("round-trips nullable", () => {
    const original = nullable(string());
    const doc = exportSchema(original);
    const imported = importSchema(doc);
    expect(imported.parse(null)).toBe(null);
    expect(imported.parse("hello")).toBe("hello");
  });

  it("round-trips coercion", () => {
    const original = int().coerce({ from: "string" });
    const doc = exportSchema(original);
    const imported = importSchema(doc);
    expect(imported.parse("42")).toBe(42);
  });

  it("throws on missing kind field", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1",
      root: {},
      definitions: {},
      extensions: {},
    };
    expect(() => importSchema(doc as any)).toThrow();
  });

  it("throws on null/empty root", () => {
    expect(() =>
      importSchema({ anyvaliVersion: "1.0", schemaVersion: "1", root: null } as any)
    ).toThrow();
    expect(() =>
      importSchema({ anyvaliVersion: "1.0", schemaVersion: "1" } as any)
    ).toThrow();
  });

  it("imports invalid regex patterns without throwing", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: { kind: "string", pattern: "(" },
      definitions: {},
      extensions: {},
    };

    const imported = importSchema(doc as any);

    expect(() => imported.safeParse("abc")).not.toThrow();
    expect(imported.safeParse("abc").success).toBe(false);
  });

  it("imports array schemas with canonical items keys", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: {
        kind: "array",
        items: { kind: "int" },
      },
      definitions: {},
      extensions: {},
    };

    const imported = importSchema(doc as any);
    expect(imported.parse([1, 2, 3])).toEqual([1, 2, 3]);
    expect(imported.safeParse(["a"]).success).toBe(false);
  });

  it("imports union schemas with canonical variants keys", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1.1",
      root: {
        kind: "union",
        variants: [{ kind: "string" }, { kind: "int" }],
      },
      definitions: {},
      extensions: {},
    };

    const imported = importSchema(doc as any);
    expect(imported.parse("hello")).toBe("hello");
    expect(imported.parse(42)).toBe(42);
    expect(imported.safeParse(true).success).toBe(false);
  });

  it("imports __proto__ property names without prototype pollution", () => {
    const doc = JSON.parse(`{
      "anyvaliVersion": "1.0",
      "schemaVersion": "1.1",
      "root": {
        "kind": "object",
        "properties": {
          "__proto__": { "kind": "string" }
        },
        "required": ["__proto__"],
        "unknownKeys": "reject"
      },
      "definitions": {},
      "extensions": {}
    }`);
    const input = JSON.parse('{"__proto__":"safe"}') as Record<string, unknown>;

    const imported = importSchema(doc as any);

    expect(() => imported.parse(input)).not.toThrow();
    const result = imported.parse(input) as Record<string, unknown>;
    expect(Object.getPrototypeOf(result)).toBe(Object.prototype);
    expect(Object.prototype.hasOwnProperty.call(result, "__proto__")).toBe(true);
    expect(Object.getOwnPropertyDescriptor(result, "__proto__")?.value).toBe(
      "safe"
    );
    expect(({} as Record<string, unknown>).safe).toBeUndefined();
  });
});
