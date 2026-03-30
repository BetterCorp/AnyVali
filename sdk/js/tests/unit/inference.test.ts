import { describe, it, expect, expectTypeOf } from "vitest";
import {
  object,
  string,
  int,
  int64,
  number,
  bool,
  null_,
  array,
  tuple,
  record,
  union,
  intersection,
  optional,
  nullable,
  literal,
  enum_,
  type Infer,
} from "../../src/index.js";

describe("Type Inference", () => {
  describe("Infer<T> utility type", () => {
    it("infers string", () => {
      const schema = string();
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string>();
    });

    it("infers number", () => {
      const schema = number();
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<number>();
    });

    it("infers int as number", () => {
      const schema = int();
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<number>();
    });

    it("infers boolean", () => {
      const schema = bool();
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<boolean>();
    });

    it("infers null", () => {
      const schema = null_();
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<null>();
    });

    it("infers literal string", () => {
      const schema = literal("hello");
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<"hello">();
    });

    it("infers literal number", () => {
      const schema = literal(42);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<42>();
    });
  });

  describe("Object type inference", () => {
    it("infers flat object shape", () => {
      const schema = object({
        name: string(),
        age: int(),
      });
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<{ name: string; age: number }>();
    });

    it("infers object with optional fields", () => {
      const schema = object({
        name: string(),
        nick: optional(string()),
      });
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<{
        name: string;
        nick?: string | undefined;
      }>();
    });

    it("infers object with nullable fields", () => {
      const schema = object({
        name: string(),
        bio: nullable(string()),
      });
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<{
        name: string;
        bio: string | null;
      }>();
    });

    it("infers nested objects", () => {
      const schema = object({
        user: object({
          name: string(),
          age: int(),
        }),
      });
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<{
        user: { name: string; age: number };
      }>();
    });

    it("parse returns typed value, not Record<string, unknown>", () => {
      const schema = object({
        a: string(),
      });
      const result = schema.parse({ a: "hello" });
      expectTypeOf(result).toEqualTypeOf<{ a: string }>();

      // This should be accessible:
      expectTypeOf(result.a).toEqualTypeOf<string>();
    });

    it("safeParse data is typed on success", () => {
      const schema = object({
        a: string(),
        b: int(),
      });
      const result = schema.safeParse({ a: "hello", b: 1 });
      if (result.success) {
        expectTypeOf(result.data).toEqualTypeOf<{
          a: string;
          b: number;
        }>();
      }
    });

    it("rejects accessing unknown properties at type level", () => {
      const schema = object({ a: string() });
      type T = Infer<typeof schema>;
      // T should have 'a' but not 'b'
      expectTypeOf<T>().toHaveProperty("a");
      expectTypeOf<T>().not.toHaveProperty("b");
    });
  });

  describe("Array type inference", () => {
    it("infers array of strings", () => {
      const schema = array(string());
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string[]>();
    });

    it("infers array of objects", () => {
      const schema = array(
        object({
          id: int64(),
          name: string(),
        }),
      );
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<
        { id: number; name: string }[]
      >();
    });
  });

  describe("Tuple type inference", () => {
    it("infers tuple element types", () => {
      const schema = tuple([string(), int(), bool()]);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<[string, number, boolean]>();
    });
  });

  describe("Record type inference", () => {
    it("infers record value type", () => {
      const schema = record(number());
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<Record<string, number>>();
    });
  });

  describe("Union type inference", () => {
    it("infers union of types", () => {
      const schema = union([string(), int()]);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string | number>();
    });

    it("infers union of literals", () => {
      const schema = union([literal("a"), literal("b")]);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<"a" | "b">();
    });
  });

  describe("Intersection type inference", () => {
    it("infers intersection of object types", () => {
      const schema = intersection([
        object({ name: string() }),
        object({ age: int() }),
      ]);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<{
        name: string;
        age: number;
      }>();
    });
  });

  describe("Enum type inference", () => {
    it("infers enum values as union", () => {
      const schema = enum_(["free", "pro", "enterprise"] as const);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<"free" | "pro" | "enterprise">();
    });
  });

  describe("Modifier type inference", () => {
    it("optional adds undefined", () => {
      const schema = optional(string());
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string | undefined>();
    });

    it("nullable adds null", () => {
      const schema = nullable(string());
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string | null>();
    });

    it("optional nullable adds both", () => {
      const schema = optional(nullable(string()));
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string | null | undefined>();
    });
  });

  describe("Method chaining preserves types", () => {
    it("string with constraints stays string", () => {
      const schema = string().minLength(1).maxLength(100).format("email");
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string>();
    });

    it("int with constraints stays number", () => {
      const schema = int().min(0).max(100);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<number>();
    });

    it("array with constraints preserves element type", () => {
      const schema = array(string()).minItems(1).maxItems(10);
      type T = Infer<typeof schema>;
      expectTypeOf<T>().toEqualTypeOf<string[]>();
    });
  });

  describe("Runtime behavior unchanged", () => {
    it("object parse still works correctly", () => {
      const schema = object({ a: string() });
      const result = schema.parse({ a: "hello" });
      expect(result).toEqual({ a: "hello" });
      expect(result.a).toBe("hello");
    });

    it("object safeParse still works correctly", () => {
      const schema = object({ a: string(), b: int() });
      const result = schema.safeParse({ a: "hello", b: 42 });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.a).toBe("hello");
        expect(result.data.b).toBe(42);
      }
    });

    it("object safeParse rejects invalid and unknown properties", () => {
      const schema = object({ a: string() });
      const result = schema.safeParse({ a: "hello", b: "extra" });
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.issues[0].code).toBe("unknown_key");
      }
    });

    it("typed parse result correctly narrows - accessing known property is ok", () => {
      const schema = object({ a: string() });
      const res = schema.parse({ a: "a" });
      // res.a should work fine at runtime and type level
      expect(res.a).toBe("a");
      expectTypeOf(res.a).toEqualTypeOf<string>();
    });
  });
});
