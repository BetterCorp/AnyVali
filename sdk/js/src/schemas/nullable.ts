import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema, ABSENT, MAX_DEPTH } from "./base.js";

export class NullableSchema<
  T extends BaseSchema<any, any> = BaseSchema,
> extends BaseSchema<unknown, T["_output"] | null> {
  /** @internal */ _inner: T;

  constructor(inner: T) {
    super();
    this._inner = inner;
  }

  /** @internal Resolve presence without moving defaults across coercion steps. */
  _hasDefault(depth = 0): boolean {
    return this._defaultValue !== ABSENT ||
      (depth < MAX_DEPTH && this._inner._hasDefault(depth + 1));
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
    if (this._metadata?.sensitive === true && ctx.sensitiveMode) {
      return super._runPipeline(input, ctx);
    }
    if ((input === undefined || input === ABSENT) && this._defaultValue !== ABSENT) {
      return super._runPipeline(input, ctx);
    }
    return this._inner._runPipeline(input, ctx);
  }

  _children(): BaseSchema[] { return [this._inner]; }

  _toNode(): SchemaNode {
    const node = {
      kind: "nullable" as const,
      inner: this._inner._toNode(),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}
