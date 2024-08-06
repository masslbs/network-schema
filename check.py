# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import re
import os

lineFormatRe = re.compile(r"^(\w+)\s+(\d+)$")


def check_format(file_path):
    with open(file_path, "r") as file:
        for line_number, line in enumerate(file, start=1):
            line = line.strip()
            # Skip empty lines and comments
            if not line or line.startswith("#"):
                continue

            # Use regular expression to check for the required format
            match = re.match(lineFormatRe, line)
            if not match:
                print(f"Error in line {line_number}: {line} - Incorrect format")
                return False

    print("passed: {}".format(file_path))
    return True


assert check_format("./constants.txt")
