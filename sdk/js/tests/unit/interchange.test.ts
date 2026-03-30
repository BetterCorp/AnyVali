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
    expect(doc.schemaVersion).toBe("1");
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
});
