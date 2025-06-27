# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import os
from typing import Optional


def get_vectors_dir() -> str:
    """
    Get the test vectors directory path.

    Can be overridden with the MASS_TEST_VECTORS_DIR environment variable.
    Default is the vectors directory relative to the test files.

    Returns:
        str: Absolute path to the vectors directory
    """
    # Check for environment variable override
    env_path = os.environ.get("MASS_TEST_VECTORS_DIR")
    if env_path:
        return os.path.abspath(env_path)

    # Default: vectors directory relative to this file
    # This gives us network-schema/vectors/
    default_path = os.path.join(os.path.dirname(__file__), "..", "..", "vectors")
    return os.path.abspath(default_path)


def get_vector_file(filename: str) -> str:
    """
    Get the full path to a specific test vector file.

    Args:
        filename: Name of the vector file (e.g., "ManifestOkay.cbor")

    Returns:
        str: Absolute path to the vector file
    """
    return os.path.join(get_vectors_dir(), filename)


def list_vector_files(pattern: Optional[str] = None) -> list[str]:
    """
    List all vector files in the vectors directory.

    Args:
        pattern: Optional pattern to filter files (e.g., "Okay.cbor")

    Returns:
        list[str]: List of filenames matching the pattern
    """
    vectors_dir = get_vectors_dir()
    if not os.path.exists(vectors_dir):
        return []

    files = os.listdir(vectors_dir)

    if pattern:
        files = [f for f in files if pattern in f]

    return files
