import type { BaseSchema } from "./schemas/base.js";

/** Any AnyVali schema. Use as a generic constraint. Equivalent to Zod's ZodTypeAny. */
export type SchemaAny = BaseSchema<any, any>;

/** Extract the output type from a schema. Equivalent to Zod's z.infer. */
export type Infer<T extends SchemaAny> = T["_output"];

/** Extract the input type from a schema. */
export type InferInput<T extends SchemaAny> = T extends BaseSchema<infer I, any>
  ? I
  : never;
