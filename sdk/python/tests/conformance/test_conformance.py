"""Conformance tests driven by the shared AnyVali test corpus.

Reads JSON test files from spec/corpus/, imports the schema via the
AnyVali importer, runs safe_parse, and compares results against
expected outcomes.
"""

from __future__ import annotations

from typing import Any

import pytest

import anyvali as v
from .runner import corpus_ids, load_corpus

_CORPUS = load_corpus()


@pytest.mark.skipif(not _CORPUS, reason="No conformance corpus files found")
@pytest.mark.parametrize("case", _CORPUS, ids=corpus_ids(_CORPUS))
def test_conformance(case: dict[str, Any]) -> None:
    """Run a single conformance test case."""
    schema_doc = case["schema"]
    input_value = case["input"]
    expected_valid = case["valid"]
    expected_output = case.get("output")
    expected_issues = case.get("issues", [])

    # Import the schema from the document
    schema = v.import_schema(schema_doc)

    # Run safe_parse
    result = schema.safe_parse(input_value)

    # Check success/failure
    assert result.success == expected_valid, (
        f"Expected valid={expected_valid}, got success={result.success}. "
        f"Issues: {[{'code': i.code, 'message': i.message, 'path': i.path} for i in result.issues]}"
    )

    # If expected success, check output data
    if expected_valid and expected_output is not None:
        assert result.data == expected_output, (
            f"Expected output={expected_output!r}, got data={result.data!r}"
        )

    # If expected failure, check issue codes
    if not expected_valid and expected_issues:
        actual_codes = [issue.code for issue in result.issues]
        for exp_issue in expected_issues:
            exp_code = exp_issue["code"]
            assert exp_code in actual_codes, (
                f"Expected issue code '{exp_code}' not found. "
                f"Actual codes: {actual_codes}"
            )

            # Check paths if specified
            if "path" in exp_issue:
                matching = [
                    i for i in result.issues
                    if i.code == exp_code and i.path == exp_issue["path"]
                ]
                assert matching, (
                    f"Expected issue with code={exp_code} "
                    f"at path={exp_issue['path']} not found. "
                    f"Actual issues: {[(i.code, i.path) for i in result.issues]}"
                )
