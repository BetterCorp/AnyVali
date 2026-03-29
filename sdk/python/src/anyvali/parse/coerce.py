"""Coercion functions for the parse pipeline."""

from __future__ import annotations

from typing import Any

from ..issue_codes import COERCION_FAILED
from ..schemas.base import CoercionConfig, ValidationContext


def apply_coercion(value: Any, config: CoercionConfig, ctx: ValidationContext) -> Any:
    """Apply configured coercions to a value.

    Coercion order: type coercions first, then string transformations.
    """
    result = value

    # Type coercions (string -> target type)
    if config.to_int and isinstance(result, str):
        result = _coerce_to_int(result, ctx)
        if ctx.issues:
            return result

    if config.to_number and isinstance(result, str):
        result = _coerce_to_number(result, ctx)
        if ctx.issues:
            return result

    if config.to_bool and isinstance(result, str):
        result = _coerce_to_bool(result, ctx)
        if ctx.issues:
            return result

    # String transformations (only apply to strings)
    if isinstance(result, str):
        if config.trim:
            result = result.strip()
        if config.lower:
            result = result.lower()
        if config.upper:
            result = result.upper()

    return result


def _coerce_to_int(value: str, ctx: ValidationContext) -> Any:
    """Coerce a string to an integer."""
    stripped = value.strip()
    try:
        # Try integer parse first
        return int(stripped)
    except ValueError:
        pass
    try:
        # Try float parse, then check if it's a whole number
        f = float(stripped)
        if f == int(f):
            return int(f)
    except (ValueError, OverflowError):
        pass
    ctx.add_issue(
        COERCION_FAILED,
        f"Failed to coerce '{value}' to integer",
        expected="integer",
        received=value,
    )
    return value


def _coerce_to_number(value: str, ctx: ValidationContext) -> Any:
    """Coerce a string to a float."""
    stripped = value.strip()
    try:
        return float(stripped)
    except ValueError:
        ctx.add_issue(
            COERCION_FAILED,
            f"Failed to coerce '{value}' to number",
            expected="number",
            received=value,
        )
        return value


def _coerce_to_bool(value: str, ctx: ValidationContext) -> Any:
    """Coerce a string to a boolean."""
    lower = value.strip().lower()
    if lower in ("true", "1", "yes"):
        return True
    if lower in ("false", "0", "no"):
        return False
    ctx.add_issue(
        COERCION_FAILED,
        f"Failed to coerce '{value}' to boolean",
        expected="boolean",
        received=value,
    )
    return value
