"""Format validators using only Python stdlib."""

from __future__ import annotations

import ipaddress
import re
import uuid as _uuid_mod
from datetime import date, datetime

# Basic email regex – covers common cases without external deps.
_EMAIL_RE = re.compile(
    r"^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9]"
    r"(?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?"
    r"(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$"
)

# URL regex – accepts http/https with basic structure.
_URL_RE = re.compile(
    r"^https?://"
    r"[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?"
    r"(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*"
    r"(?::\d{1,5})?"
    r"(?:/[^\s]*)?$"
)

# ISO 8601 date: YYYY-MM-DD
_DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}$")

# ISO 8601 date-time – timezone is REQUIRED per AnyVali spec
_DATETIME_RE = re.compile(
    r"^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}"
    r"(?:\.\d+)?"
    r"(?:Z|[+-]\d{2}:\d{2})$"
)


def validate_format(fmt: str, value: str) -> bool:
    """Validate a string value against a named format.

    Returns True if valid, False otherwise.
    """
    validator = _VALIDATORS.get(fmt)
    if validator is None:
        return True  # Unknown formats pass (lenient)
    return validator(value)


def _validate_email(value: str) -> bool:
    return _EMAIL_RE.match(value) is not None


def _validate_url(value: str) -> bool:
    return _URL_RE.match(value) is not None


def _validate_uuid(value: str) -> bool:
    try:
        _uuid_mod.UUID(value)
        return True
    except (ValueError, AttributeError):
        return False


def _validate_ipv4(value: str) -> bool:
    try:
        addr = ipaddress.IPv4Address(value)
        return True
    except (ipaddress.AddressValueError, ValueError):
        return False


def _validate_ipv6(value: str) -> bool:
    try:
        addr = ipaddress.IPv6Address(value)
        return True
    except (ipaddress.AddressValueError, ValueError):
        return False


def _validate_date(value: str) -> bool:
    if not _DATE_RE.match(value):
        return False
    try:
        date.fromisoformat(value)
        return True
    except ValueError:
        return False


def _validate_datetime(value: str) -> bool:
    if not _DATETIME_RE.match(value):
        return False
    try:
        # Python 3.11+ handles Z suffix; for older versions, replace Z with +00:00
        normalized = value.replace("Z", "+00:00")
        datetime.fromisoformat(normalized)
        return True
    except ValueError:
        return False


_VALIDATORS: dict[str, type[None] | object] = {
    "email": _validate_email,
    "url": _validate_url,
    "uuid": _validate_uuid,
    "ipv4": _validate_ipv4,
    "ipv6": _validate_ipv6,
    "date": _validate_date,
    "date-time": _validate_datetime,
}
