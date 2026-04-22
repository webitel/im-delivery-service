
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
                          "id": {
                            "type": "string",
                            "description": "Member identifier",
                            "x-parser-schema-id": "<anonymous-schema-12>"
                          },
                          "contact": {
                            "type": "object",
                            "required": [
                              "sub",
                              "type"
                            ],
                            "properties": {
                              "sub": {
                                "type": "string",
                                "description": "Subject identifier",
                                "x-parser-schema-id": "<anonymous-schema-13>"
                              },
                              "iss": {
                                "type": "string",
                                "description": "Issuer",
                                "x-parser-schema-id": "<anonymous-schema-14>"
                              },
                              "name": {
                                "type": "string",
                                "x-parser-schema-id": "<anonymous-schema-15>"
                              },
                              "username": {
                                "type": "string",
                                "x-parser-schema-id": "<anonymous-schema-16>"
                              },
                              "type": {
                                "type": "string",
                                "description": "Normalized type (e.g. user, bot, agent)",
                                "x-parser-schema-id": "<anonymous-schema-17>"
                              },
                              "is_bot": {
                                "type": "boolean",
                                "x-parser-schema-id": "<anonymous-schema-18>"
                              }
                            },
                            "x-parser-schema-id": "WSContact"
                          },
                          "role": {
                            "type": "string",
                            "description": "Role (e.g. member, admin)",
                            "x-parser-schema-id": "<anonymous-schema-19>"
                          }
                        },
                        "x-parser-schema-id": "WSPeer"
                      },
                      "to": {
                        "type": "array",
                        "items": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                        "x-parser-schema-id": "<anonymous-schema-20>"
                      },
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-21>"
                      },
                      "edited_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-22>"
                      },
                      "body": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-23>"
                      },
                      "type": {
                        "type": "string",
                        "enum": [
                          "text",
                          "image",
                          "document"
                        ],
                        "x-parser-schema-id": "<anonymous-schema-24>"
                      },
                      "images": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "additionalProperties": true,
                          "x-parser-schema-id": "<anonymous-schema-26>"
                        },
                        "x-parser-schema-id": "<anonymous-schema-25>"
                      },
                      "documents": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "additionalProperties": true,
                          "x-parser-schema-id": "<anonymous-schema-28>"
                        },
                        "x-parser-schema-id": "<anonymous-schema-27>"
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
                        "x-parser-schema-id": "<anonymous-schema-29>"
                      },
                      "domain_id": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-30>"
                      },
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-31>"
                      },
                      "subject": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-32>"
                      },
                      "type": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-33>"
                      },
                      "members": {
                        "type": "array",
                        "items": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                        "x-parser-schema-id": "<anonymous-schema-34>"
                      }
                    },
                    "x-parser-schema-id": "WSThread"
                  },
                  "ack_event": {
                    "type": "object",
                    "properties": {
                      "id": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-35>"
                      },
                      "status": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-36>"
                      }
                    },
                    "x-parser-schema-id": "AckPayload"
                  },
                  "error_event": {
                    "type": "object",
                    "properties": {
                      "code": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-37>"
                      },
                      "message": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-38>"
                      },
                      "details": {
                        "type": "object",
                        "additionalProperties": true,
                        "x-parser-schema-id": "<anonymous-schema-39>"
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
                        "x-parser-schema-id": "<anonymous-schema-40>"
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
      "WSContact": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender.properties.contact",
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
  