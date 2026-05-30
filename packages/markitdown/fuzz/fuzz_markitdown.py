#!/usr/bin/env python3
"""Atheris fuzz targets for the markitdown package.

Run with:
    python fuzz_markitdown.py          # runs until interrupted
    python fuzz_markitdown.py -atheris_runs=50000   # bounded run for CI

Requires: pip install atheris markitdown
"""
import io
import sys

import atheris

with atheris.instrument_imports():
    from markitdown import MarkItDown
    from markitdown._uri_utils import parse_data_uri

_md = MarkItDown()


def TestOneInput(data: bytes) -> None:
    fdp = atheris.FuzzedDataProvider(data)

    # Pick a file extension to drive converter selection.
    extensions = [".html", ".txt", ".csv", ".xml", ".json", ".md", ".rst"]
    ext = fdp.PickValueInList(extensions)

    content = fdp.ConsumeBytes(fdp.ConsumeIntInRange(0, len(data)))
    try:
        _md.convert_stream(io.BytesIO(content), file_extension=ext)
    except Exception:
        # Conversion errors (unsupported format, parse failure, etc.) are
        # expected and not a bug — only crashes and hangs matter here.
        pass

    # Fuzz the data-URI parser independently; it handles untrusted input.
    uri_bytes = fdp.ConsumeBytes(fdp.ConsumeIntInRange(0, max(0, len(data) - 1)))
    try:
        parse_data_uri(uri_bytes.decode("latin-1"))
    except Exception:
        pass


def main() -> None:
    atheris.Setup(sys.argv, TestOneInput)
    atheris.Fuzz()


if __name__ == "__main__":
    main()
