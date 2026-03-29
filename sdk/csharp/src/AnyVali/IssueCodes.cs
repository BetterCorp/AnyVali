namespace AnyVali;

/// <summary>
/// Standard issue codes for validation failures.
/// </summary>
public static class IssueCodes
{
    public const string InvalidType = "invalid_type";
    public const string Required = "required";
    public const string UnknownKey = "unknown_key";
    public const string TooSmall = "too_small";
    public const string TooLarge = "too_large";
    public const string InvalidString = "invalid_string";
    public const string InvalidNumber = "invalid_number";
    public const string InvalidLiteral = "invalid_literal";
    public const string InvalidUnion = "invalid_union";
    public const string CustomValidationNotPortable = "custom_validation_not_portable";
    public const string UnsupportedExtension = "unsupported_extension";
    public const string UnsupportedSchemaKind = "unsupported_schema_kind";
    public const string CoercionFailed = "coercion_failed";
    public const string DefaultInvalid = "default_invalid";
}
