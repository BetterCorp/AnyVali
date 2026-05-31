"""Pattern regex caching + ReDoS defense-in-depth (Python SDK)."""

from __future__ import annotations

import anyvali as v


def test_match_and_reject_preserved():
    s = v.string().pattern("^[a-z]+$")
    assert s.safe_parse("abc").success is True
    bad = s.safe_parse("ABC")
    assert bad.success is False
    assert bad.issues[0].code == v.INVALID_STRING


def test_cached_regex_correct_across_many_values():
    s = v.array(v.string().pattern(r"^\d{3}$"))
    assert s.safe_parse(["123", "456", "789"]).success is True
    assert s.safe_parse(["123", "xx", "789"]).success is False


def test_invalid_pattern_reported():
    s = v.string().pattern("(")
    r = s.safe_parse("anything")
    assert r.success is False
    assert r.issues[0].code == v.INVALID_STRING
    assert "Invalid regex pattern" in r.issues[0].message


def test_repatterning_resets_cache():
    base = v.string().pattern("^a$")
    repat = base.pattern("^b$")
    assert base.safe_parse("a").success is True
    assert repat.safe_parse("b").success is True
    assert repat.safe_parse("a").success is False


def test_long_input_with_safe_pattern_validates():
    # Caching must not change behavior for legitimate large inputs.
    s = v.string().pattern("^x+$")
    big = "x" * 1_000_000
    assert s.safe_parse(big).success is True
