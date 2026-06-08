import {
  ICredentialType,
  INodeProperties,
  Icon,
} from 'n8n-workflow';

export class JabberBotApi implements ICredentialType {
  name = 'jabberBotApi';

  displayName = 'Jabber Bot';

  documentationUrl = 'https://github.com/aleskapot/jabber-bot';

  icon: Icon = 'file:icons/jabber-bot.svg';

  properties: INodeProperties[] = [
    {
      displayName: 'Base URL',
      name: 'baseUrl',
      type: 'string',
      default: 'http://localhost:8080/api/v1',
      placeholder: 'http://localhost:8080/api/v1',
      description: 'The base URL of the Jabber Bot API (including /api/v1)',
    },
    {
      displayName: 'API Key',
      name: 'apiKey',
      type: 'string',
      typeOptions: {
        password: true,
      },
      default: '',
      description: 'API Key for authentication (set in API configuration)',
    },
  ];
}