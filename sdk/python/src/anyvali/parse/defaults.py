"""Default value application for the parse pipeline."""

from __future__ import annotations

import copy
from typing import Any

from ..schemas.base import _SENTINEL


def apply_default(value: Any, default: Any) -> Any:
    """Apply a default value if the input is absent (sentinel).

    Returns a deep copy of the default to avoid shared mutable state.
    """
    if value is _SENTINEL:
        return copy.deepcopy(default)
    return value
