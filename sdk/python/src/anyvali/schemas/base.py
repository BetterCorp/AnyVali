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

from ..issue_codes import COERCION_FAILED, DEFAULT_INVALID, TOO_DEEP

# Maximum validation recursion depth. Bounds recursion through recursive
# ``$ref`` schemas (and deeply nested data) so a crafted payload cannot
# exhaust the call stack. Kept below CPython's native recursion limit so the
# guard trips deterministically before a RecursionError.
MAX_DEPTH = 200
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
    depth: int = 0
    inherited_unknown_keys: str | None = None

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
            depth=self.depth,
            inherited_unknown_keys=self.inherited_unknown_keys,
        )
        # Propagate circular reference tracker
        if hasattr(self, '_seen'):
            ctx._seen = self._seen  # type: ignore[attr-defined]
        return ctx


_SENTINEL = object()

_RESERVED_METADATA_KEYS = frozenset({
    'title', 'description', 'deprecated', 'deprecatedMessage',
    'notStable', 'since', 'sensitive', 'readonly', 'writeonly', 'examples',
})


class BaseSchema(ABC, Generic[T]):
    """Abstract base for all schema types."""

    _coercion: CoercionConfig | None
    _default_value: Any
    _has_default: bool
    _metadata: dict[str, Any] | None

    def __init__(self) -> None:
        self._coercion = None
        self._default_value = _SENTINEL
        self._has_default = False
        self._metadata = None

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
        """Parse input, returning a ParseResult.

        Never raises: the depth guard bounds recursion, and a RecursionError
        backstop converts any residual stack exhaustion into a TOO_DEEP issue
        so a crafted deeply nested payload cannot crash the caller (DoS).
        """
        ctx = ValidationContext()
        try:
            value = self._run_pipeline(input, ctx)
        except RecursionError:
            return ParseResult(
                success=False,
                issues=[
                    ValidationIssue(
                        code=TOO_DEEP,
                        message="Maximum validation depth exceeded",
                        path=[],
                        expected="bounded nesting",
                        received="too deep",
                    )
                ],
            )
        if ctx.issues:
            return ParseResult(success=False, issues=ctx.issues)
        return ParseResult(success=True, data=value)

    # ── 5-step pipeline ───────────────────────────────────────────

    def _run_pipeline(self, input: Any, ctx: ValidationContext) -> Any:
        # Depth guard: bound recursion (recursive $ref + deep data) so a
        # crafted payload cannot exhaust the call stack. Returns a clean issue
        # instead of letting safe_parse raise RecursionError.
        ctx.depth += 1
        if ctx.depth > MAX_DEPTH:
            ctx.add_issue(
                TOO_DEEP,
                f"Maximum validation depth of {MAX_DEPTH} exceeded",
                expected=f"<= {MAX_DEPTH} levels",
                received="too deep",
            )
            ctx.depth -= 1
            return None
        try:
            return self._run_pipeline_inner(input, ctx)
        finally:
            ctx.depth -= 1

    def _run_pipeline_inner(self, input: Any, ctx: ValidationContext) -> Any:
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
        result = apply_coercion(
            value, self._coercion, ctx, target=self._coercion_target()
        )
        return result

    def _coercion_target(self) -> str | None:
        """Inferred coercion target for the bare ``.coerce()`` form.

        Returns the portable target ("int" / "number" / "bool") this schema's
        kind expects a string to be coerced into when ``.coerce()`` is called
        with no explicit ``to_*`` source flag. String-typed schemas return None
        (trim/lower/upper only). Mirrors the JS rule: "coercion enabled +
        non-string target + no explicit source => coerce from string."
        """
        return None

    # ── Private helpers ────────────────────────────────────────────

    def _validate_describe_opts(
        self,
        description: str,
        *,
        title: str | None = None,
        deprecated: bool | None = None,
        deprecated_message: str | None = None,
        not_stable: bool | None = None,
        since: str | None = None,
        sensitive: bool | None = None,
        readonly: bool | None = None,
        writeonly: bool | None = None,
        examples: list | None = None,
    ) -> dict[str, Any]:
        if not isinstance(description, str):
            raise TypeError("describe(): description must be a string")

        meta: dict[str, Any] = {"description": description}

        if title is not None:
            if not isinstance(title, str):
                raise TypeError("describe(): title must be a string")
            meta["title"] = title
        if deprecated is not None:
            if not isinstance(deprecated, bool):
                raise TypeError("describe(): deprecated must be a bool")
            meta["deprecated"] = deprecated
        if deprecated_message is not None:
            if not isinstance(deprecated_message, str):
                raise TypeError("describe(): deprecatedMessage must be a string")
            if not deprecated:
                raise ValueError("describe(): deprecatedMessage requires deprecated=True")
            meta["deprecatedMessage"] = deprecated_message
        if not_stable is not None:
            if not isinstance(not_stable, bool):
                raise TypeError("describe(): notStable must be a bool")
            meta["notStable"] = not_stable
        if since is not None:
            if not isinstance(since, str):
                raise TypeError("describe(): since must be a string")
            meta["since"] = since
        if sensitive is not None:
            if not isinstance(sensitive, bool):
                raise TypeError("describe(): sensitive must be a bool")
            meta["sensitive"] = sensitive
        if readonly is not None:
            if not isinstance(readonly, bool):
                raise TypeError("describe(): readonly must be a bool")
            meta["readonly"] = readonly
        if writeonly is not None:
            if not isinstance(writeonly, bool):
                raise TypeError("describe(): writeonly must be a bool")
            meta["writeonly"] = writeonly
        if readonly and writeonly:
            raise ValueError("describe(): readonly and writeonly cannot both be True")
        if examples is not None:
            if not isinstance(examples, list):
                raise TypeError("describe(): examples must be a list")
            meta["examples"] = examples

        return meta

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

    def describe(
        self,
        description: str,
        *,
        title: str | None = None,
        deprecated: bool | None = None,
        deprecated_message: str | None = None,
        not_stable: bool | None = None,
        since: str | None = None,
        sensitive: bool | None = None,
        readonly: bool | None = None,
        writeonly: bool | None = None,
        examples: list | None = None,
    ) -> BaseSchema[T]:
        reserved_meta = self._validate_describe_opts(
            description,
            title=title,
            deprecated=deprecated,
            deprecated_message=deprecated_message,
            not_stable=not_stable,
            since=since,
            sensitive=sensitive,
            readonly=readonly,
            writeonly=writeonly,
            examples=examples,
        )
        new = self._copy()
        existing = new._metadata or {}
        new._metadata = {**existing, **reserved_meta}
        return new

    def metadata(
        self,
        meta: dict[str, Any],
        *,
        replace: bool = False,
    ) -> BaseSchema[T]:
        """Attach arbitrary metadata to this schema. Reserved keys must use describe()."""
        for key in meta:
            if key in _RESERVED_METADATA_KEYS:
                raise ValueError(
                    f'metadata(): "{key}" is a reserved key. Use describe() instead.'
                )
        new = self._copy()
        if replace:
            # Keep reserved keys from existing metadata, replace the rest
            existing = new._metadata or {}
            reserved = {k: v for k, v in existing.items() if k in _RESERVED_METADATA_KEYS}
            new._metadata = {**reserved, **meta}
        else:
            # Shallow merge (default)
            existing = new._metadata or {}
            new._metadata = {**existing, **meta}
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
        if self._metadata:
            node["metadata"] = dict(self._metadata)
        return node
