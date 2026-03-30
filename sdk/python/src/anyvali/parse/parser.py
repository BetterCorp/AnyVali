"""Core parse pipeline.

The 5-step pipeline is implemented in BaseSchema._run_pipeline:
1. Presence check
2. Coercion (if present and configured)
3. Default application (if absent and configured)
4. Validation
5. Return result

This module provides top-level parse/safe_parse helpers.
"""

from __future__ import annotations

from typing import Any, TypeVar

from ..schemas.base import BaseSchema
from ..types import ParseResult, ValidationError

T = TypeVar("T")


def parse(schema: BaseSchema[T], input: Any) -> T:
    """Parse input against a schema, raising ValidationError on failure."""
    return schema.parse(input)


def safe_parse(schema: BaseSchema[T], input: Any) -> ParseResult[T]:
    """Parse input against a schema, returning a ParseResult."""
    return schema.safe_parse(input)
