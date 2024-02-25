#!/usr/bin/env python

import requests

def create_group_chat(executor_id, group_chat_name, write_api_server_base_url):
  headers = {
    "Content-Type": "application/json"
  }

  data = {
    "query": "mutation CreateGroupChat($input: CreateGroupChatInput!) { createGroupChat(input: $input) { groupChatId } }",
    "variables": {
      "input": {
        "name": group_chat_name,
        "executorId": executor_id
      }
    }
  }

  response = requests.post(f"{write_api_server_base_url}/query", headers=headers, json=data)
  response_json = response.json()

  group_chat_id = response_json.get('data', {}).get('createGroupChat', {}).get('groupChatId')

  return group_chat_id

def add_member_to_group_chat(group_chat_id, user_account_id, role, executor_id, write_api_server_base_url):
  headers = {
    "Content-Type": "application/json"
  }

  data = {
    "query": "mutation AddMember($input: AddMemberInput!) { addMember(input: $input) { groupChatId } }",
    "variables": {
      "input": {
        "groupChatId": group_chat_id,
        "userAccountId": user_account_id,
        "role": role,
        "executorId": executor_id
      }
    }
  }

  response = requests.post(f"{write_api_server_base_url}/query", headers=headers, json=data)
  response_json = response.json()

  result = response_json.get('data', {}).get('addMember', {}).get('groupChatId')

  return result

write_api_server_base_url = "http://localhost:28080"  # 実際のAPIサーバーのベースURLに置き換えてください
executor_id = "UserAccount-01H42K4ABWQ5V2XQEP3A48VE0Z"
group_chat_name = "group-chat-example"

group_chat_id = create_group_chat(executor_id, group_chat_name, write_api_server_base_url)
print(group_chat_id)
