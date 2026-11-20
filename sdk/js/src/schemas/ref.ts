import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema, ABSENT, MAX_DEPTH } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";

export class RefSchema extends BaseSchema<unknown, unknown> {
  private _ref: string;
  private _resolver?: () => BaseSchema;

  constructor(ref: string, resolver?: () => BaseSchema) {
    super();
    this._ref = ref;
    this._resolver = resolver;
  }

  /** @internal Bound cyclic reference lookup without evaluating defaults. */
  _hasDefault(depth = 0): boolean {
    return this._defaultValue !== ABSENT ||
      (depth < MAX_DEPTH && this._resolver !== undefined &&
        this._resolver()._hasDefault(depth + 1));
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (this._resolver) {
      const resolved = this._resolver();
      return resolved._runPipeline(input, ctx);
    }

    ctx.issues.push({
      code: ISSUE_CODES.UNSUPPORTED_SCHEMA_KIND,
      message: `Unresolved ref: ${this._ref}`,
      path: [...ctx.path],
    });
    return undefined;
  }

  /** Run occurrence-specific modifiers before validating the resolved target. */
  _runPipeline(input: unknown, ctx: ParseContext): unknown {
    if (this._coercionConfig !== undefined ||
        ((input === undefined || input === ABSENT) && this._defaultValue !== ABSENT) ||
        (this._metadata?.sensitive === true && ctx.sensitiveMode)) {
      return super._runPipeline(input, ctx);
    }
    return this._resolver
      ? this._resolver()._runPipeline(input, ctx)
      : this._validate(input, ctx);
  }

  /** @internal Resolve conversion kinds with the same bound as default lookup. */
  _getCoercionTarget(depth = 0): string {
    return depth < MAX_DEPTH && this._resolver
      ? this._resolver()._getCoercionTarget(depth + 1) : "unknown";
  }

  _toNode(): SchemaNode {
    return this._addDefault({
      kind: "ref" as const,
      ref: this._ref,
    } as unknown as SchemaNode);
  }
}
