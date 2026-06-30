import json
import locale
import subprocess
from typing import Any, BinaryIO, Union

_verified_paths: set[str] = set()


def exiftool_metadata(
    file_stream: BinaryIO,
    *,
    exiftool_path: Union[str, None],
) -> Any:  # Need a better type for json data
    # Nothing to do
    if not exiftool_path:
        return {}

    # Verify exiftool version (once per path)
    if exiftool_path not in _verified_paths:
        try:
            version_output = subprocess.run(
                [exiftool_path, "-ver"],
                capture_output=True,
                text=True,
                check=True,
            ).stdout.strip()
            version = tuple(map(int, version_output.split(".")))
            min_version = (12, 24)
            if version < min_version:
                raise RuntimeError(
                    f"ExifTool version {version_output} is vulnerable to CVE-2021-22204. "
                    "Please upgrade to version 12.24 or later."
                )
        except (subprocess.CalledProcessError, ValueError) as e:
            raise RuntimeError("Failed to verify ExifTool version.") from e
        _verified_paths.add(exiftool_path)

    # Run exiftool
    cur_pos = file_stream.tell()
    try:
        output = subprocess.run(
            [exiftool_path, "-json", "-"],
            input=file_stream.read(),
            capture_output=True,
            text=False,
        ).stdout

        return json.loads(
            output.decode(locale.getpreferredencoding(False)),
        )[0]
    finally:
        file_stream.seek(cur_pos)
