"""Array schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE, TOO_LARGE, TOO_SMALL
from .base import BaseSchema, ValidationContext, _anyvali_type_name


class ArraySchema(BaseSchema):
    """Schema for arrays/lists with element validation."""

    def __init__(
        self,
        items: BaseSchema,
        *,
        min_items: int | None = None,
        max_items: int | None = None,
    ) -> None:
        super().__init__()
        self._items = items
        self._min_items = min_items
        self._max_items = max_items

    def min_items(self, n: int) -> ArraySchema:
        new = self._copy()
        assert isinstance(new, ArraySchema)
        new._min_items = n
        return new

    def max_items(self, n: int) -> ArraySchema:
        new = self._copy()
        assert isinstance(new, ArraySchema)
        new._max_items = n
        return new

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if not isinstance(input, list):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected array, received {received}", expected="array", received=received)
            return None

        if self._min_items is not None and len(input) < self._min_items:
            ctx.add_issue(TOO_SMALL, f"Array must have at least {self._min_items} item(s)", expected=self._min_items, received=len(input))

        if self._max_items is not None and len(input) > self._max_items:
            ctx.add_issue(TOO_LARGE, f"Array must have at most {self._max_items} item(s)", expected=self._max_items, received=len(input))

        result = []
        for i, item in enumerate(input):
            child_ctx = ctx.child(i)
            parsed = self._items._run_pipeline(item, child_ctx)
            result.append(parsed)

        return result

    def _to_node(self) -> dict[str, Any]:
        node: dict[str, Any] = {"kind": "array", "items": self._items._to_node()}
        if self._min_items is not None:
            node["minItems"] = self._min_items
        if self._max_items is not None:
            node["maxItems"] = self._max_items
        return self._add_common_node_fields(node)
