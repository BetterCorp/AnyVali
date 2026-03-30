"""Core types for AnyVali Python SDK."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Generic, Literal, TypeVar

T = TypeVar("T")

ExportMode = Literal["portable", "extended"]
UnknownKeyMode = Literal["reject", "strip", "allow"]


@dataclass(frozen=True)
class ValidationIssue:
    """A single validation issue."""

    code: str
    message: str
    path: list[str | int] = field(default_factory=list)
    expected: Any = None
    received: Any = None
    meta: dict[str, Any] | None = None


@dataclass(frozen=True)
class ParseResult(Generic[T]):
    """Result of a safe_parse call."""

    success: bool
    data: T | None = None
    issues: list[ValidationIssue] = field(default_factory=list)


@dataclass(frozen=True)
class AnyValiDocument:
    """Top-level AnyVali interchange document."""

    anyvali_version: str = "1.0"
    schema_version: str = "1"
    root: dict[str, Any] = field(default_factory=dict)
    definitions: dict[str, Any] = field(default_factory=dict)
    extensions: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        d: dict[str, Any] = {
            "anyvaliVersion": self.anyvali_version,
            "schemaVersion": self.schema_version,
            "root": self.root,
        }
        if self.definitions:
            d["definitions"] = self.definitions
        if self.extensions:
            d["extensions"] = self.extensions
        return d

    @classmethod
    def from_dict(cls, d: dict[str, Any]) -> AnyValiDocument:
        return cls(
            anyvali_version=d.get("anyvaliVersion", "1.0"),
            schema_version=d.get("schemaVersion", "1"),
            root=d.get("root", {}),
            definitions=d.get("definitions", {}),
            extensions=d.get("extensions", {}),
        )


class ValidationError(Exception):
    """Raised by parse() when validation fails."""

    def __init__(self, issues: list[ValidationIssue]) -> None:
        self.issues = issues
        messages = "; ".join(i.message for i in issues)
        super().__init__(f"Validation failed: {messages}")
