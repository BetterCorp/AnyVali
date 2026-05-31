import { describe, it, expect } from "vitest";
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
} from "../../src/index.js";

describe("Export", () => {
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
