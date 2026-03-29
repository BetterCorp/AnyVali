"""Shared test fixtures for AnyVali Python SDK tests."""

from __future__ import annotations

import sys
from pathlib import Path

import pytest

# Ensure the src directory is on the path for imports
_src = Path(__file__).resolve().parent.parent / "src"
if str(_src) not in sys.path:
    sys.path.insert(0, str(_src))

import anyvali as v


@pytest.fixture
def string_schema():
    return v.string()


@pytest.fixture
def number_schema():
    return v.number()


@pytest.fixture
def int_schema():
    return v.int_()


@pytest.fixture
def bool_schema():
    return v.bool_()


@pytest.fixture
def null_schema():
    return v.null()


@pytest.fixture
def user_schema():
    """A typical user object schema for integration tests."""
    return v.object_({
        "id": v.int_().min(1),
        "name": v.string().min_length(1),
        "email": v.string().format("email"),
        "age": v.int_().min(0).max(150).optional(),
    })
