# Sensitive Data

AnyVali can find and transform fields marked with `sensitive: true` without choosing a cipher, key manager, or storage system for you.

## JavaScript / TypeScript

```typescript
import { decrypt, encrypt, object, safeParseEncrypted, string } from "anyvali";

const Account = object({
  email: string().format("email"),
  token: string().minLength(16).describe("API token", { sensitive: true }),
});

const stored = encrypt(Account, input, (path, value) =>
  `encrypted:${encryptWithYourKms(path, value)}`,
);

const storageResult = safeParseEncrypted(Account, stored);

const account = decrypt(Account, stored, (path, value) =>
  decryptWithYourKms(path, value.slice("encrypted:".length)),
);
```

The callback receives `(path, value)`. A path is an array such as `["profile", "token"]`; array indexes are numbers. The encryption callback must return the complete `encrypted:<payload>` string. The decryption callback receives that complete string and returns the original value.

## Validation order

- `safeParseEncrypted(schema, data)` validates normal fields normally. Each non-null sensitive node must instead be a string beginning with `encrypted:` and containing a payload.
- `encrypt(schema, data, callback)` performs a normal parse, transforms sensitive nodes, then runs encrypted-state validation. A broken callback cannot silently produce an invalid storage object.
- `decrypt(schema, data, callback)` validates the encrypted state, transforms sensitive nodes, then performs a normal parse. A wrong or stale decrypted value is rejected by the original schema.

Normal `parse` and `safeParse` behavior is unchanged. They continue to validate the original plaintext schema.

When an object, array, tuple, or other composite node itself is sensitive, its whole value is passed to the callback once and its children are not visited. Missing optional fields stay missing, and `null` stays `null` when the schema permits it.

## APIs by SDK

| SDK | Encrypted validation | Encrypt | Decrypt |
|---|---|---|---|
| JavaScript / TypeScript | `safeParseEncrypted` | `encrypt` | `decrypt` |
| Python | `safe_parse_encrypted` | `encrypt` | `decrypt` |
| Go | `SafeParseEncrypted` | `Encrypt` | `Decrypt` |
| Rust | `safe_parse_encrypted` | `encrypt` | `decrypt` |
| C# | `V.SafeParseEncrypted` | `V.Encrypt` | `V.Decrypt` |
| Java | `AnyVali.safeParseEncrypted` | `AnyVali.encrypt` | `AnyVali.decrypt` |
| Kotlin | `safeParseEncrypted` | `encrypt` | `decrypt` |
| PHP | `AnyVali::safeParseEncrypted` | `AnyVali::encrypt` | `AnyVali::decrypt` |
| Ruby | `AnyVali.safe_parse_encrypted` | `AnyVali.encrypt` | `AnyVali.decrypt` |
| C++ | `safe_parse_encrypted` | `encrypt` | `decrypt` |

All initial APIs are synchronous. Exceptions or errors from the callback propagate to the caller.

## Security boundary

`encrypted:` is an envelope marker, not proof that data is encrypted or authentic. Use authenticated encryption (for example, an AEAD mode), keep keys outside the data, and include any key/version metadata your encryption system needs inside the payload. AnyVali deliberately stores no keys and implements no cryptography.
