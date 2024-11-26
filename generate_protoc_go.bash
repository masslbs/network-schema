#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

set -euo pipefail

# Get version and commit
SCHEMA_VERSION=`cat VERSION`

# protobuf file and encoding helpers from network schema
test -d go/pb && rm -r go/pb
mkdir go/pb
chmod u+w *.proto
for input in *.proto; do
  goFname="$(basename $input | sed 's/.proto/.pb.go/')"
  echo "Generating $goFname from $input"
  protoc \
    -I=. \
    --go_out=paths=source_relative:. \
    --go_opt="Msubscription.proto=github.com/masslbs/network-schema/pb;pb" \
    --go_opt="Mtransport.proto=github.com/masslbs/network-schema/pb;pb" \
    --go_opt="Mshop_requests.proto=github.com/masslbs/network-schema/pb;pb" \
    --go_opt="Mauthentication.proto=github.com/masslbs/network-schema/pb;pb" \
    --go_opt="Mbase_types.proto=github.com/masslbs/network-schema/pb;pb" \
    --go_opt="Merror.proto=github.com/masslbs/network-schema/pb;pb" \
    --go_opt="M$(basename $input)=github.com/masslbs/network-schema/pb;pb" \
    $input
  # Prepend comment with versioning info under SPDX header
  sed -i "5i // Generated from network-schema/$input at version v$SCHEMA_VERSION\n" $goFname
  mv $goFname go/pb/$goFname
done

# export interface
sed -i 's/isEnvelope_Message/IsEnvelope_Message/' go/pb/envelope.pb.go

# shorten error codes
sed -i 's/ErrorCodes_ERROR_CODES_/ErrorCodes_/' go/pb/error.pb.go

go fmt go/...
