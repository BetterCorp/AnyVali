import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";

/** Map a tuple of schemas to a tuple of their output types. */
type InferTuple<T extends BaseSchema<any, any>[]> = {
  [K in keyof T]: T[K]["_output"];
};

export class TupleSchema<
  T extends BaseSchema<any, any>[] = BaseSchema[],
> extends BaseSchema<unknown[], InferTuple<T>> {
  private _items: T;

  constructor(items: [...T]) {
    super();
    this._items = items as T;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (!Array.isArray(input)) {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_TYPE,
        message: `Expected tuple, received ${describeType(input)}`,
        path: [...ctx.path],
        expected: "tuple",
        received: describeType(input),
      });
      return undefined;
    }

    if (input.length < this._items.length) {
      ctx.issues.push({
        code: ISSUE_CODES.TOO_SMALL,
        message: `Tuple must have exactly ${this._items.length} element(s), received ${input.length}`,
        path: [...ctx.path],
        expected: String(this._items.length),
        received: String(input.length),
      });
      return undefined;
    }

    if (input.length > this._items.length) {
      ctx.issues.push({
        code: ISSUE_CODES.TOO_LARGE,
        message: `Tuple must have exactly ${this._items.length} element(s), received ${input.length}`,
        path: [...ctx.path],
        expected: String(this._items.length),
        received: String(input.length),
      });
      return undefined;
    }

    const result: unknown[] = [];
    for (let i = 0; i < this._items.length; i++) {
      ctx.path.push(i);
      const val = this._items[i]._runPipeline(input[i], ctx);
      result.push(val);
      ctx.path.pop();
    }

    return result;
  }

  _toNode(): SchemaNode {
    const node = {
      kind: "tuple" as const,
      elements: this._items.map((s) => s._toNode()),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}
