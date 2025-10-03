from openai import OpenAI
import json
import config


api_key = config.ACCOUNT_TOKEN

client = OpenAI(api_key=api_key, base_url="https://api.ai.sakura.ad.jp/v1")

model = "gpt-oss-120b"
messages = [{"role": "user", "content": "今日の東京の天気とUSD/JPYのレートを教えて"}]

tools = [
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "指定した都市の天気を取得する",
            "parameters": {
                "type": "object",
                "properties": {"city": {"type": "string"}},
                "required": ["city"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "get_exchange_rate",
            "description": "通貨ペアの現在の為替レートを取得する",
            "parameters": {
                "type": "object",
                "properties": {"from": {"type": "string"}, "to": {"type": "string"}},
                "required": ["from", "to"],
            },
        },
    },
]
completion = client.chat.completions.create(
    model=model,
    messages=messages,
    tools=tools,
    tool_choice="auto",
)

print(completion.model_dump_json())

choice = completion.choices[0]
messages.append(choice.message)
if choice.finish_reason == "tool_calls":
    for tool in choice.message.tool_calls:
        args = json.loads(tool.function.arguments)
        if tool.function.name == "get_weather":
            messages.append(
                {
                    "role": "tool",
                    "tool_call_id": tool.id,
                    "content": "%sの天気は晴れです" % (args["city"]),
                }
            )
        if tool.function.name == "get_exchange_rate":
            messages.append(
                {
                    "role": "tool",
                    "tool_call_id": tool.id,
                    "content": "%sから%sに変換すると150円です。ｸﾞﾌﾌ"
                    % (args["from"], args["to"]),
                }
            )

completion = client.chat.completions.create(
    model=model,
    messages=messages,
    tools=tools,
    tool_choice="auto",
)

print(completion.model_dump_json())

choice = completion.choices[0]
messages.append(choice.message)
if choice.finish_reason == "tool_calls":
    for tool in choice.message.tool_calls:
        args = json.loads(tool.function.arguments)
        if tool.function.name == "get_weather":
            messages.append(
                {
                    "role": "tool",
                    "tool_call_id": tool.id,
                    "content": "%sの天気は晴れです" % (args["city"]),
                }
            )
        if tool.function.name == "get_exchange_rate":
            messages.append(
                {
                    "role": "tool",
                    "tool_call_id": tool.id,
                    "content": "%sから%sに変換すると150円です。ｸﾞﾌﾌ"
                    % (args["from"], args["to"]),
                }
            )

completion = client.chat.completions.create(
    model=model,
    messages=messages,
    tools=tools,
    tool_choice="auto",
)

print(completion.model_dump_json())
