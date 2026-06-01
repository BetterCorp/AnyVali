import { describe, it, expect } from "vitest";
import { object, int, array } from "../../src/index.js";

// A shared (but NON-circular) object reference appearing in sibling positions
// must validate. The circular-reference guard must only trip on true cycles
// (an ancestor reappearing), not on a DAG / repeated sibling.
describe("Shared non-circular references (false-positive guard)", () => {
  it("same object used in two object fields validates", () => {
    const inner = object({ a: int() });
    const s = object({ x: inner, y: inner });
    const shared = { a: 1 };
    const r = s.safeParse({ x: shared, y: shared });
    expect(r.success).toBe(true);
  });

  it("same object repeated in an array validates", () => {
    const s = array(object({ a: int() }));
    const shared = { a: 1 };
    const r = s.safeParse([shared, shared, shared]);
    expect(r.success).toBe(true);
  });

  it("a genuinely circular input is still rejected (not an infinite loop)", () => {
    const s = object({ self: object({ a: int() }) });
    const cyclic: any = {};
    cyclic.self = cyclic; // true cycle
    let r: ReturnType<typeof s.safeParse>;
    expect(() => {
      r = s.safeParse(cyclic);
    }).not.toThrow();
    expect(r!.success).toBe(false);
  });
});
