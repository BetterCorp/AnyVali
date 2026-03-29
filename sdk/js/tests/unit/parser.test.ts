import { describe, it, expect } from "vitest";
import { parse, safeParse } from "../../src/parse/parser.js";
import { applyDefault } from "../../src/parse/defaults.js";
import { ABSENT } from "../../src/schemas/base.js";
import {
  createDocument,
  ANYVALI_VERSION,
  SCHEMA_VERSION,
} from "../../src/interchange/document.js";
import { StringSchema } from "../../src/schemas/string.js";
import { IntSchema } from "../../src/schemas/int.js";
import { importSchema } from "../../src/interchange/importer.js";

describe("parse/parser.ts - standalone parse", () => {
  it("parse returns validated value", () => {
    const s = new StringSchema();
    expect(parse(s, "hello")).toBe("hello");
  });

  it("parse throws on invalid input", () => {
    const s = new StringSchema();
    expect(() => parse(s, 42)).toThrow();
  });

  it("safeParse returns success result", () => {
    const s = new StringSchema();
    const result = safeParse(s, "hi");
    expect(result.success).toBe(true);
    if (result.success) expect(result.data).toBe("hi");
  });

  it("safeParse returns failure result", () => {
    const s = new StringSchema();
    const result = safeParse(s, 123);
    expect(result.success).toBe(false);
    if (!result.success) expect(result.issues.length).toBeGreaterThan(0);
  });
});

describe("parse/defaults.ts - applyDefault", () => {
  it("returns default when input is undefined", () => {
    expect(applyDefault(undefined, "fallback")).toBe("fallback");
  });

  it("returns default when input is ABSENT", () => {
    expect(applyDefault(ABSENT, "fallback")).toBe("fallback");
  });

  it("returns input when it has a value", () => {
    expect(applyDefault("hello", "fallback")).toBe("hello");
  });

  it("returns input when default is ABSENT", () => {
    expect(applyDefault(undefined, ABSENT)).toBe(undefined);
  });

  it("returns input (even falsy) when present", () => {
    expect(applyDefault(0, "fallback")).toBe(0);
    expect(applyDefault("", "fallback")).toBe("");
    expect(applyDefault(null, "fallback")).toBe(null);
    expect(applyDefault(false, "fallback")).toBe(false);
  });
});

describe("interchange/document.ts", () => {
  it("exports constants", () => {
    expect(ANYVALI_VERSION).toBe("1.0");
    expect(SCHEMA_VERSION).toBe("1");
  });

  it("createDocument creates a valid document", () => {
    const root = { kind: "string" as const };
    const doc = createDocument(root as any);
    expect(doc.anyvaliVersion).toBe("1.0");
    expect(doc.schemaVersion).toBe("1");
    expect(doc.root).toBe(root);
    expect(doc.definitions).toEqual({});
    expect(doc.extensions).toEqual({});
  });

  it("createDocument with definitions and extensions", () => {
    const root = { kind: "ref" as const, ref: "#/definitions/Foo" };
    const defs = { Foo: { kind: "string" as const } };
    const exts = { myExt: { foo: "bar" } };
    const doc = createDocument(root as any, defs as any, exts);
    expect(doc.definitions).toBe(defs);
    expect(doc.extensions).toBe(exts);
  });
});

describe("importer.ts edge cases", () => {
  it("throws on unsupported schema kind", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1",
      root: { kind: "foobar" },
      definitions: {},
      extensions: {},
    };
    expect(() => importSchema(doc as any)).toThrow("Unsupported schema kind: foobar");
  });

  it("throws on unresolved ref definition", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1",
      root: { kind: "ref", ref: "#/definitions/Missing" },
      definitions: {},
      extensions: {},
    };
    const imported = importSchema(doc as any);
    // The ref is lazy, so it only throws when we try to parse
    expect(() => imported.parse("anything")).toThrow("Unresolved definition: Missing");
  });

  it("imports ref that resolves to a definition", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1",
      root: { kind: "ref", ref: "#/definitions/Name" },
      definitions: { Name: { kind: "string" } },
      extensions: {},
    };
    const imported = importSchema(doc as any);
    expect(imported.parse("hello")).toBe("hello");
  });

  it("imports with coerce as string format", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1",
      root: { kind: "int", coerce: "string->int" },
      definitions: {},
      extensions: {},
    };
    const imported = importSchema(doc as any);
    expect(imported.parse("42")).toBe(42);
  });

  it("imports with coerce as array format", () => {
    const doc = {
      anyvaliVersion: "1.0",
      schemaVersion: "1",
      root: { kind: "string", coerce: ["trim", "lower"] },
      definitions: {},
      extensions: {},
    };
    const imported = importSchema(doc as any);
    expect(imported.parse("  HELLO  ")).toBe("hello");
  });
});
