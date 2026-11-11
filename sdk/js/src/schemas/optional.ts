import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema, ABSENT, MAX_DEPTH } from "./base.js";

export class OptionalSchema<
  T extends BaseSchema<any, any> = BaseSchema,
> extends BaseSchema<unknown, T["_output"] | undefined> {
  /** @internal */ _inner: T;
  /** @internal */ _isOptionalWrapper = true;

  constructor(inner: T) {
    super();
    this._inner = inner;
  }

  /** @internal Resolve presence without moving defaults across coercion steps. */
  _hasDefault(depth = 0): boolean {
    return this._defaultValue !== ABSENT ||
      (depth < MAX_DEPTH && this._inner._hasDefault(depth + 1));
  }

  /** Validate a materialized wrapper default through the inner pipeline. */
  _validate(input: unknown, ctx: ParseContext): unknown {
    if (input === undefined || input === ABSENT) {
      return undefined;
    }
    return this._inner._runPipeline(input, ctx);
  }

  /** Preserve absence while applying defaults and coercions at their declaring nodes. */
  _runPipeline(input: unknown, ctx: ParseContext): unknown {
    const isAbsent = input === undefined || input === ABSENT;

    if ((!isAbsent && this._coercionConfig !== undefined) ||
        (!isAbsent && this._metadata?.sensitive === true && ctx.sensitiveMode)) {
      return super._runPipeline(input, ctx);
    }

    // If absent and we have a default on this wrapper, use the normal
    // BaseSchema default path so validation and cloning still run.
    if (isAbsent && this._defaultValue !== ABSENT) {
      return super._runPipeline(input, ctx);
    }

    if (isAbsent) {
      return this._inner._hasDefault()
        ? this._inner._runPipeline(input, ctx)
        : undefined;
    }

    // Delegate to inner's pipeline for coercion etc.
    return this._inner._runPipeline(input, ctx);
  }

  /** @internal Infer conversion targets through transparent wrappers. */
  _getCoercionTarget(depth = 0): string {
    return depth < MAX_DEPTH ? this._inner._getCoercionTarget(depth + 1) : "unknown";
  }

  _children(): BaseSchema[] { return [this._inner]; }

  _toNode(): SchemaNode {
    const node = {
      kind: "optional" as const,
      inner: this._inner._toNode(),
    };
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}
