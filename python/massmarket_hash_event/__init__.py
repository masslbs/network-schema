# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

__all__ = ['hash_event', 'Hasher',
           'transport_pb2', 'authentication_pb2', 'store_requests_pb2', 'error_pb2',
           'store_events_pb2']

import json
import binascii
from pprint import pprint
from importlib.resources import files

from eth_account.messages import encode_structured_data

from massmarket_hash_event import store_events_pb2

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
    def hash_event(self, evt: store_events_pb2.StoreEvent):
        return hash_event(evt, self.chain_id, self.storeRegAddress)

def hash_event(evt: store_events_pb2.StoreEvent, chain_id, storeRegAddress):
    event_type = getattr(evt, evt.WhichOneof('union'))
    event_name = event_type.DESCRIPTOR.name
    event_td_spec = event_typedData_spec[event_name].copy()

    types = {
        "EIP712Domain": [
            {"name": "name", "type": "string"},
            {"name": "version", "type": "string"},
            {"name": "chainId", "type": "uint256"},
            {"name": "verifyingContract", "type": "address"},
        ],
    }

    # step 1: remove fields from event_td_spec that are not set in event_td_data
    if event_name == "UpdateStoreManifest":
        field = event_type.field
        if field == store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_DOMAIN:
            event_td_spec.remove({"name": "tag_id", "type": "bytes32"})
            event_td_spec.remove({"name": "erc20_addr", "type": "address"})
        elif field == store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_PUBLISHED_TAG:
            event_td_spec.remove({"name": "string", "type": "string"})
            event_td_spec.remove({"name": "erc20_addr", "type": "address"})
        elif field == store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_ADD_ERC20:
            event_td_spec.remove({"name": "tag_id", "type": "bytes32"})
            event_td_spec.remove({"name": "string", "type": "string"})
        elif field == store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_REMOVE_ERC20:
            event_td_spec.remove({"name": "tag_id", "type": "bytes32"})
            event_td_spec.remove({"name": "string", "type": "string"})
        else:
            raise Exception(f"Unknown field: {field}")
    elif event_name == "UpdateItem":
        field = event_type.field
        if field == store_events_pb2.UpdateItem.ITEM_FIELD_PRICE:
            event_td_spec.remove({"name": "metadata", "type": "bytes"})
        elif field == store_events_pb2.UpdateItem.ITEM_FIELD_METADATA:
            event_td_spec.remove({"name": "price", "type": "string"})
        else:
            raise Exception(f"Unknown field: {field}")
    elif event_name == "UpdateTag":
        action = event_type.action
        if action == store_events_pb2.UpdateTag.TAG_ACTION_ADD_ITEM:
            event_td_spec.remove({"name": "new_name", "type": "string"})
            event_td_spec.remove({"name": "delete", "type": "bool"})
        elif action == store_events_pb2.UpdateTag.TAG_ACTION_REMOVE_ITEM:
            event_td_spec.remove({"name": "new_name", "type": "string"})
            event_td_spec.remove({"name": "delete", "type": "bool"})
        elif action == store_events_pb2.UpdateTag.TAG_ACTION_RENAME:
            event_td_spec.remove({"name": "item_id", "type": "bytes32"})
            event_td_spec.remove({"name": "delete", "type": "bool"})
        elif action == store_events_pb2.UpdateTag.TAG_ACTION_DELETE_TAG:
            event_td_spec.remove({"name": "item_id", "type": "bytes32"})
            event_td_spec.remove({"name": "new_name", "type": "string"})
        else:
            raise Exception(f"Unknown action on UpdateTag: {action}")
    elif event_name == "UpdateOrder":
        # UpdateOrder follows a nested message pattern and thus needs slightly differnt handling
        # we are not just pruning the unused values but add the used type to typed_data
        action = event_type.WhichOneof("action")

        # build map of action > message
        name2msgtd = {}
        msgs = [td for td in event_td_spec if "message" in td]
        for td in msgs:
            name2msgtd[td["name"]] = td["message"]

        # other fields are used as is
        event_td_spec = [td for td in event_td_spec if "message" not in td]

        if action not in name2msgtd:
         raise Exception(f"Unhandled action on UpdateOrder: {action}")

        event_td_spec.append({"name": action, "type": action})
        types[action] = name2msgtd[action].copy()


    # step 2: convert event to typedData
    event_td_data = event_to_typedData_dict(evt, event_td_spec)
    if event_name == "ChangeStock":
        # remove order_id if empty
        if len(event_td_data["order_id"]) == 0:
            event_td_spec.remove({"name": "order_id", "type": "bytes32"})
            event_td_spec.remove({"name": "tx_hash", "type": "bytes32"})
            del event_td_data["order_id"]
            del event_td_data["tx_hash"]
    elif event_name == "UpdateOrder":
        # we need to change the nested pb message into a subscriptable thing
        # for encode_structured_data to be happy
        action = event_type.WhichOneof("action")
        action_msg = getattr(event_type, action)
        action_obj = {}

        action_fields = action_msg.ListFields()
        for field in action_fields:
            name = field[0].name
            val = getattr(action_msg, name)
            action_obj[name] = val

        # Bools that are false are not set on a pb message
        if "is_payment_endpoint" not in action_fields:
            action_obj["is_payment_endpoint"] = False

        # remove currency_addr if empty
        if action == "items_finalized" and "currency_addr" not in action_obj:
            types[action].remove({"name": "currency_addr", "type": "address"})

        # replace nested protobuf thing
        event_td_data[action] = action_obj

    types[event_name] = event_td_spec
    typed_data = {
        "types": types,
        "primaryType": event_name,
        "domain": {
            "name": "MassMarket",
            "version": "1",
            "chainId": chain_id,
            "verifyingContract": storeRegAddress,
        },
        "message": event_td_data
    }
    pprint(typed_data)
    return encode_structured_data(typed_data)

def event_to_typedData_dict(evt: store_events_pb2.StoreEvent, spec_for_event):
    which = evt.WhichOneof("union")
    if which is None:
        raise Exception("No event type set")
    unwrapped = getattr(evt, which)
    field_names = [field["name"] for field in spec_for_event]
    event_td_data = {}
    for field in field_names:
        event_td_data[field] = getattr(unwrapped, field)
    return event_td_data
