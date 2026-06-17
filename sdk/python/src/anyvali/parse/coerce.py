"""Coercion functions for the parse pipeline."""

from __future__ import annotations

import re
from typing import Any

from ..issue_codes import COERCION_FAILED
from ..schemas.base import CoercionConfig, ValidationContext

# Strict ASCII decimal grammars. Python's int()/float() are far more permissive
# than the ECMA-262 reference (JS): they accept Unicode digits ("１２３", "١٢٣"),
# digit-group underscores ("1_000"), a leading "+", and float()/int() accept
# "inf"/"nan". Gate on these before parsing so coercion behaves identically
# across SDKs (spec 5.1: parse as DECIMAL integer / floating-point only).
_DECIMAL_INT_RE = re.compile(r"^-?[0-9]+$")
_DECIMAL_FLOAT_RE = re.compile(r"^[+-]?(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$")


def apply_coercion(
    value: Any,
    config: CoercionConfig,
    ctx: ValidationContext,
    target: str | None = None,
) -> Any:
    """Apply configured coercions to a value.

    Coercion order: type coercions first, then string transformations.

    ``target`` is the schema's inferred coercion target ("int"/"number"/"bool"
    or None for string-typed schemas). When coercion is enabled but NO explicit
    ``to_*`` source flag is set, we infer the conversion from the schema's own
    kind and coerce from string. Mirrors the JS rule: "coercion enabled +
    non-string target + no explicit source => coerce from string." We do not
    auto-coerce when ``target`` is None (string kind): trim/lower/upper only.
    """
    result = value

    # Resolve effective type-coercion flags. Explicit to_* flags always win and
    # remain interoperable with the typed string->int/number/bool interchange
    # tokens; only when none is set do we fall back to the inferred target.
    to_int = config.to_int
    to_number = config.to_number
    to_bool = config.to_bool
    if not (to_int or to_number or to_bool) and target is not None:
        if target == "int":
            to_int = True
        elif target == "number":
            to_number = True
        elif target == "bool":
            to_bool = True

    # Type coercions (string -> target type)
    if to_int and isinstance(result, str):
        result = _coerce_to_int(result, ctx)
        if ctx.issues:
            return result

    if to_number and isinstance(result, str):
        result = _coerce_to_number(result, ctx)
        if ctx.issues:
            return result

    if to_bool and isinstance(result, str):
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
    if _DECIMAL_INT_RE.match(stripped):
        try:
            return int(stripped)
        except ValueError:
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
    if _DECIMAL_FLOAT_RE.match(stripped):
        try:
            f = float(stripped)
            # float() can still yield inf for huge magnitudes (e.g. "1e400");
            # NaN/inf are not decimal-representable JSON numbers (spec 5.1).
            if f == f and f not in (float("inf"), float("-inf")):
                return f
        except (ValueError, OverflowError):
            pass
    ctx.add_issue(
        COERCION_FAILED,
        f"Failed to coerce '{value}' to number",
        expected="number",
        received=value,
    )
    return value


def _coerce_to_bool(value: str, ctx: ValidationContext) -> Any:
    """Coerce a string to a boolean."""
    # Spec 5.1: only "true"/"1" => True and "false"/"0" => False (case-insensitive).
    # "yes"/"no" are NOT portable and diverge from the JS reference.
    lower = value.strip().lower()
    if lower in ("true", "1"):
        return True
    if lower in ("false", "0"):
        return False
    ctx.add_issue(
        COERCION_FAILED,
        f"Failed to coerce '{value}' to boolean",
        expected="boolean",
        received=value,
    )
    return value
