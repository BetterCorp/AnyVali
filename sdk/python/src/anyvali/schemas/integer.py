"""Integer schema types with explicit bit-width bounds."""

from __future__ import annotations

import math
from typing import Any

from ..issue_codes import INVALID_NUMBER, INVALID_TYPE, TOO_LARGE, TOO_SMALL
from .base import BaseSchema, ValidationContext, _anyvali_type_name

# Range definitions for each integer type
_INT_RANGES: dict[str, tuple[int, int]] = {
    "int8": (-(2**7), 2**7 - 1),
    "int16": (-(2**15), 2**15 - 1),
    "int32": (-(2**31), 2**31 - 1),
    "int64": (-(2**63), 2**63 - 1),
    "uint8": (0, 2**8 - 1),
    "uint16": (0, 2**16 - 1),
    "uint32": (0, 2**32 - 1),
    "uint64": (0, 2**64 - 1),
}


class _BaseIntSchema(BaseSchema):
    """Base integer schema with range checking and constraints."""

    _kind: str = "int64"
    _range_min: int
    _range_max: int

    def __init__(
        self,
        kind: str = "int64",
        *,
        min: int | None = None,
        max: int | None = None,
        exclusive_min: int | None = None,
        exclusive_max: int | None = None,
        multiple_of: int | None = None,
    ) -> None:
        super().__init__()
        self._kind = kind
        rng = _INT_RANGES.get(kind, _INT_RANGES["int64"])
        self._range_min, self._range_max = rng
        self._min = min
        self._max = max
        self._exclusive_min = exclusive_min
        self._exclusive_max = exclusive_max
        self._multiple_of = multiple_of

    def min(self, v: int) -> _BaseIntSchema:
        new = self._copy()
        assert isinstance(new, _BaseIntSchema)
        new._min = v
        return new

    def max(self, v: int) -> _BaseIntSchema:
        new = self._copy()
        assert isinstance(new, _BaseIntSchema)
        new._max = v
        return new

    def exclusive_min(self, v: int) -> _BaseIntSchema:
        new = self._copy()
        assert isinstance(new, _BaseIntSchema)
        new._exclusive_min = v
        return new

    def exclusive_max(self, v: int) -> _BaseIntSchema:
        new = self._copy()
        assert isinstance(new, _BaseIntSchema)
        new._exclusive_max = v
        return new

    def multiple_of(self, v: int) -> _BaseIntSchema:
        new = self._copy()
        assert isinstance(new, _BaseIntSchema)
        new._multiple_of = v
        return new

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        # Reject booleans (bool is subclass of int in Python)
        if isinstance(input, bool):
            ctx.add_issue(INVALID_TYPE, f"Expected integer, received boolean", expected=self._kind, received="boolean")
            return None

        if isinstance(input, float):
            if math.isnan(input) or math.isinf(input) or input != int(input):
                received = _anyvali_type_name(input)
                ctx.add_issue(INVALID_TYPE, f"Expected integer, received {received}", expected=self._kind, received=received)
                return None
            input = int(input)

        if not isinstance(input, int):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected integer, received {received}", expected=self._kind, received=received)
            return None

        # Range check for the specific integer width
        if input > self._range_max:
            ctx.add_issue(
                TOO_LARGE,
                f"Value {input} is out of {self._kind} range [{self._range_min}, {self._range_max}]",
                expected=self._kind,
                received=str(input),
            )
            return None

        if input < self._range_min:
            ctx.add_issue(
                TOO_SMALL,
                f"Value {input} is out of {self._kind} range [{self._range_min}, {self._range_max}]",
                expected=self._kind,
                received=str(input),
            )
            return None

        # User constraints
        if self._min is not None and input < self._min:
            ctx.add_issue(TOO_SMALL, f"Number must be >= {self._min}", expected=self._min, received=input)
        if self._max is not None and input > self._max:
            ctx.add_issue(TOO_LARGE, f"Number must be <= {self._max}", expected=self._max, received=input)
        if self._exclusive_min is not None and input <= self._exclusive_min:
            ctx.add_issue(TOO_SMALL, f"Number must be > {self._exclusive_min}", expected=self._exclusive_min, received=input)
        if self._exclusive_max is not None and input >= self._exclusive_max:
            ctx.add_issue(TOO_LARGE, f"Number must be < {self._exclusive_max}", expected=self._exclusive_max, received=input)
        if self._multiple_of is not None and self._multiple_of != 0:
            if input % self._multiple_of != 0:
                ctx.add_issue(INVALID_NUMBER, f"Number must be a multiple of {self._multiple_of}", expected=self._multiple_of, received=input)

        return input

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


class IntSchema(_BaseIntSchema):
    """int (alias for int64)."""

    def __init__(self, **kw: Any) -> None:
        super().__init__("int64", **kw)
        self._kind = "int"


class Int8Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("int8", **kw)


class Int16Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("int16", **kw)


class Int32Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("int32", **kw)


class Int64Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("int64", **kw)


class Uint8Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("uint8", **kw)


class Uint16Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("uint16", **kw)


class Uint32Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("uint32", **kw)


class Uint64Schema(_BaseIntSchema):
    def __init__(self, **kw: Any) -> None:
        super().__init__("uint64", **kw)
