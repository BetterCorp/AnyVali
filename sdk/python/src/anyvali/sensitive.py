"""Explicit encrypted-storage helpers for ``sensitive`` schema nodes."""

from typing import Any, Callable, TypeVar

from .schemas.base import BaseSchema, ValidationContext
from .types import ParseResult, ValidationError

T = TypeVar("T")
SensitiveTransform = Callable[[list[str | int], Any], Any]


def _transform_sensitive(
    schema: BaseSchema,
    data: Any,
    mode: str,
    transform: SensitiveTransform,
) -> Any:
    result = schema._safe_parse_with_context(
        data,
        ValidationContext(
            sensitive_mode=mode,  # type: ignore[arg-type]
            sensitive_transform=transform,
            sensitive_cache={},
        ),
    )
    if not result.success:
        raise ValidationError(result.issues)
    return result.data


def safe_parse_encrypted(schema: BaseSchema, data: Any) -> ParseResult[Any]:
    """Validate encrypted storage data without decrypting sensitive values."""
    return schema._safe_parse_with_context(
        data, ValidationContext(sensitive_mode="encrypted")
    )


def encrypt(schema: BaseSchema, data: Any, transform: SensitiveTransform) -> Any:
    """Parse plaintext, transform sensitive values, then validate storage data."""
    encrypted = _transform_sensitive(schema, schema.parse(data), "encrypt", transform)
    result = safe_parse_encrypted(schema, encrypted)
    if not result.success:
        raise ValidationError(result.issues)
    return result.data


def decrypt(schema: BaseSchema[T], data: Any, transform: SensitiveTransform) -> T:
    """Validate storage data, transform sensitive values, then parse plaintext."""
    encrypted = safe_parse_encrypted(schema, data)
    if not encrypted.success:
        raise ValidationError(encrypted.issues)
    return schema.parse(_transform_sensitive(schema, encrypted.data, "decrypt", transform))
