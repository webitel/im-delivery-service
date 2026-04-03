
    const schema = {
  "asyncapi": "3.0.0",
  "info": {
    "title": "Webitel IM Delivery Service",
    "version": "1.0.0",
    "description": "WebSocket API for real-time chat delivery."
  },
  "channels": {
    "ws": {
      "address": "/ws/im",
      "messages": {
        "ServerEvent": {
          "name": "ServerEvent",
          "payload": {
            "type": "object",
            "properties": {
              "id": {
                "type": "string",
                "format": "uuid",
                "x-parser-schema-id": "<anonymous-schema-1>"
              },
              "created_at": {
                "type": "integer",
                "format": "int64",
                "x-parser-schema-id": "<anonymous-schema-2>"
              },
              "priority": {
                "type": "integer",
                "enum": [
                  0,
                  1,
                  2,
                  3
                ],
                "x-enumNames": [
                  "unspecified",
                  "high",
                  "normal",
                  "low"
                ],
                "x-parser-schema-id": "EventPriority"
              },
              "payload": {
                "type": "object",
                "properties": {
                  "connected_event": {
                    "type": "object",
                    "properties": {
                      "ok": {
                        "type": "boolean",
                        "x-parser-schema-id": "<anonymous-schema-3>"
                      },
                      "connection_id": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-4>"
                      },
                      "server_version": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-5>"
                      }
                    },
                    "x-parser-schema-id": "ConnectedPayload"
                  },
                  "disconnected_event": {
                    "type": "object",
                    "properties": {
                      "reason": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-6>"
                      },
                      "code": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-7>"
                      },
                      "status": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-8>"
                      }
                    },
                    "x-parser-schema-id": "DisconnectedPayload"
                  },
                  "message_event": {
                    "type": "object",
                    "properties": {
                      "id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-9>"
                      },
                      "send_id": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-10>"
                      },
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-11>"
                      },
                      "sender": {
                        "type": "object",
                        "properties": {
                          "sub": {
                            "type": "string",
                            "description": "Subject identifier (p.Sub)",
                            "x-parser-schema-id": "<anonymous-schema-12>"
                          },
                          "iss": {
                            "type": "string",
                            "description": "Issuer (p.Issuer)",
                            "x-parser-schema-id": "<anonymous-schema-13>"
                          },
                          "name": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-14>"
                          },
                          "type": {
                            "type": "string",
                            "description": "Normalized type (e.g. user, bot, agent)",
                            "x-parser-schema-id": "<anonymous-schema-15>"
                          },
                          "is_bot": {
                            "type": "boolean",
                            "x-parser-schema-id": "<anonymous-schema-16>"
                          }
                        },
                        "x-parser-schema-id": "WSPeer"
                      },
                      "to": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-17>"
                      },
                      "edited_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-18>"
                      },
                      "body": {
                        "type": "string",
                        "description": "The main text content",
                        "x-parser-schema-id": "<anonymous-schema-19>"
                      },
                      "type": {
                        "type": "string",
                        "enum": [
                          "text",
                          "image",
                          "document"
                        ],
                        "description": "Discriminator for the content field",
                        "x-parser-schema-id": "<anonymous-schema-20>"
                      },
                      "content": {
                        "type": "object",
                        "additionalProperties": true,
                        "description": "Media payload (image/document metadata)",
                        "x-parser-schema-id": "<anonymous-schema-21>"
                      }
                    },
                    "x-parser-schema-id": "WSMessage"
                  },
                  "thread_created_event": {
                    "type": "object",
                    "properties": {
                      "id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-22>"
                      },
                      "domain_id": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-23>"
                      },
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-24>"
                      },
                      "subject": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-25>"
                      },
                      "type": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-26>"
                      },
                      "members": {
                        "type": "array",
                        "items": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                        "x-parser-schema-id": "<anonymous-schema-27>"
                      }
                    },
                    "x-parser-schema-id": "WSThread"
                  },
                  "ack_event": {
                    "type": "object",
                    "properties": {
                      "id": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-28>"
                      },
                      "status": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-29>"
                      }
                    },
                    "x-parser-schema-id": "AckPayload"
                  },
                  "error_event": {
                    "type": "object",
                    "properties": {
                      "code": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-30>"
                      },
                      "message": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-31>"
                      },
                      "details": {
                        "type": "object",
                        "additionalProperties": true,
                        "x-parser-schema-id": "<anonymous-schema-32>"
                      }
                    },
                    "x-parser-schema-id": "ErrorPayload"
                  },
                  "ping_event": {
                    "type": "object",
                    "properties": {
                      "timestamp": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-33>"
                      }
                    },
                    "x-parser-schema-id": "PingPayload"
                  }
                },
                "x-parser-schema-id": "EventPayload"
              }
            },
            "x-parser-schema-id": "ServerEventPayload"
          },
          "x-parser-unique-object-id": "ServerEvent"
        }
      },
      "x-parser-unique-object-id": "ws"
    }
  },
  "operations": {
    "receiveEvents": {
      "action": "receive",
      "channel": "$ref:$.channels.ws",
      "messages": [
        "$ref:$.channels.ws.messages.ServerEvent"
      ],
      "x-parser-unique-object-id": "receiveEvents"
    }
  },
  "components": {
    "messages": {
      "ServerEvent": "$ref:$.channels.ws.messages.ServerEvent"
    },
    "schemas": {
      "ServerEventPayload": "$ref:$.channels.ws.messages.ServerEvent.payload",
      "EventPriority": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.priority",
      "EventPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload",
      "ConnectedPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.connected_event",
      "DisconnectedPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.disconnected_event",
      "AckPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.ack_event",
      "ErrorPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.error_event",
      "PingPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.ping_event",
      "WSPeer": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
      "WSMessage": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event",
      "WSThread": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.thread_created_event"
    }
  },
  "x-parser-spec-parsed": true,
  "x-parser-api-version": 3,
  "x-parser-spec-stringified": true
};
    const config = {"show":{"sidebar":true},"sidebar":{"showOperations":"byDefault"}};
    const appRoot = document.getElementById('root');
    AsyncApiStandalone.render(
        { schema, config, }, appRoot
    );
  