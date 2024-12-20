# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: subscription.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder

# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()
from massmarket_hash_event import base_types_pb2 as base__types__pb2
from massmarket_hash_event import transport_pb2 as transport__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(
    b'\n\x12subscription.proto\x12\x0bmarket.mass\x1a\x10\x62\x61se_types.proto\x1a\x0ftransport.proto"\x86\x02\n\x13SubscriptionRequest\x12\x19\n\x11start_shop_seq_no\x18\x01 \x01(\x04\x12%\n\x07shop_id\x18\x02 \x01(\x0b\x32\x14.market.mass.Uint256\x12\x38\n\x07\x66ilters\x18\x03 \x03(\x0b\x32\'.market.mass.SubscriptionRequest.Filter\x1as\n\x06\x46ilter\x12,\n\x0bobject_type\x18\x03 \x01(\x0e\x32\x17.market.mass.ObjectType\x12-\n\tobject_id\x18\x04 \x01(\x0b\x32\x15.market.mass.ObjectIdH\x00\x88\x01\x01\x42\x0c\n\n_object_id"\xc2\x01\n\x17SubscriptionPushRequest\x12\x17\n\x0fsubscription_id\x18\x01 \x01(\x0c\x12\x43\n\x06\x65vents\x18\x02 \x03(\x0b\x32\x33.market.mass.SubscriptionPushRequest.SequencedEvent\x1aI\n\x0eSequencedEvent\x12\'\n\x05\x65vent\x18\x01 \x01(\x0b\x32\x18.market.mass.SignedEvent\x12\x0e\n\x06seq_no\x18\x02 \x01(\x04"4\n\x19SubscriptionCancelRequest\x12\x17\n\x0fsubscription_id\x18\x01 \x01(\x0c*\xbc\x01\n\nObjectType\x12\x1b\n\x17OBJECT_TYPE_UNSPECIFIED\x10\x00\x12\x17\n\x13OBJECT_TYPE_LISTING\x10\x01\x12\x13\n\x0fOBJECT_TYPE_TAG\x10\x02\x12\x15\n\x11OBJECT_TYPE_ORDER\x10\x03\x12\x17\n\x13OBJECT_TYPE_ACCOUNT\x10\x04\x12\x18\n\x14OBJECT_TYPE_MANIFEST\x10\x05\x12\x19\n\x15OBJECT_TYPE_INVENTORY\x10\x06\x62\x06proto3'
)

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, "subscription_pb2", _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
    DESCRIPTOR._options = None
    _globals["_OBJECTTYPE"]._serialized_start = 587
    _globals["_OBJECTTYPE"]._serialized_end = 775
    _globals["_SUBSCRIPTIONREQUEST"]._serialized_start = 71
    _globals["_SUBSCRIPTIONREQUEST"]._serialized_end = 333
    _globals["_SUBSCRIPTIONREQUEST_FILTER"]._serialized_start = 218
    _globals["_SUBSCRIPTIONREQUEST_FILTER"]._serialized_end = 333
    _globals["_SUBSCRIPTIONPUSHREQUEST"]._serialized_start = 336
    _globals["_SUBSCRIPTIONPUSHREQUEST"]._serialized_end = 530
    _globals["_SUBSCRIPTIONPUSHREQUEST_SEQUENCEDEVENT"]._serialized_start = 457
    _globals["_SUBSCRIPTIONPUSHREQUEST_SEQUENCEDEVENT"]._serialized_end = 530
    _globals["_SUBSCRIPTIONCANCELREQUEST"]._serialized_start = 532
    _globals["_SUBSCRIPTIONCANCELREQUEST"]._serialized_end = 584
# @@protoc_insertion_point(module_scope)
