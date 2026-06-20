import { describe, it, expect } from "vitest";
import { any, string, int, object, optional, nullable, bool, exportSchema } from "../../src/index.js";

describe("Defaults", () => {
  it("applies default when value is absent", () => {
    const s = string().default("fallback");
    expect(s.parse(undefined)).toBe("fallback");
  });

  it("does not apply default when value is present", () => {
    const s = string().default("fallback");
    expect(s.parse("provided")).toBe("provided");
  });

  it("validates the defaulted value", () => {
    const s = string().minLength(5).default("hi");
    // default "hi" is only 2 chars, should fail validation with default_invalid
    const result = s.safeParse(undefined);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("default_invalid");
    }
  });

  it("applies defaults in object fields", () => {
    const s = object({
      name: string(),
      role: string().default("user"),
    });
    const result = s.parse({ name: "Alice" });
    expect(result.role).toBe("user");
  });

  it("does not override present object field with default", () => {
    const s = object({
      role: string().default("user"),
    });
    const result = s.parse({ role: "admin" });
    expect(result.role).toBe("admin");
  });

  it("default on optional field", () => {
    const s = object({
      count: optional(int().default(0)),
    });
    const result = s.parse({});
    expect(result).toEqual({ count: 0 });
  });

  it("applies default on optional wrapper field", () => {
    const s = object({
      host: optional(string()).default("localhost"),
    });
    expect(s.parse({})).toEqual({ host: "localhost" });
  });

  it("does not override present optional wrapper field with default", () => {
    const s = object({
      host: optional(string()).default("localhost"),
    });
    expect(s.parse({ host: "example.com" })).toEqual({ host: "example.com" });
  });

  it("validates default on optional wrapper field", () => {
    const s = object({
      host: optional(string().minLength(5)).default("hi"),
    });
    const result = s.safeParse({});
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.issues[0].code).toBe("default_invalid");
      expect(result.issues[0].path).toEqual(["host"]);
    }
  });

  it("exports default on optional wrapper", () => {
    const doc = exportSchema(optional(string()).default("localhost"));
    expect(doc.root.kind).toBe("optional");
    expect(doc.root.default).toBe("localhost");
  });

  it("clones mutable default on optional wrapper field", () => {
    const s = object({
      config: optional(any()).default({ tags: [] as string[] }),
    });
    const first = s.parse({}) as any;
    first.config.tags.push("mutated");
    expect(s.parse({})).toEqual({ config: { tags: [] } });
  });

  it("does not apply default when nullable field is null", () => {
    const s = object({
      value: nullable(string()).default("fallback"),
    });
    expect(s.parse({ value: null })).toEqual({ value: null });
  });

  it("applies falsy defaults", () => {
    const s = object({
      count: int().default(0),
      name: string().default(""),
      active: bool().default(false),
    });
    expect(s.parse({})).toEqual({ count: 0, name: "", active: false });
  });

  it("applies nested defaults", () => {
    const s = object({
      user: object({
        name: string(),
        role: string().default("guest"),
      }),
    });
    expect(s.parse({ user: { name: "Bob" } })).toEqual({
      user: { name: "Bob", role: "guest" },
    });
  });

  it("applies required object-field defaults across scalar types", () => {
    const s = object({
      host: string().minLength(1).default("0.0.0.0"),
      port: int().min(1).default(3200),
      timeoutMs: int().default(5000).min(1000),
      enabled: bool().default(false),
    });

    expect(s.parse({})).toEqual({
      host: "0.0.0.0",
      port: 3200,
      timeoutMs: 5000,
      enabled: false,
    });
  });

  it("applies defaults in optional inner and optional wrapper forms", () => {
    const s = object({
      innerDefault: optional(string().default("inner")),
      wrapperDefault: optional(string()).default("wrapper"),
    });

    expect(s.parse({})).toEqual({
      innerDefault: "inner",
      wrapperDefault: "wrapper",
    });
  });

  it("applies defaults in a plugin-style config object", () => {
    const BetterPortalConfigSchema = object({
      tenantId: string().minLength(1),
      authressApiUrl: string().minLength(1),
      authressApplicationId: string().minLength(1),
    });
    const PluginConfigSchema = object({
      host: string().minLength(1).default("0.0.0.0"),
      port: int().min(1).default(3200),
      betterportal: BetterPortalConfigSchema,
      dbLocation: string().minLength(1).default("data"),
      redisUrl: string().minLength(1).default("redis://localhost:6379"),
      gotenbergBaseUrl: string().minLength(1).default("http://localhost:3000"),
      gotenbergTimeoutMs: int().default(5000).min(1000),
    }, { unknownKeys: "strip" });

    expect(PluginConfigSchema.parse({
      port: 3211,
      betterportal: {
        tenantId: "tenant-lorem",
        authressApiUrl: "https://auth.example.test",
        authressApplicationId: "app_loremIpsum123",
      },
      redisUrl: "redis://10.1.1.4:6379",
      gotenbergBaseUrl: "http://10.1.1.16:3000",
      extra: true,
    })).toEqual({
      host: "0.0.0.0",
      port: 3211,
      betterportal: {
        tenantId: "tenant-lorem",
        authressApiUrl: "https://auth.example.test",
        authressApplicationId: "app_loremIpsum123",
      },
      dbLocation: "data",
      redisUrl: "redis://10.1.1.4:6379",
      gotenbergBaseUrl: "http://10.1.1.16:3000",
      gotenbergTimeoutMs: 5000,
    });
  });
});
