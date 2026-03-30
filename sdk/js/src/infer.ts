import type { BaseSchema } from "./schemas/base.js";

/** Extract the output type from a schema. Equivalent to Zod's z.infer. */
export type Infer<T extends BaseSchema<any, any>> = T["_output"];

/** Extract the input type from a schema. */
export type InferInput<T extends BaseSchema<any, any>> = T extends BaseSchema<
  infer I,
  any
>
  ? I
  : never;
