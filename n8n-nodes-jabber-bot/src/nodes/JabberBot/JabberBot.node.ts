import {
  IExecuteFunctions,
  INodeExecutionData,
  INodeType,
  INodeTypeDescription,
  IHttpRequestOptions,
  IDataObject,
} from 'n8n-workflow';

// noinspection ExceptionCaughtLocallyJS
export class JabberBot implements INodeType {
  description: INodeTypeDescription = {
    displayName: 'Jabber Bot',
    name: 'jabberBot',
    icon: 'file:icons/jabber-bot.svg',
    group: ['output'],
    version: 1,
    description: 'Send messages and files via XMPP Jabber Bot',
    documentationUrl: 'https://github.com/aleskapot/jabber-bot',
    subtitle: '={{$parameter.operation}}',
    defaults: {
      name: 'Jabber Bot',
    },
    usableAsTool: true,
    inputs: ['main'],
    outputs: ['main'],
    credentials: [
      {
        name: 'jabberBotApi',
        required: true,
      },
    ],
    properties: [
      {
        displayName: 'Operation',
        name: 'operation',
        type: 'options',
        noDataExpression: true,
        options: [
          {
            name: 'Send Message',
            value: 'sendMessage',
            description: 'Send an XMPP message to a user',
          },
          {
            name: 'Send Chat State',
            value: 'sendChatState',
            description: 'Send a chat state notification (typing, paused, etc.)',
          },
          {
            name: 'Send File',
            value: 'sendFile',
            description: 'Upload and send a file via XMPP',
          },
        ],
        default: 'sendMessage',
      },
      {
        displayName: 'To',
        name: 'to',
        type: 'string',
        default: '',
        placeholder: 'user@example.com',
        description: 'The JID of the message recipient',
        required: true,
      },
      {
        displayName: 'Message Type',
        name: 'type',
        type: 'options',
        options: [
          {
            name: 'Chat',
            value: 'chat',
          },
          {
            name: 'Group Chat',
            value: 'groupchat',
          },
          {
            name: 'Headline',
            value: 'headline',
          },
          {
            name: 'Normal',
            value: 'normal',
          },
        ],
        default: 'chat',
        displayOptions: {
          show: {
            operation: ['sendMessage'],
          },
        },
        description: 'Type of XMPP message',
      },
      {
        displayName: 'Body',
        name: 'body',
        type: 'string',
        typeOptions: {
          rows: 4,
        },
        default: '',
        placeholder: 'Hello, world!',
        displayOptions: {
          show: {
            operation: ['sendMessage'],
          },
        },
        description: 'The message content (max 10,000 characters)',
      },
      {
        displayName: 'Chat State',
        name: 'state',
        type: 'options',
        options: [
          {
            name: 'Active',
            value: 'active',
            description: 'User is actively participating',
          },
          {
            name: 'Composing',
            value: 'composing',
            description: 'User is typing a message',
          },
          {
            name: 'Paused',
            value: 'paused',
            description: 'User paused while composing',
          },
          {
            name: 'Inactive',
            value: 'inactive',
            description: 'User not actively participating',
          },
          {
            name: 'Gone',
            value: 'gone',
            description: 'User ended the session',
          },
        ],
        default: 'composing',
        displayOptions: {
          show: {
            operation: ['sendChatState'],
          },
        },
        description: 'Chat state notification (XEP-0085)',
      },
      {
        displayName: 'File',
        name: 'file',
        type: 'resourceLocator',
        default: { mode: 'list', value: '' },
        displayOptions: {
          show: {
            operation: ['sendFile'],
          },
        },
        description: 'The file to upload and send',
      },
      {
        displayName: 'Description',
        name: 'description',
        type: 'string',
        default: '',
        displayOptions: {
          show: {
            operation: ['sendFile'],
          },
        },
        description: 'Optional description of the file',
      },
      {
        displayName: 'Additional Fields',
        name: 'additionalFields',
        type: 'collection',
        placeholder: 'Add Field',
        default: {},
        options: [
          {
            displayName: 'Include Raw Response',
            name: 'includeRawResponse',
            type: 'boolean',
            default: false,
            description: 'Include the raw API response in the output',
          },
        ],
      },
    ],
  };

  async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
    const items = this.getInputData();
    const credentials = await this.getCredentials('jabberBotApi');
    const baseUrl = (credentials.baseUrl as string).replace(/\/$/, '');
    const apiKey = credentials.apiKey as string;

    const operation = this.getNodeParameter('operation', 0) as string;

    const results: INodeExecutionData[] = [];

    for (let i = 0; i < items.length; i++) {
      try {
        const to = this.getNodeParameter('to', i) as string;

        const requestOptions: IHttpRequestOptions = {
          headers: {
            'API-Key': apiKey,
          },
          method: 'POST',
          url: '',
          json: true,
        };

        switch (operation) {
          case 'sendMessage': {
            const messageType = this.getNodeParameter('type', i) as string;
            const messageBody = this.getNodeParameter('body', i) as string;

            requestOptions.url = `${baseUrl}/send`;
            requestOptions.body = {
              to,
              body: messageBody,
              type: messageType,
            };
            break;
          }

          case 'sendChatState': {
            const state = this.getNodeParameter('state', i) as string;

            requestOptions.url = `${baseUrl}/chat-state`;
            requestOptions.body = {
              to,
              state,
            };
            break;
          }

          case 'sendFile': {
            const binaryData = this.getNodeParameter('file', i) as {
              mode: string;
              value: string;
              binaryPropertyName?: string;
            };

            if (!binaryData?.binaryPropertyName) {
              throw new Error('No file attached');
            }

            const item = items[i];
            const binaryFile = item.binary?.[binaryData.binaryPropertyName];

            if (!binaryFile) {
              throw new Error('No file data found');
            }

            const fileBuffer = Buffer.from(binaryFile.data, 'base64');
            const description = this.getNodeParameter('description', i) as string;

            const formData = new (require('form-data'))();
            formData.append('to', to);
            if (description) {
              formData.append('description', description);
            }
            formData.append('file', fileBuffer, {
              filename: binaryFile.fileName || binaryData.value || 'file',
            });

            requestOptions.url = `${baseUrl}/send-file`;
            requestOptions.body = formData as unknown as object;
            requestOptions.headers = {
              'API-Key': apiKey,
              ...formData.getHeaders(),
            };
            requestOptions.json = false;
            break;
          }

          default:
            throw new Error(`Unknown operation: ${operation}`);
        }

        const response = await this.helpers.httpRequest(requestOptions);

        const additionalFields = this.getNodeParameter('additionalFields', i) as {
          includeRawResponse?: boolean;
        };

        const outputData: IDataObject = {
          success: response.success as boolean,
          message: response.message as string,
          ...(response.data as IDataObject),
        };

        if (additionalFields.includeRawResponse) {
          outputData.rawResponse = response;
        }

        results.push({
          json: outputData,
        });
      } catch (error) {
        if (this.continueOnFail()) {
          results.push({
            json: {
              error: error instanceof Error ? error.message : 'Unknown error',
            },
          });
        } else {
          throw error;
        }
      }
    }

    return [results];
  }
}