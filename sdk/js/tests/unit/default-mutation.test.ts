import { describe, it, expect } from "vitest";
import { any, unknown, object } from "../../src/index.js";

// Repro for AVV-007: mutable default values shared between validations.
// Pass-through schemas (any/unknown) return the default value by reference,
// so mutating one parse result corrupts the default for the next parse.
describe("Mutable default isolation (AVV-007)", () => {
  it("any() array default is not shared between parses", () => {
    const s = any().default([] as unknown[]);
    const a = s.parse(undefined) as unknown[];
    a.push("mutated");
    const b = s.parse(undefined) as unknown[];
    expect(b).toEqual([]);
  });

  it("unknown() object default is not shared between parses", () => {
    const s = unknown().default({ tags: [] as string[] });
    const a = s.parse(undefined) as { tags: string[] };
    a.tags.push("x");
    const b = s.parse(undefined) as { tags: string[] };
    expect(b.tags).toEqual([]);
  });

  it("object field with any() default is not shared", () => {
    const s = object({ meta: any().default({ count: 0, items: [] as number[] }) });
    const a = s.parse({}) as { meta: { count: number; items: number[] } };
    a.meta.items.push(1);
    a.meta.count = 99;
    const b = s.parse({}) as { meta: { count: number; items: number[] } };
    expect(b.meta).toEqual({ count: 0, items: [] });
  });
});
