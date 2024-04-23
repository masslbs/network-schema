# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

__all__ = ['hash_event', 'Hasher', 'schema_pb2']

import json
import binascii
from importlib.resources import files

from eth_account.messages import encode_structured_data

from massmarket_hash_event import schema_pb2

event_typedData_spec = json.loads(files("massmarket_hash_event").joinpath("typedData.json").read_text())

class Hasher:
    def __init__(self, chain_id: int, storeRegAddress: str):
        self.chain_id = chain_id
        if not storeRegAddress.startswith("0x"):
            raise Exception(f"Invalid contract address: {storeRegAddress}")
        data = binascii.unhexlify(storeRegAddress[2:])
        if len(data) != 20:
            raise Exception(f"Invalid contract address: {storeRegAddress}")
        self.storeRegAddress = storeRegAddress
    def hash_event(self, evt: schema_pb2.Event):
        return hash_event(evt, self.chain_id, self.storeRegAddress)

def hash_event(evt: schema_pb2.Event, chain_id, storeRegAddress):
    event_type = getattr(evt, evt.WhichOneof('union'))
    event_name = event_type.DESCRIPTOR.name
    event_td_spec = event_typedData_spec[event_name].copy()
    # remove fields from event_td_spec that are not set in event_td_data
    if event_name == "UpdateManifest":
        field = event_type.field
        if field == schema_pb2.UpdateManifest.MANIFEST_FIELD_DOMAIN:
            event_td_spec.remove({"name": "tag_id", "type": "bytes32"})
            event_td_spec.remove({"name": "erc20_addr", "type": "address"})
        elif field == schema_pb2.UpdateManifest.MANIFEST_FIELD_PUBLISHED_TAG:
            event_td_spec.remove({"name": "string", "type": "string"})
            event_td_spec.remove({"name": "erc20_addr", "type": "address"})
        elif field == schema_pb2.UpdateManifest.MANIFEST_FIELD_ADD_ERC20:
            event_td_spec.remove({"name": "tag_id", "type": "bytes32"})
            event_td_spec.remove({"name": "string", "type": "string"})
        elif field == schema_pb2.UpdateManifest.MANIFEST_FIELD_REMOVE_ERC20:
            event_td_spec.remove({"name": "tag_id", "type": "bytes32"})
            event_td_spec.remove({"name": "string", "type": "string"})
        else:
            raise Exception(f"Unknown field: {field}")
    elif event_name == "UpdateItem":
        field = event_type.field
        if field == schema_pb2.UpdateItem.ITEM_FIELD_PRICE:
            event_td_spec.remove({"name": "metadata", "type": "bytes"})
        elif field == schema_pb2.UpdateItem.ITEM_FIELD_METADATA:
            event_td_spec.remove({"name": "price", "type": "string"})
        else:
            raise Exception(f"Unknown field: {field}")
    # convert event to typedData
    event_td_data = event_to_typedData_dict(evt, event_td_spec)
    if event_name == "ChangeStock":
        # remove cart_id if empty
        if len(event_td_data["cart_id"]) == 0:
            event_td_spec.remove({"name": "cart_id", "type": "bytes32"})
            event_td_spec.remove({"name": "tx_hash", "type": "bytes32"})
            del event_td_data["cart_id"]
            del event_td_data["tx_hash"]
    elif event_name == "CartFinalized":
        # remove erc20_addr if empty
        if len(event_td_data["erc20_addr"]) == 0:
            event_td_spec.remove({"name": "erc20_addr", "type": "address"})
            del event_td_data["erc20_addr"]

    typed_data = {
        "types": {
            "EIP712Domain": [
                {"name": "name", "type": "string"},
                {"name": "version", "type": "string"},
                {"name": "chainId", "type": "uint256"},
                {"name": "verifyingContract", "type": "address"},
            ],
            event_name: event_td_spec,
        },
        "primaryType": event_name,
        "domain": {
            "name": "MassMarket",
            "version": "1",
            "chainId": chain_id,
            "verifyingContract": storeRegAddress,
        },
        "message": event_td_data
    }
    return encode_structured_data(typed_data)

def event_to_typedData_dict(evt: schema_pb2.Event, spec_for_event):
    which = evt.WhichOneof("union")
    if which is None:
        raise Exception("No event type set")
    unwrapped = getattr(evt, which)
    field_names = [field["name"] for field in spec_for_event]
    event_td_data = {}
    for field in field_names:
        event_td_data[field] = getattr(unwrapped, field)
    return event_td_data
