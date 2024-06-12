# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import json

from massmarket_hash_event import shop_events_pb2

# helpers

TYPE_DOUBLE = 1
TYPE_FLOAT = 2
TYPE_INT64 = 3
TYPE_UINT64 = 4
TYPE_INT32 = 5
TYPE_FIXED64 = 6
TYPE_FIXED32 = 7
TYPE_BOOL = 8
TYPE_STRING = 9
TYPE_GROUP = 10
TYPE_MESSAGE = 11
TYPE_BYTES = 12
TYPE_UINT32 = 13
TYPE_ENUM = 14
TYPE_SFIXED32 = 15
TYPE_SFIXED64 = 16
TYPE_SINT32 = 17
TYPE_SINT64 = 18

name_to_number = {
    "TYPE_DOUBLE": TYPE_DOUBLE,
    "TYPE_FLOAT": TYPE_FLOAT,
    "TYPE_INT64": TYPE_INT64,
    "TYPE_UINT64": TYPE_UINT64,
    "TYPE_INT32": TYPE_INT32,
    "TYPE_FIXED64": TYPE_FIXED64,
    "TYPE_FIXED32": TYPE_FIXED32,
    "TYPE_BOOL": TYPE_BOOL,
    "TYPE_STRING": TYPE_STRING,
    "TYPE_GROUP": TYPE_GROUP,
    "TYPE_MESSAGE": TYPE_MESSAGE,
    "TYPE_BYTES": TYPE_BYTES,
    "TYPE_UINT32": TYPE_UINT32,
    "TYPE_ENUM": TYPE_ENUM,
    "TYPE_SFIXED32": TYPE_SFIXED32,
    "TYPE_SFIXED64": TYPE_SFIXED64,
    "TYPE_SINT32": TYPE_SINT32,
    "TYPE_SINT64": TYPE_SINT64
}

number_to_name = {v: k for k, v in name_to_number.items()}

# https://eips.ethereum.org/EIPS/eip-712#definition-of-typed-structured-data-%F0%9D%95%8A
# Definition: The atomic types are bytes1 to bytes32, uint8 to uint256, int8 to int256, bool and address.
# These correspond to their definition in Solidity. Note that there are no aliases uint and int.
# Note that contract addresses are always plain address. Fixed point numbers are not supported by the standard.
# Future versions of this standard may add new atomic types.
protobuf_to_typedData = {
    "TYPE_BOOL": "bool",
    "TYPE_STRING": "string",
    "TYPE_BYTES": "string",
    # numbers
    "TYPE_INT64": "int64",
    "TYPE_UINT64": "uint64",
    "TYPE_INT32": "int32",
    "TYPE_FIXED64": "uint64",
    "TYPE_FIXED32": "uint32",
    "TYPE_UINT32": "uint32",
    "TYPE_SFIXED32": "int32",
    "TYPE_SFIXED64": "int64",
    "TYPE_SINT32": "int32",
    "TYPE_SINT64": "int64",
    # Unsupported
    "TYPE_DOUBLE": None,
    "TYPE_FLOAT": None,
    "TYPE_GROUP": None,  # There is no direct match in Solidity as it seems to be specific to Protobuf.
    # Need to define the length of the bytes array manually
    "TYPE_ENUM": "int256"  # Enums in solidity are represented as integers.
}

# if we get None, we need another hereuistic, like names that end with _id are bytes32

def typed_data_definition(message):
    fields = []
    print(f"\nmessage: {message.DESCRIPTOR.name}")
    for field in message.DESCRIPTOR.fields:
        if field.message_type:  # field has a message type, it's a nested message.
            #raise Exception(f"Fix generic nested messages {field.name}")
             value = getattr(message, field.name)
             if field.label == field.LABEL_REPEATED:  # repeated field contains multiple instances of the message
                 raise Exception(f"repeated fields on {field.name} not supported")
                 #for item in value:
                 #    fields.extend(typed_data_definition(item))
             else:  # only a single nested message.
                 fields.append({"name": field.name, "message": typed_data_definition(value)})

        else:  # field is not a message (i.e., simple type field like int32, string, etc).
            tipe = number_to_name[field.type]

            td = protobuf_to_typedData[tipe]
            # overrides
            if field.name.endswith("_id") or field.name.endswith("hash"):
                td = "bytes32"
            elif field.name.endswith("addr"):
                td = "address"
            elif field.name.endswith("_ids"):
                td = "bytes32[]"
            elif field.name.endswith("_key"):
                td = "bytes" # TODO: compressed public key
            elif field.name.endswith("_signature"):
                td = "bytes"

            # hard overwrites
            if field.name == "metadata":
                td = "bytes"
            elif field.name == "diffs":
                td = "int32[]"

            if td is None:
                raise Exception(f"Unknown type: {field.name} {tipe}")
            print(f"{field.name}: {tipe} <> {td}")
            fields.append({'name':field.name, 'type': td})
    return fields

# get all the union field in the Event message
evt = shop_events_pb2.ShopEvent()
union = evt.DESCRIPTOR.oneofs[0]

from google.protobuf import message_factory

messages = {}
# iterate over all the union fields and get the message class for each type
for type in [f.message_type for f in union.fields]:
    proto_class = message_factory.GetMessageClass(type)
    msg = proto_class()
    fields = typed_data_definition(msg)
    messages[msg.DESCRIPTOR.name] = fields

with open(f"typedData.json", "w") as f:
  f.write(json.dumps(messages, indent=2))
