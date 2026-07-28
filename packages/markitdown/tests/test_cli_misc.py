#!/usr/bin/env python3 -m pytest
import os
import subprocess
from markitdown import __version__

# This file contains CLI tests that are not directly tested by the FileTestVectors.
# This includes things like help messages, version numbers, and invalid flags.


def test_version() -> None:
    result = subprocess.run(
        ["python", "-m", "markitdown", "--version"], capture_output=True, text=True
    )

    assert result.returncode == 0, f"CLI exited with error: {result.stderr}"
    assert __version__ in result.stdout, f"Version not found in output: {result.stdout}"


def test_invalid_flag() -> None:
    result = subprocess.run(
        ["python", "-m", "markitdown", "--foobar"], capture_output=True, text=True
    )

    assert result.returncode != 0, f"CLI exited with error: {result.stderr}"
    assert (
        "unrecognized arguments" in result.stderr
    ), "Expected 'unrecognized arguments' to appear in STDERR"
    assert "SYNTAX" in result.stderr, "Expected 'SYNTAX' to appear in STDERR"


def test_llm_model_missing_api_key() -> None:
    env = {k: v for k, v in os.environ.items() if k != "OPENAI_API_KEY"}
    test_file = os.path.join(os.path.dirname(__file__), "test_files", "test.pdf")
    result = subprocess.run(
        ["python", "-m", "markitdown", "--llm-model", "gpt-4o", test_file],
        capture_output=True,
        text=True,
        env=env,
    )

    assert result.returncode != 0, f"CLI unexpectedly succeeded: {result.stdout}"
    assert (
        "Traceback" not in result.stderr
    ), f"Expected a clean error message, got a traceback: {result.stderr}"
    assert (
        "Failed to create LLM client" in result.stderr
    ), f"Expected a clean error message in STDERR: {result.stderr}"


if __name__ == "__main__":
    """Runs this file's tests from the command line."""
    test_version()
    test_invalid_flag()
    test_llm_model_missing_api_key()
    print("All tests passed!")
