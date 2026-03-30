import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema, ABSENT } from "./base.js";

export class OptionalSchema<
  T extends BaseSchema<any, any> = BaseSchema,
> extends BaseSchema<unknown, T["_output"] | undefined> {
  /** @internal */ _inner: T;
  /** @internal */ _isOptionalWrapper = true;

  constructor(inner: T) {
    super();
    this._inner = inner;
    // Inherit defaults/coercion from inner
    this._defaultValue = inner._defaultValue as any;
    this._coercionConfig = inner._coercionConfig;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (input === undefined || input === ABSENT) {
      return undefined;
    }
    return this._inner._validate(input, ctx);
  }

  _runPipeline(input: unknown, ctx: ParseContext): unknown {
    const isAbsent = input === undefined || input === ABSENT;

    // If absent and we have a default from inner, apply it
    if (isAbsent && this._inner._defaultValue !== ABSENT) {
      return this._inner._runPipeline(input, ctx);
    }

    if (isAbsent) {
      return undefined;
    }

    // Delegate to inner's pipeline for coercion etc.
    return this._inner._runPipeline(input, ctx);
  }

  _toNode(): SchemaNode {
    const node = {
      kind: "optional" as const,
      inner: this._inner._toNode(),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}
