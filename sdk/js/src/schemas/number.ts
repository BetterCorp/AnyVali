import type { ParseContext, SchemaNode, SchemaKind } from "../types.js";
import { BaseSchema } from "./base.js";
import { ISSUE_CODES } from "../issue-codes.js";
import { describeType } from "../util.js";

export class NumberSchema extends BaseSchema<number, number> {
  protected _kind: SchemaKind;
  protected _min?: number;
  protected _max?: number;
  protected _exclusiveMin?: number;
  protected _exclusiveMax?: number;
  protected _multipleOf?: number;

  constructor(kind: SchemaKind = "number") {
    super();
    this._kind = kind;
  }

  _getCoercionTarget(): string {
    return this._kind;
  }

  min(n: number): this {
    const clone = this._clone();
    clone._min = n;
    return clone;
  }

  max(n: number): this {
    const clone = this._clone();
    clone._max = n;
    return clone;
  }

  exclusiveMin(n: number): this {
    const clone = this._clone();
    clone._exclusiveMin = n;
    return clone;
  }

  exclusiveMax(n: number): this {
    const clone = this._clone();
    clone._exclusiveMax = n;
    return clone;
  }

  multipleOf(n: number): this {
    const clone = this._clone();
    clone._multipleOf = n;
    return clone;
  }

  _validate(input: unknown, ctx: ParseContext): unknown {
    if (typeof input !== "number" || !Number.isFinite(input)) {
      ctx.issues.push({
        code: ISSUE_CODES.INVALID_TYPE,
        message: `Expected ${this._kind}, received ${describeType(input)}`,
        path: [...ctx.path],
        expected: this._kind,
        received: describeType(input),
      });
      return undefined;
    }

    this._validateConstraints(input, ctx);
    return input;
  }

  protected _validateConstraints(val: number, ctx: ParseContext): void {
    if (this._min !== undefined && val < this._min) {
      ctx.issues.push({
        code: ISSUE_CODES.TOO_SMALL,
        message: `Number must be >= ${this._min}`,
        path: [...ctx.path],
        expected: String(this._min),
        received: String(val),
      });
    }

    if (this._max !== undefined && val > this._max) {
      ctx.issues.push({
        code: ISSUE_CODES.TOO_LARGE,
        message: `Number must be <= ${this._max}`,
        path: [...ctx.path],
        expected: String(this._max),
        received: String(val),
      });
    }

    if (this._exclusiveMin !== undefined && val <= this._exclusiveMin) {
      ctx.issues.push({
        code: ISSUE_CODES.TOO_SMALL,
        message: `Number must be > ${this._exclusiveMin}`,
        path: [...ctx.path],
        expected: String(this._exclusiveMin),
        received: String(val),
      });
    }

    if (this._exclusiveMax !== undefined && val >= this._exclusiveMax) {
      ctx.issues.push({
        code: ISSUE_CODES.TOO_LARGE,
        message: `Number must be < ${this._exclusiveMax}`,
        path: [...ctx.path],
        expected: String(this._exclusiveMax),
        received: String(val),
      });
    }

    if (this._multipleOf !== undefined) {
      const remainder = val % this._multipleOf;
      if (
        Math.abs(remainder) > 1e-10 &&
        Math.abs(remainder - this._multipleOf) > 1e-10
      ) {
        ctx.issues.push({
          code: ISSUE_CODES.INVALID_NUMBER,
          message: `Number must be a multiple of ${this._multipleOf}`,
          path: [...ctx.path],
          expected: String(this._multipleOf),
          received: String(val),
        });
      }
    }
  }

  _toNode(): SchemaNode {
    const node: Record<string, unknown> = { kind: this._kind };
    if (this._min !== undefined) node.min = this._min;
    if (this._max !== undefined) node.max = this._max;
    if (this._exclusiveMin !== undefined)
      node.exclusiveMin = this._exclusiveMin;
    if (this._exclusiveMax !== undefined)
      node.exclusiveMax = this._exclusiveMax;
    if (this._multipleOf !== undefined) node.multipleOf = this._multipleOf;
    this._addDefault(node as unknown as SchemaNode);
    return node as unknown as SchemaNode;
  }
}

export class Float32Schema extends NumberSchema {
  constructor() {
    super("float32");
  }
}

export class Float64Schema extends NumberSchema {
  constructor() {
    super("float64");
  }
}
