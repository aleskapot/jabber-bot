# n8n Nodes for Jabber Bot

Custom n8n node package for integrating with [Jabber Bot](https://github.com/aleskapot/jabber-bot) - an XMPP messenger with REST API.

![npm](https://img.shields.io/npm/v/n8n-nodes-jabber-bot)
![License](https://img.shields.io/github/license/aleskapot/jabber-bot)
![n8n version](https://img.shields.io/badge/n8n-v2.4%2B-blue)

## Features

- **Send Message** - Send XMPP messages to users
- **Send Chat State** - Send typing notifications (XEP-0085)
- **Send File** - Upload and send files via XMPP
- **Webhook Trigger** - Trigger workflows on incoming messages

## Requirements

- n8n v2.4 or higher
- Node.js 18+

## Installation

### From npm (not published yet)

```bash
npm install n8n-nodes-jabber-bot
```

### Development

```bash
# Clone the repository
git clone https://github.com/aleskapot/jabber-bot.git
cd jabber-bot/n8n-nodes-jabber-bot

# Install dependencies
npm install

# Build
npm run build

# Copy to n8n nodes directory
cp -r dist/* ~/.n8n/nodes/
```

Restart n8n to load the new nodes.

## Configuration

### 1. Create Credentials

1. Open n8n
2. Go to **Settings** → **Credentials**
3. Add new credential type: **Jabber Bot API**
4. Fill in:
   - **Base URL**: Your Jabber Bot API URL (e.g., `http://192.168.1.100:8080/api/v1`)
   - **API Key**: Your API key (configured in Jabber Bot's `config.yaml`)

### 2. Configure Jabber Bot (for Trigger)

To use the trigger node, configure Jabber Bot's webhook URL:

```yaml
# config.yaml
webhook:
  url: "https://your-n8n-instance.com/webhook/your-unique-path"
```

## Usage

### Send Message

1. Add **Jabber Bot** node to your workflow
2. Select operation: **Send Message**
3. Fill in:
   - **To**: Recipient JID (e.g., `user@example.com`)
   - **Body**: Message text
   - **Message Type**: chat, groupchat, headline, or normal

### Send Chat State

1. Add **Jabber Bot** node
2. Select operation: **Send Chat State**
3. Fill in:
   - **To**: Recipient JID
   - **Chat State**: active, composing, paused, inactive, or gone

### Send File

1. Add **Jabber Bot** node
2. Select operation: **Send File**
3. Fill in:
   - **To**: Recipient JID
   - **File**: Attach binary file
   - **Description**: Optional file description

### Webhook Trigger

1. Add **Jabber Bot Trigger** node
2. Activate the workflow
3. Copy the webhook URL shown in the node
4. Configure Jabber Bot's `webhook.url` to point to this URL
5. When Jabber Bot receives a message, it will trigger your workflow

### Tool Usage (AI Agent)

The Jabber Bot node can be used as a tool in AI Agent workflows. When connected to an AI Agent, the agent can autonomously send XMPP messages, typing notifications, and files.

**Setup:**

1. Add an **AI Agent** node to your workflow
2. Connect a Chat Model (OpenAI, Anthropic, etc.) to the Agent
3. Connect the **Jabber Bot** node to the Agent's **Tool** input
4. Configure Jabber Bot credentials

**Example workflow:**

```
[Chat Trigger] → [AI Agent] → [Respond to Webhook]
                    ↑
            [Jabber Bot] (tool)
```

The AI Agent decides when to call the Jabber Bot tool based on the conversation context.

## Example Workflows

### Forward Jabber messages to Telegram

```
[Jabber Bot Trigger] → [Telegram: Send Message]
```

### Log all incoming messages

```
[Jabber Bot Trigger] → [Write File]
```

### Send notification on specific keywords

```
[Jabber Bot Trigger] → [IF: contains "urgent"] → [Slack: Send Message]
```

### AI Agent sends Jabber messages

```
[Chat Trigger] → [AI Agent] → [Respond to Webhook]
                    ↑
            [Jabber Bot] (tool)
```

## API Documentation

For complete API details, see the [Jabber Bot OpenAPI Specification](../docs/openapi.yaml).

## License

MIT License - see [LICENSE](../LICENSE) for details.

## Support

- Issues: https://github.com/aleskapot/jabber-bot/issues
- GitHub: https://github.com/aleskapot/jabber-bot