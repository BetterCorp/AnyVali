"""Shared (non-circular) object references must not be flagged as circular."""

from __future__ import annotations

import anyvali as v


def test_same_object_in_two_fields_validates():
    inner = v.object_({"a": v.int_()})
    s = v.object_({"x": inner, "y": inner})
    shared = {"a": 1}
    r = s.safe_parse({"x": shared, "y": shared})
    assert r.success is True


def test_same_object_repeated_in_array_validates():
    s = v.array(v.object_({"a": v.int_()}))
    shared = {"a": 1}
    r = s.safe_parse([shared, shared, shared])
    assert r.success is True


def test_true_cycle_still_rejected():
    s = v.object_({"self": v.object_({"a": v.int_()})})
    cyclic: dict = {}
    cyclic["self"] = cyclic
    r = s.safe_parse(cyclic)  # must not hang/crash
    assert r.success is False
