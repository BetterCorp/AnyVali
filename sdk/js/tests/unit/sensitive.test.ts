import { describe, expect, it } from "vitest";
import {
  array, decrypt, encrypt, nullable, object, safeParseEncrypted, string,
} from "../../src/index.js";

const sensitive = <T extends ReturnType<typeof string>>(schema: T): T =>
  schema.describe("secret", { sensitive: true });

describe("sensitive data", () => {
  const schema = object({
    name: string(),
    secret: sensitive(string().minLength(3)),
    profile: object({ token: string() }).describe("profile", { sensitive: true }),
    aliases: array(sensitive(string())),
    note: nullable(sensitive(string())),
  });

  const plain = {
    name: "Ada",
    secret: "abc",
    profile: { token: "xyz" },
    aliases: ["one", "two"],
    note: null,
  };

  it("encrypts, validates, and decrypts sensitive leaves", () => {
    const seen: string[] = [];
    const encrypted = encrypt(schema, plain, (path, value) => {
      seen.push(path.join("."));
      return `encrypted:${JSON.stringify(value)}`;
    });

    expect(encrypted).toEqual({
      name: "Ada",
      secret: 'encrypted:"abc"',
      profile: 'encrypted:{"token":"xyz"}',
      aliases: ['encrypted:"one"', 'encrypted:"two"'],
      note: null,
    });
    expect(seen).toEqual(["secret", "profile", "aliases.0", "aliases.1"]);
    expect(safeParseEncrypted(schema, encrypted).success).toBe(true);
    expect(decrypt(schema, encrypted, (_path, value) =>
      JSON.parse((value as string).slice("encrypted:".length))
    )).toEqual(plain);
  });

  it("rejects broken envelopes and encrypt callbacks", () => {
    expect(safeParseEncrypted(schema, { ...plain, secret: "abc" }).success).toBe(false);
    expect(() => encrypt(schema, plain, () => "broken")).toThrow();
  });
});
