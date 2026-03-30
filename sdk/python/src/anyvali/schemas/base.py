"""Base schema class for all AnyVali schema types."""

from __future__ import annotations

import copy
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import TYPE_CHECKING, Any, Generic, TypeVar

if TYPE_CHECKING:
    from .nullable import NullableSchema
    from .optional import OptionalSchema

T = TypeVar("T")

from ..issue_codes import COERCION_FAILED, DEFAULT_INVALID
from ..types import (
    AnyValiDocument,
    ExportMode,
    ParseResult,
    ValidationError,
    ValidationIssue,
)


def _anyvali_type_name(value: Any) -> str:
    """Map a Python value to its AnyVali portable type name."""
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, int):
        return "integer"
    if isinstance(value, float):
        return "number"
    if isinstance(value, str):
        return "string"
    if isinstance(value, list):
        return "array"
    if isinstance(value, dict):
        return "object"
    return type(value).__name__


@dataclass(frozen=True)
class CoercionConfig:
    """Configuration for coercion behavior."""

    to_int: bool = False
    to_number: bool = False
    to_bool: bool = False
    trim: bool = False
    lower: bool = False
    upper: bool = False


@dataclass
class ValidationContext:
    """Context accumulated during validation."""

    path: list[str | int] = field(default_factory=list)
    issues: list[ValidationIssue] = field(default_factory=list)
    definitions: dict[str, Any] = field(default_factory=dict)

    def add_issue(
        self,
        code: str,
        message: str,
        *,
        expected: Any = None,
        received: Any = None,
        meta: dict[str, Any] | None = None,
    ) -> None:
        self.issues.append(
            ValidationIssue(
                code=code,
                message=message,
                path=list(self.path),
                expected=expected,
                received=received,
                meta=meta,
            )
        )

    def child(self, key: str | int) -> ValidationContext:
        ctx = ValidationContext(
            path=[*self.path, key],
            issues=self.issues,
            definitions=self.definitions,
        )
        return ctx


_SENTINEL = object()


class BaseSchema(ABC, Generic[T]):
    """Abstract base for all schema types."""

    _coercion: CoercionConfig | None
    _default_value: Any
    _has_default: bool

    def __init__(self) -> None:
        self._coercion = None
        self._default_value = _SENTINEL
        self._has_default = False

    def _copy(self) -> BaseSchema[T]:
        return copy.deepcopy(self)

    # ── Public parse API ──────────────────────────────────────────

    def parse(self, input: Any) -> T:
        """Parse input, raising ValidationError on failure."""
        result = self.safe_parse(input)
        if not result.success:
            raise ValidationError(result.issues)
        return result.data  # type: ignore[return-value]

    def safe_parse(self, input: Any) -> ParseResult[T]:
        """Parse input, returning a ParseResult."""
        ctx = ValidationContext()
        value = self._run_pipeline(input, ctx)
        if ctx.issues:
            return ParseResult(success=False, issues=ctx.issues)
        return ParseResult(success=True, data=value)

    # ── 5-step pipeline ───────────────────────────────────────────

    def _run_pipeline(self, input: Any, ctx: ValidationContext) -> Any:
        # Step 1: presence check
        is_absent = input is _SENTINEL or input is None and not self._accepts_none()

        # For optional/nullable wrappers, None is 'present'
        if input is None and self._accepts_none():
            is_absent = False

        # Step 2: coercion (only if present)
        value = input
        if not is_absent and self._coercion is not None:
            value = self._apply_coercion(value, ctx)
            if ctx.issues:
                return None

        # Step 3: defaults (only if absent)
        used_default = False
        if is_absent and self._has_default:
            value = copy.deepcopy(self._default_value)
            is_absent = False
            used_default = True

        # If still absent, validate will handle required errors
        if is_absent and not self._has_default:
            # Let _validate handle it (it should produce REQUIRED or INVALID_TYPE)
            pass

        # Step 4: validate
        if used_default:
            # Validate default in isolation so we can replace issues with default_invalid
            issues_before = len(ctx.issues)
            result = self._validate(value, ctx)
            if len(ctx.issues) > issues_before:
                # Replace the new issues with a single default_invalid issue
                new_issues = ctx.issues[issues_before:]
                del ctx.issues[issues_before:]
                ctx.add_issue(
                    DEFAULT_INVALID,
                    f"Default value {self._default_value!r} is invalid",
                    expected=str(new_issues[0].expected) if new_issues else None,
                    received=str(self._default_value),
                )
                return None
        else:
            result = self._validate(value, ctx)

        # Step 5: return
        return result

    def _accepts_none(self) -> bool:
        """Override in NullSchema and NullableSchema."""
        return False

    def _apply_coercion(self, value: Any, ctx: ValidationContext) -> Any:
        """Apply configured coercions to the value."""
        from ..parse.coerce import apply_coercion

        if self._coercion is None:
            return value
        result = apply_coercion(value, self._coercion, ctx)
        return result

    # ── Abstract ──────────────────────────────────────────────────

    @abstractmethod
    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        """Validate and return parsed value, adding issues to ctx."""
        ...

    @abstractmethod
    def _to_node(self) -> dict[str, Any]:
        """Return the schema node dict for interchange."""
        ...

    # ── Modifier methods (return new instances) ───────────────────

    def optional(self) -> OptionalSchema[T]:
        from .optional import OptionalSchema

        return OptionalSchema(self)

    def nullable(self) -> NullableSchema[T]:
        from .nullable import NullableSchema

        return NullableSchema(self)

    def default(self, value: Any) -> BaseSchema[T]:
        new = self._copy()
        new._default_value = value
        new._has_default = True
        return new

    def coerce(
        self,
        *,
        to_int: bool = False,
        to_number: bool = False,
        to_bool: bool = False,
        trim: bool = False,
        lower: bool = False,
        upper: bool = False,
    ) -> BaseSchema[T]:
        new = self._copy()
        new._coercion = CoercionConfig(
            to_int=to_int,
            to_number=to_number,
            to_bool=to_bool,
            trim=trim,
            lower=lower,
            upper=upper,
        )
        return new

    # ── Export ─────────────────────────────────────────────────────

    def export(self, mode: ExportMode = "portable") -> dict[str, Any]:
        """Export this schema as an AnyVali document dict."""
        node = self._to_node()
        doc = AnyValiDocument(root=node)
        return doc.to_dict()

    def _add_common_node_fields(self, node: dict[str, Any]) -> dict[str, Any]:
        """Add default/coercion fields to a node dict."""
        if self._has_default:
            node["default"] = self._default_value
        if self._coercion is not None:
            coerce_dict: dict[str, Any] = {}
            if self._coercion.to_int:
                coerce_dict["toInt"] = True
            if self._coercion.to_number:
                coerce_dict["toNumber"] = True
            if self._coercion.to_bool:
                coerce_dict["toBool"] = True
            if self._coercion.trim:
                coerce_dict["trim"] = True
            if self._coercion.lower:
                coerce_dict["lower"] = True
            if self._coercion.upper:
                coerce_dict["upper"] = True
            if coerce_dict:
                node["coerce"] = coerce_dict
        return node
