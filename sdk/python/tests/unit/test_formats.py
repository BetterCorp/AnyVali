"""Unit tests for all format validators with valid and invalid edge cases."""

from __future__ import annotations

import pytest

from anyvali.format.validators import validate_format


class TestEmailFormat:
    @pytest.mark.parametrize("value", [
        "user@example.com",
        "test.user@example.com",
        "user+tag@example.com",
        "a@b.co",
        "user@sub.domain.example.com",
        "user123@example.org",
    ])
    def test_valid(self, value: str):
        assert validate_format("email", value) is True

    @pytest.mark.parametrize("value", [
        "",
        "not-an-email",
        "@example.com",
        "user@",
        "user@.com",
        "user@com",
        "user @example.com",
        "user@@example.com",
    ])
    def test_invalid(self, value: str):
        assert validate_format("email", value) is False


class TestUrlFormat:
    @pytest.mark.parametrize("value", [
        "http://example.com",
        "https://example.com",
        "https://example.com/path",
        "https://sub.domain.example.com",
        "http://example.com:8080",
        "https://example.com/path/to/resource?q=1",
        "http://example.com:3000/api",
    ])
    def test_valid(self, value: str):
        assert validate_format("url", value) is True

    @pytest.mark.parametrize("value", [
        "",
        "not-a-url",
        "ftp://example.com",
        "example.com",
        "http://",
        "://example.com",
    ])
    def test_invalid(self, value: str):
        assert validate_format("url", value) is False


class TestUuidFormat:
    @pytest.mark.parametrize("value", [
        "550e8400-e29b-41d4-a716-446655440000",
        "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
        "00000000-0000-0000-0000-000000000000",
    ])
    def test_valid(self, value: str):
        assert validate_format("uuid", value) is True

    @pytest.mark.parametrize("value", [
        "",
        "not-a-uuid",
        "550e8400-e29b-41d4-a716",
        "gggggggg-gggg-gggg-gggg-gggggggggggg",
    ])
    def test_invalid(self, value: str):
        assert validate_format("uuid", value) is False


class TestIpv4Format:
    @pytest.mark.parametrize("value", [
        "0.0.0.0",
        "192.168.1.1",
        "255.255.255.255",
        "127.0.0.1",
        "10.0.0.1",
    ])
    def test_valid(self, value: str):
        assert validate_format("ipv4", value) is True

    @pytest.mark.parametrize("value", [
        "",
        "not-an-ip",
        "999.999.999.999",
        "256.0.0.0",
        "1.2.3",
        "1.2.3.4.5",
        "::1",
    ])
    def test_invalid(self, value: str):
        assert validate_format("ipv4", value) is False


class TestIpv6Format:
    @pytest.mark.parametrize("value", [
        "::1",
        "::ffff:192.168.1.1",
        "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
        "fe80::1",
        "::",
    ])
    def test_valid(self, value: str):
        assert validate_format("ipv6", value) is True

    @pytest.mark.parametrize("value", [
        "",
        "not-an-ipv6",
        "192.168.1.1",
        "gggg::1",
        "12345::1",
    ])
    def test_invalid(self, value: str):
        assert validate_format("ipv6", value) is False


class TestDateFormat:
    @pytest.mark.parametrize("value", [
        "2024-01-15",
        "2000-12-31",
        "1999-01-01",
        "2024-02-29",
    ])
    def test_valid(self, value: str):
        assert validate_format("date", value) is True

    @pytest.mark.parametrize("value", [
        "",
        "not-a-date",
        "2024/01/15",
        "01-15-2024",
        "2024-13-01",
        "2024-02-30",
        "2023-02-29",
        "24-01-15",
    ])
    def test_invalid(self, value: str):
        assert validate_format("date", value) is False


class TestDateTimeFormat:
    @pytest.mark.parametrize("value", [
        "2024-01-15T10:30:00Z",
        "2024-01-15T10:30:00+00:00",
        "2024-01-15T10:30:00-05:00",
        "2024-01-15T10:30:00.123Z",
        "2024-01-15T10:30:00.123456+02:00",
    ])
    def test_valid(self, value: str):
        assert validate_format("date-time", value) is True

    @pytest.mark.parametrize("value", [
        "",
        "not-a-datetime",
        "2024-01-15",
        "2024-01-15T10:30:00",
        "2024-01-15 10:30:00",
        "2024-13-15T10:30:00Z",
    ])
    def test_invalid(self, value: str):
        assert validate_format("date-time", value) is False


class TestUnknownFormat:
    def test_unknown_format_passes(self):
        assert validate_format("unknown-format", "anything") is True

    def test_custom_format_passes(self):
        assert validate_format("x-custom", "value") is True
