import { describe, it, expect } from "vitest";
import { importSchema } from "../../src/index.js";
import type { AnyValiDocument } from "../../src/types.js";

// Recursive schema: Node = object({ value: int, next: optional(ref Node) }).
function recursiveSchemaDoc(): AnyValiDocument {
  return {
    anyvaliVersion: "1.0",
    schemaVersion: "1",
    root: { kind: "ref", ref: "#/definitions/Node" } as any,
    definitions: {
      Node: {
        kind: "object",
        properties: {
          next: { kind: "ref", ref: "#/definitions/Node" },
        },
        required: [],
        unknownKeys: "strip",
      } as any,
    },
    extensions: {},
  } as AnyValiDocument;
}

// Build deeply nested data: { next: { next: { ... } } }.
function deepData(depth: number): unknown {
  let cur: any = {};
  for (let i = 0; i < depth; i++) cur = { next: cur };
  return cur;
}

// Build a deeply nested schema document: array(array(array(...))).
function deepArraySchemaDoc(depth: number): AnyValiDocument {
  let node: any = { kind: "int" };
  for (let i = 0; i < depth; i++) node = { kind: "array", items: node };
  return {
    anyvaliVersion: "1.0",
    schemaVersion: "1",
    root: node,
    definitions: {},
    extensions: {},
  } as AnyValiDocument;
}

describe("Recursion DoS (AVV-002 / AVV-003)", () => {
  it("safeParse must not throw on deeply nested recursive-schema input", () => {
    const schema = importSchema(recursiveSchemaDoc());
    const data = deepData(100_000);
    // safeParse contract: never throws, always returns a result.
    let result: ReturnType<typeof schema.safeParse>;
    expect(() => {
      result = schema.safeParse(data);
    }).not.toThrow();
    // Bounded: extremely deep input is rejected, not accepted blindly.
    expect(result!.success).toBe(false);
  });

  it("depth guard is not bypassed by a union in the recursion path", () => {
    // Node = object({ next: union(null, ref Node) }) -- the recursive ref is
    // reached *through* a union variant. The union must propagate depth so the
    // guard still bounds recursion.
    const doc: AnyValiDocument = {
      anyvaliVersion: "1.0",
      schemaVersion: "1",
      root: { kind: "ref", ref: "#/definitions/Node" } as any,
      definitions: {
        Node: {
          kind: "object",
          properties: {
            next: {
              kind: "union",
              variants: [{ kind: "null" }, { kind: "ref", ref: "#/definitions/Node" }],
            },
          },
          required: ["next"],
          unknownKeys: "strip",
        } as any,
      },
      extensions: {},
    } as AnyValiDocument;
    const schema = importSchema(doc);

    let cur: any = { next: null };
    for (let i = 0; i < 100_000; i++) cur = { next: cur };

    let result: ReturnType<typeof schema.safeParse>;
    expect(() => {
      result = schema.safeParse(cur);
    }).not.toThrow();
    expect(result!.success).toBe(false);
  });

  it("importSchema must not stack-overflow on a deeply nested schema doc", () => {
    const doc = deepArraySchemaDoc(100_000);
    // Should fail in a controlled way, not throw RangeError (stack overflow).
    let err: unknown;
    try {
      importSchema(doc);
    } catch (e) {
      err = e;
    }
    expect(err).toBeInstanceOf(Error);
    expect((err as Error) instanceof RangeError).toBe(false);
  });
});
