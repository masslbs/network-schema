# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from io import BytesIO

import cbor2


def mass_types(encoder, obj):
    mapped = obj
    if hasattr(obj, "to_cbor_dict"):
        mapped = obj.to_cbor_dict()
    return encoder.encode(mapped)


# construct default encoder
def cbor_encode(obj):
    with BytesIO() as fp:
        cbor2.CBOREncoder(
            fp,
            canonical=True,
            date_as_datetime=True,
            default=mass_types,
        ).encode(obj)
        return fp.getvalue()
