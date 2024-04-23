# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import re


lineFormatRe = re.compile(r'^(\w+)\s+(\d+)$')

def check_format(file_path):
    with open(file_path, 'r') as file:
        for line_number, line in enumerate(file, start=1):
            line = line.strip()
            # Skip empty lines and comments
            if not line or line.startswith('#'):
                continue

            # Use regular expression to check for the required format
            match = re.match(lineFormatRe, line)
            if not match:
                print(f"Error in line {line_number}: {line} - Incorrect format")
                return False

    print("passed: {}".format(file_path))
    return True

assert check_format("./encoding.txt")
assert check_format("./constants.txt")

def check_typecodes_unique():
    with open("./encoding.txt", 'r') as file:
        used = {}
        for line_number, line in enumerate(file, start=1):
            line = line.strip()
            # Skip empty lines and comments
            if not line or line.startswith('#'):
                continue

            # extract name and code
            match = re.match(lineFormatRe, line)
            messageName = match.group(1)
            code = int(match.group(2))

            # check for duplicate codes
            if code in used:
                print(f"Error in line {line_number}: {line} - Duplicate code {code} for {messageName} and {used[code]}")
                return False
            used[code] = messageName
    print("passed: unique type codes")
    return True

assert check_typecodes_unique()

def read_all_reqresp_messages():
    pattern = re.compile(r"message\s+(.*?(Request|Response))[\s{]")
    messages = {}
    with open("./schema.proto") as f:
        lines = f.readlines()
        for line in lines:
            # print(f"line: {line}")
            match = pattern.match(line)
            if match:
                # print(f"match: {match.group(1)}")
                name = match.group(1)
                messages[name.lower()] = True
    return messages

# find each message name in encoding.txt
def check_typecodes_are_present():
    proto_messages = read_all_reqresp_messages()
    consts = open("./encoding.txt", 'r')
    for line_number, line in enumerate(consts, start=1):
        line = line.strip()
        # Skip empty lines and comments
        if not line or line.startswith('#'):
            continue

        # extract name and code
        match = re.match(lineFormatRe, line)
        messageName = match.group(1).lower()
        if messageName not in proto_messages:
            print(f"Error in line {line_number}: {line} - Unknown message name {messageName}")
            return False
        del proto_messages[messageName]

    if len(proto_messages) != 0:
        print(f"Error: messages with no encoding: {proto_messages}")
        return False

    print("passed: all request/response messages have a code")
    return True

assert check_typecodes_are_present()