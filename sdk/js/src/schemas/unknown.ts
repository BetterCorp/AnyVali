import type { ParseContext, SchemaNode } from "../types.js";
import { BaseSchema } from "./base.js";

export class UnknownSchema extends BaseSchema<unknown, unknown> {
  _validate(input: unknown, _ctx: ParseContext): unknown {
    return input;
  }

  _toNode(): SchemaNode {
    const node: SchemaNode = { kind: "unknown" } as SchemaNode;
    this._addDefault(node);
    return node;
  }
}
