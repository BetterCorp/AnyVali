"""Number/float schema types."""

from __future__ import annotations

import math
from typing import Any

from ..issue_codes import INVALID_NUMBER, INVALID_TYPE, TOO_LARGE, TOO_SMALL
from .base import BaseSchema, ValidationContext, _anyvali_type_name


class _BaseNumberSchema(BaseSchema):
    """Base for numeric schemas with constraints."""

    _kind: str = "float64"

    def __init__(
        self,
        *,
        min: float | int | None = None,
        max: float | int | None = None,
        exclusive_min: float | int | None = None,
        exclusive_max: float | int | None = None,
        multiple_of: float | int | None = None,
    ) -> None:
        super().__init__()
        self._min = min
        self._max = max
        self._exclusive_min = exclusive_min
        self._exclusive_max = exclusive_max
        self._multiple_of = multiple_of

    def min(self, v: float | int) -> _BaseNumberSchema:
        new = self._copy()
        assert isinstance(new, _BaseNumberSchema)
        new._min = v
        return new

    def max(self, v: float | int) -> _BaseNumberSchema:
        new = self._copy()
        assert isinstance(new, _BaseNumberSchema)
        new._max = v
        return new

    def exclusive_min(self, v: float | int) -> _BaseNumberSchema:
        new = self._copy()
        assert isinstance(new, _BaseNumberSchema)
        new._exclusive_min = v
        return new

    def exclusive_max(self, v: float | int) -> _BaseNumberSchema:
        new = self._copy()
        assert isinstance(new, _BaseNumberSchema)
        new._exclusive_max = v
        return new

    def multiple_of(self, v: float | int) -> _BaseNumberSchema:
        new = self._copy()
        assert isinstance(new, _BaseNumberSchema)
        new._multiple_of = v
        return new

    def _check_is_number(self, input: Any, ctx: ValidationContext) -> bool:
        if isinstance(input, bool):
            ctx.add_issue(INVALID_TYPE, f"Expected {self._kind}, received boolean", expected=self._kind, received="boolean")
            return False
        if not isinstance(input, (int, float)):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected {self._kind}, received {received}", expected=self._kind, received=received)
            return False
        if isinstance(input, float) and (math.isnan(input) or math.isinf(input)):
            ctx.add_issue(INVALID_NUMBER, "Number must be finite", expected="finite number", received=input)
            return False
        return True

    def _check_constraints(self, value: float | int, ctx: ValidationContext) -> None:
        if self._min is not None and value < self._min:
            ctx.add_issue(TOO_SMALL, f"Number must be >= {self._min}", expected=self._min, received=value)
        if self._max is not None and value > self._max:
            ctx.add_issue(TOO_LARGE, f"Number must be <= {self._max}", expected=self._max, received=value)
        if self._exclusive_min is not None and value <= self._exclusive_min:
            ctx.add_issue(TOO_SMALL, f"Number must be > {self._exclusive_min}", expected=self._exclusive_min, received=value)
        if self._exclusive_max is not None and value >= self._exclusive_max:
            ctx.add_issue(TOO_LARGE, f"Number must be < {self._exclusive_max}", expected=self._exclusive_max, received=value)
        if self._multiple_of is not None and self._multiple_of != 0:
            if value % self._multiple_of != 0:
                ctx.add_issue(INVALID_NUMBER, f"Number must be a multiple of {self._multiple_of}", expected=self._multiple_of, received=value)

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if not self._check_is_number(input, ctx):
            return None
        # Preserve the original numeric type (int stays int, float stays float)
        value = input
        self._check_range(value, ctx)
        self._check_constraints(value, ctx)
        return value

    def _check_range(self, value: float | int, ctx: ValidationContext) -> None:
        """Override in Float32Schema to check range."""
        pass

    def _to_node(self) -> dict[str, Any]:
        node: dict[str, Any] = {"kind": self._kind}
        if self._min is not None:
            node["min"] = self._min
        if self._max is not None:
            node["max"] = self._max
        if self._exclusive_min is not None:
            node["exclusiveMin"] = self._exclusive_min
        if self._exclusive_max is not None:
            node["exclusiveMax"] = self._exclusive_max
        if self._multiple_of is not None:
            node["multipleOf"] = self._multiple_of
        return self._add_common_node_fields(node)


class NumberSchema(_BaseNumberSchema):
    """number (alias for float64)."""

    _kind = "number"


class Float64Schema(_BaseNumberSchema):
    """float64 schema."""

    _kind = "float64"


class Float32Schema(_BaseNumberSchema):
    """float32 schema with range checks."""

    _kind = "float32"
    _FLOAT32_MAX = 3.4028235e+38

    def _check_range(self, value: float | int, ctx: ValidationContext) -> None:
        if abs(value) > self._FLOAT32_MAX and value != 0.0:
            ctx.add_issue(
                INVALID_NUMBER,
                f"Value {value} is out of float32 range",
                expected="float32",
                received=value,
            )
