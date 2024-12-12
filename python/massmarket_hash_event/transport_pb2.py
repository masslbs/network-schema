# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: transport.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder

# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()
from massmarket_hash_event import base_types_pb2 as base__types__pb2
from google.protobuf import any_pb2 as google_dot_protobuf_dot_any__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(
    b'\n\x0ftransport.proto\x12\x0bmarket.mass\x1a\x10\x62\x61se_types.proto\x1a\x19google/protobuf/any.proto"]\n\x0bSignedEvent\x12#\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x14.google.protobuf.Any\x12)\n\tsignature\x18\x02 \x01(\x0b\x32\x16.market.mass.Signature"=\n\x11\x45ventWriteRequest\x12(\n\x06\x65vents\x18\x01 \x03(\x0b\x32\x18.market.mass.SignedEvent",\n\x11SyncStatusRequest\x12\x17\n\x0funpushed_events\x18\x01 \x01(\x04"\r\n\x0bPingRequestb\x06proto3'
)

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, "transport_pb2", _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
    DESCRIPTOR._options = None
    _globals["_SIGNEDEVENT"]._serialized_start = 77
    _globals["_SIGNEDEVENT"]._serialized_end = 170
    _globals["_EVENTWRITEREQUEST"]._serialized_start = 172
    _globals["_EVENTWRITEREQUEST"]._serialized_end = 233
    _globals["_SYNCSTATUSREQUEST"]._serialized_start = 235
    _globals["_SYNCSTATUSREQUEST"]._serialized_end = 279
    _globals["_PINGREQUEST"]._serialized_start = 281
    _globals["_PINGREQUEST"]._serialized_end = 294
# @@protoc_insertion_point(module_scope)
