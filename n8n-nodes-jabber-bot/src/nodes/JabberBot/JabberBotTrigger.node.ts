import {
  IWebhookFunctions,
  INodeType,
  INodeTypeDescription,
  IWebhookResponseData,
  IDataObject,
} from 'n8n-workflow';

export class JabberBotTrigger implements INodeType {
  description: INodeTypeDescription = {
    displayName: 'Jabber Bot Trigger',
    name: 'jabberBotTrigger',
    icon: 'file:icons/jabber-bot-trigger.svg',
    group: ['trigger'],
    version: 1,
    description: 'Trigger workflow when Jabber Bot receives a message',
    documentationUrl: 'https://github.com/aleskapot/jabber-bot',
    defaults: {
      name: 'Jabber Bot Trigger',
    },
    inputs: [],
    outputs: ['main'],
    webhooks: [
      {
        name: 'default',
        httpMethod: 'POST',
        path: 'webhook',
        responseMode: 'onReceived',
        responseData: 'allEntries',
      },
    ],
    properties: [],
  };

  async webhook(this: IWebhookFunctions): Promise<IWebhookResponseData> {
    const body = this.getBodyData();

    let outputData: IDataObject = {};

    if (body) {
      const payload = body as {
        message?: {
          id?: string;
          from?: string;
          to?: string;
          body?: string;
          type?: string;
          subject?: string;
          thread?: string;
          stamp?: string;
        };
        timestamp?: string;
        source?: string;
      };

      if (payload.message) {
        outputData = {
          id: payload.message.id as string,
          from: payload.message.from as string,
          to: payload.message.to as string,
          body: payload.message.body as string,
          type: payload.message.type as string,
          subject: payload.message.subject as string,
          thread: payload.message.thread as string,
          stamp: payload.message.stamp as string,
          timestamp: payload.timestamp as string,
          source: payload.source as string,
        };
      } else {
        outputData = body as IDataObject;
      }
    }

    return {
      workflowData: [[{ json: outputData }]],
    };
  }
}