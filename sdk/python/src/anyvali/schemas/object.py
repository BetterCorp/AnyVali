"""Object schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE, REQUIRED, UNKNOWN_KEY
from ..types import UnknownKeyMode
from .base import BaseSchema, ValidationContext, _SENTINEL, _anyvali_type_name


class ObjectSchema(BaseSchema[dict[str, Any]]):
    """Schema for objects/dicts with named properties."""

    def __init__(
        self,
        properties: dict[str, BaseSchema],
        *,
        required: list[str] | None = None,
        unknown_keys: UnknownKeyMode | None = None,
    ) -> None:
        super().__init__()
        self._properties = dict(properties)
        # If required is None, all properties are required by default
        self._required = set(required) if required is not None else set(properties.keys())
        self._unknown_keys = unknown_keys or "strip"
        self._unknown_keys_explicit = unknown_keys is not None

    def _effective_unknown_keys(self) -> UnknownKeyMode:
        return self._unknown_keys if self._unknown_keys_explicit else "strip"

    def _export_unknown_keys(self) -> UnknownKeyMode:
        return self._unknown_keys if self._unknown_keys_explicit else "strip"

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if not isinstance(input, dict):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected object, received {received}", expected="object", received=received)
            return None

        # Circular reference detection
        if not hasattr(ctx, '_seen'):
            ctx._seen = set()  # type: ignore[attr-defined]
        obj_id = id(input)
        if obj_id in ctx._seen:  # type: ignore[attr-defined]
            ctx.add_issue(INVALID_TYPE, "Circular reference detected", expected="object", received="circular")
            return None
        ctx._seen.add(obj_id)  # type: ignore[attr-defined]

        result: dict[str, Any] = {}

        # Validate known properties
        for key, schema in self._properties.items():
            if key in input:
                child_ctx = ctx.child(key)
                parsed = schema._run_pipeline(input[key], child_ctx)
                result[key] = parsed
            elif key in self._required:
                if schema._has_default:
                    # Apply default via pipeline
                    child_ctx = ctx.child(key)
                    parsed = schema._run_pipeline(_SENTINEL, child_ctx)
                    result[key] = parsed
                else:
                    child_ctx = ctx.child(key)
                    child_ctx.add_issue(REQUIRED, f"Required field '{key}' is missing", expected=key)
            else:
                # Optional and not present – check for defaults
                if schema._has_default:
                    child_ctx = ctx.child(key)
                    parsed = schema._run_pipeline(_SENTINEL, child_ctx)
                    result[key] = parsed

        # Handle unknown keys
        known = set(self._properties.keys())
        unknown = set(input.keys()) - known

        if unknown:
            mode = self._effective_unknown_keys()
            if mode == "reject":
                for key in sorted(unknown):
                    child_ctx = ctx.child(key)
                    child_ctx.add_issue(UNKNOWN_KEY, f"Unknown key '{key}'", received=key)
            elif mode == "strip":
                pass  # Just don't include them
            elif mode == "allow":
                for key in unknown:
                    result[key] = input[key]

        # Remove from the ancestor set now that this subtree is fully processed.
        # The guard tracks the current ancestor chain (true cycles), not every
        # object ever seen — otherwise a shared/repeated non-circular reference
        # in sibling positions would be falsely rejected as circular.
        ctx._seen.discard(obj_id)  # type: ignore[attr-defined]

        return result

    def _to_node(self) -> dict[str, Any]:
        props = {k: v._to_node() for k, v in self._properties.items()}
        node: dict[str, Any] = {
            "kind": "object",
            "properties": props,
            "required": sorted(self._required),
            "unknownKeys": self._export_unknown_keys(),
        }
        return self._add_common_node_fields(node)
