# Jabber Bot n8n Node Specification

## Overview

Custom n8n node package for integrating with Jabber Bot XMPP messenger. Provides actions for sending messages, chat states, and files via XMPP, plus a webhook trigger for receiving incoming messages.

## Compatibility

- **n8n version**: 2.4.x
- **Node.js**: 18+
- **TypeScript**: 5.x

## Credentials

### JabberBotApi

| Field     | Type   | Required | Description                                              |
|-----------|--------|----------|----------------------------------------------------------|
| `baseUrl` | string | Yes      | API base URL (e.g., `http://192.168.1.100:8080/api/v1/`) |
| `apiKey`  | string | Yes      | API Key for authentication                               |

**Authentication**: Header `API-Key: {apiKey}`

## Nodes

### 1. Jabber Bot (Action Node)

Send messages, chat states, and files via XMPP.

#### Operations

| Operation       | Endpoint         | Description                             |
|-----------------|------------------|-----------------------------------------|
| Send Message    | POST /send       | Send XMPP message to user               |
| Send Chat State | POST /chat-state | Send chat state notification (XEP-0085) |
| Send File       | POST /send-file  | Upload and send file via XMPP           |

#### Send Message Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `to` | string | Yes | JID of recipient (user@domain.com) |
| `body` | string | Yes | Message content (max 10,000 chars) |
| `type` | options | No | chat, groupchat, headline, normal (default: chat) |

#### Send Chat State Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `to` | string | Yes | JID of recipient |
| `state` | options | Yes | active, composing, paused, inactive, gone |

#### Send File Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `to` | string | Yes | JID of recipient |
| `file` | binary | Yes | File to upload |
| `description` | string | No | File description |

#### Output Fields

All operations return:
- `success` (boolean)
- `message` (string)
- `data` (object) with operation-specific fields

### 2. Jabber Bot Trigger (Webhook Trigger)

Trigger workflow when Jabber Bot receives a message.

#### Configuration

1. Add trigger node to workflow
2. Activate workflow to get webhook URL
3. Configure Jabber Bot `config.yaml`:
   ```yaml
   webhook:
     url: "https://your-n8n.com/webhook/unique-id"
   ```

#### Webhook Payload

```json
{
  "id": "msg123",
  "from": "sender@example.com",
  "to": "bot@example.com",
  "body": "Hello bot",
  "type": "chat",
  "subject": "",
  "thread": "",
  "stamp": "2023-12-01T12:00:00Z",
  "timestamp": "2023-12-01T12:00:00Z",
  "source": "jabber-bot"
}
```

#### Output Fields

| Field | Description |
|-------|-------------|
| `id` | Message ID |
| `from` | Sender JID |
| `to` | Recipient JID |
| `body` | Message content |
| `type` | Message type |
| `subject` | Message subject |
| `thread` | Thread ID |
| `stamp` | Message timestamp |
| `timestamp` | Webhook timestamp |
| `source` | Always "jabber-bot" |

## Installation

### Development

```bash
# Clone and build
cd n8n-nodes-jabber-bot
npm install
npm run build

# Copy to n8n nodes directory
cp -r dist/* ~/.n8n/nodes/
```

### Production

```bash
npm install n8n-nodes-jabber-bot
```

## Tool Usage (AI Agent)

The Jabber Bot node is marked with `usableAsTool: true`, allowing it to be connected as a tool to the AI Agent node in n8n workflows.

### How It Works

When the AI Agent needs to send an XMPP message, it calls the Jabber Bot tool with the following input parameters:

| Parameter | Type | Description |
|-----------|------|-------------|
| `operation` | string | `sendMessage`, `sendChatState`, or `sendFile` |
| `to` | string | Recipient JID (e.g., `user@example.com`) |
| `body` | string | Message content (for `sendMessage`) |
| `type` | string | Message type: `chat`, `groupchat`, `headline`, `normal` |
| `state` | string | Chat state (for `sendChatState`): `active`, `composing`, `paused`, `inactive`, `gone` |

### Setup

1. Add an **AI Agent** node to your workflow
2. Connect a **Chat Model** (e.g., OpenAI, Anthropic) to the Agent
3. Drag the **Jabber Bot** node onto the canvas
4. Connect the Jabber Bot node's output to the AI Agent's **Tool** input
5. Configure Jabber Bot credentials (Base URL + API Key)

### Example Workflow

```
[Chat Trigger] → [AI Agent] → [Respond to Webhook]
                    ↑
            [Jabber Bot] (tool)
```

The AI Agent can now autonomously send XMPP messages, typing notifications, and files when the conversation requires it.

## File Structure

```
n8n-nodes-jabber-bot/
├── package.json
├── tsconfig.json
├── webpack.config.js
├── src/
│   ├── index.ts                 # Entry point
│   ├── icons/
│   │   ├── jabber-bot.svg
│   │   └── jabber-bot-trigger.svg
│   ├── credentials/
│   │   └── JabberBotApi.credentials.ts
│   └── nodes/
│       └── JabberBot/
│           ├── JabberBot.node.ts
│           └── JabberBotTrigger.node.ts
├── SPEC.md
└── README.md
```

## API Reference

See [Jabber Bot OpenAPI Specification](../docs/openapi.yaml) for complete API details.