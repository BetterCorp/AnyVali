"""String schema."""

from __future__ import annotations

import re
from typing import Any

from ..issue_codes import INVALID_STRING, INVALID_TYPE, TOO_LARGE, TOO_SMALL
from .base import BaseSchema, ValidationContext, _anyvali_type_name


class StringSchema(BaseSchema[str]):
    """Schema for string values with optional constraints."""

    def __init__(
        self,
        *,
        min_length: int | None = None,
        max_length: int | None = None,
        pattern: str | None = None,
        starts_with: str | None = None,
        ends_with: str | None = None,
        includes: str | None = None,
        format_: str | None = None,
    ) -> None:
        super().__init__()
        self._min_length = min_length
        self._max_length = max_length
        self._pattern = pattern
        self._starts_with = starts_with
        self._ends_with = ends_with
        self._includes = includes
        self._format = format_

    def min_length(self, n: int) -> StringSchema:
        new = self._copy()
        assert isinstance(new, StringSchema)
        new._min_length = n
        return new

    def max_length(self, n: int) -> StringSchema:
        new = self._copy()
        assert isinstance(new, StringSchema)
        new._max_length = n
        return new

    def pattern(self, p: str) -> StringSchema:
        new = self._copy()
        assert isinstance(new, StringSchema)
        new._pattern = p
        return new

    def starts_with(self, s: str) -> StringSchema:
        new = self._copy()
        assert isinstance(new, StringSchema)
        new._starts_with = s
        return new

    def ends_with(self, s: str) -> StringSchema:
        new = self._copy()
        assert isinstance(new, StringSchema)
        new._ends_with = s
        return new

    def includes(self, s: str) -> StringSchema:
        new = self._copy()
        assert isinstance(new, StringSchema)
        new._includes = s
        return new

    def format(self, f: str) -> StringSchema:
        new = self._copy()
        assert isinstance(new, StringSchema)
        new._format = f
        return new

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if not isinstance(input, str):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected string, received {received}", expected="string", received=received)
            return None

        if self._min_length is not None and len(input) < self._min_length:
            ctx.add_issue(TOO_SMALL, f"String must have at least {self._min_length} character(s)", expected=self._min_length, received=len(input))

        if self._max_length is not None and len(input) > self._max_length:
            ctx.add_issue(TOO_LARGE, f"String must have at most {self._max_length} character(s)", expected=self._max_length, received=len(input))

        if self._pattern is not None and not re.search(self._pattern, input):
            ctx.add_issue(INVALID_STRING, f"String does not match pattern '{self._pattern}'", expected=self._pattern, received=input)

        if self._starts_with is not None and not input.startswith(self._starts_with):
            ctx.add_issue(INVALID_STRING, f"String must start with '{self._starts_with}'", expected=self._starts_with, received=input)

        if self._ends_with is not None and not input.endswith(self._ends_with):
            ctx.add_issue(INVALID_STRING, f"String must end with '{self._ends_with}'", expected=self._ends_with, received=input)

        if self._includes is not None and self._includes not in input:
            ctx.add_issue(INVALID_STRING, f"String must include '{self._includes}'", expected=self._includes, received=input)

        if self._format is not None:
            from ..format.validators import validate_format

            if not validate_format(self._format, input):
                ctx.add_issue(INVALID_STRING, f"Invalid {self._format}", expected=self._format, received=input)

        return input

    def _to_node(self) -> dict[str, Any]:
        node: dict[str, Any] = {"kind": "string"}
        if self._min_length is not None:
            node["minLength"] = self._min_length
        if self._max_length is not None:
            node["maxLength"] = self._max_length
        if self._pattern is not None:
            node["pattern"] = self._pattern
        if self._starts_with is not None:
            node["startsWith"] = self._starts_with
        if self._ends_with is not None:
            node["endsWith"] = self._ends_with
        if self._includes is not None:
            node["includes"] = self._includes
        if self._format is not None:
            node["format"] = self._format
        return self._add_common_node_fields(node)
