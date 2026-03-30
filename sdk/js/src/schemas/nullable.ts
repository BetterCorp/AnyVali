import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";

export class NullableSchema<
  T extends BaseSchema<any, any> = BaseSchema,
> extends BaseSchema<unknown, T["_output"] | null> {
  /** @internal */ _inner: T;

  constructor(inner: T) {
    super();
    this._inner = inner;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (input === null) {
      return null;
    }
    return this._inner._validate(input, ctx);
  }

  _runPipeline(input: unknown, ctx: ParseContext): unknown {
    if (input === null) {
      return null;
    }
    return this._inner._runPipeline(input, ctx);
  }

  _toNode(): SchemaNode {
    const node = {
      kind: "nullable" as const,
      inner: this._inner._toNode(),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}
