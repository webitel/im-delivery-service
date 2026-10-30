
    const schema = {
  "asyncapi": "3.0.0",
  "info": {
    "title": "Webitel IM Delivery Service",
    "version": "1.1.0",
    "description": "WebSocket API for real-time chat delivery.\n\nEvery server frame is a `ServerEvent` envelope. Its `payload` object carries\nexactly one key that names the event (e.g. `message_event`); the value is\nthat event's body. Clients switch on the key and ignore unknown keys.\n\nEvents that carry `updates_cursor` advance the recipient's GetUpdates\nposition: persist it, and on reconnect compare it with the cursor from\n`connected_event`. A `resync_event` means live events were dropped and the\nclient must catch up via GetUpdates.\n"
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
                "description": "Envelope id; a live client acks by this value",
                "x-parser-schema-id": "<anonymous-schema-1>"
              },
              "created_at": {
                "type": "integer",
                "format": "int64",
                "description": "Event time, Unix ms",
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
                      },
                      "updates_cursor": {
                        "type": "string",
                        "description": "User's GetUpdates cursor now; equal to the saved one means nothing was missed",
                        "x-parser-schema-id": "<anonymous-schema-6>"
                      }
                    },
                    "x-parser-schema-id": "ConnectedPayload"
                  },
                  "disconnected_event": {
                    "type": "object",
                    "description": "Sent right before the server closes the stream",
                    "properties": {
                      "reason": {
                        "type": "string",
                        "description": "Human-readable reason",
                        "x-parser-schema-id": "<anonymous-schema-7>"
                      },
                      "code": {
                        "type": "integer",
                        "description": "HTTP-like code: 401, 408, 429, 503",
                        "x-parser-schema-id": "<anonymous-schema-8>"
                      },
                      "status": {
                        "type": "string",
                        "title": "DisconnectStatus",
                        "enum": [
                          "SHUTDOWN",
                          "EVICTED",
                          "TIMEOUT",
                          "UNAUTHORIZED"
                        ],
                        "x-parser-schema-id": "<anonymous-schema-9>"
                      }
                    },
                    "x-parser-schema-id": "DisconnectedPayload"
                  },
                  "resync_event": {
                    "type": "object",
                    "description": "Live events for this connection were dropped; catch up via GetUpdates",
                    "properties": {
                      "cursor": {
                        "type": "string",
                        "description": "Cursor to resume from; empty means use the client's saved cursor",
                        "x-parser-schema-id": "<anonymous-schema-10>"
                      }
                    },
                    "x-parser-schema-id": "ResyncPayload"
                  },
                  "message_event": {
                    "type": "object",
                    "properties": {
                      "id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-11>"
                      },
                      "send_id": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-12>"
                      },
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-13>"
                      },
                      "sender": {
                        "type": "object",
                        "description": "Thread participant; same shape for sender, edited_by, deleted_by, member and from",
                        "properties": {
                          "id": {
                            "type": "string",
                            "description": "Member identifier",
                            "x-parser-schema-id": "<anonymous-schema-14>"
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
                                "x-parser-schema-id": "<anonymous-schema-15>"
                              },
                              "iss": {
                                "type": "string",
                                "description": "Issuer",
                                "x-parser-schema-id": "<anonymous-schema-16>"
                              },
                              "name": {
                                "type": "string",
                                "x-parser-schema-id": "<anonymous-schema-17>"
                              },
                              "username": {
                                "type": "string",
                                "x-parser-schema-id": "<anonymous-schema-18>"
                              },
                              "type": {
                                "type": "string",
                                "description": "Normalized type (e.g. user, bot, agent)",
                                "x-parser-schema-id": "<anonymous-schema-19>"
                              },
                              "is_bot": {
                                "type": "boolean",
                                "x-parser-schema-id": "<anonymous-schema-20>"
                              }
                            },
                            "x-parser-schema-id": "WSContact"
                          },
                          "role": {
                            "type": "string",
                            "description": "Role (e.g. member, admin)",
                            "x-parser-schema-id": "<anonymous-schema-21>"
                          }
                        },
                        "x-parser-schema-id": "WSPeer"
                      },
                      "to": {
                        "type": "array",
                        "items": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                        "x-parser-schema-id": "<anonymous-schema-22>"
                      },
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-23>"
                      },
                      "edited_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-24>"
                      },
                      "body": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-25>"
                      },
                      "type": {
                        "type": "string",
                        "enum": [
                          "text",
                          "image",
                          "document",
                          "system",
                          "interactive",
                          "location",
                          "contact"
                        ],
                        "x-parser-schema-id": "MessageType"
                      },
                      "images": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "properties": {
                            "id": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-27>"
                            },
                            "name": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-28>"
                            },
                            "mime": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-29>"
                            },
                            "url": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-30>"
                            }
                          },
                          "x-parser-schema-id": "WSImage"
                        },
                        "x-parser-schema-id": "<anonymous-schema-26>"
                      },
                      "documents": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "properties": {
                            "id": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-32>"
                            },
                            "name": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-33>"
                            },
                            "mime": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-34>"
                            },
                            "size": {
                              "type": "integer",
                              "format": "int64",
                              "x-parser-schema-id": "<anonymous-schema-35>"
                            },
                            "url": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-36>"
                            }
                          },
                          "x-parser-schema-id": "WSDocument"
                        },
                        "x-parser-schema-id": "<anonymous-schema-31>"
                      },
                      "interactive": {
                        "type": "object",
                        "additionalProperties": true,
                        "description": "Raw interactive (buttons/keyboard) payload, passed through as-is",
                        "x-parser-schema-id": "<anonymous-schema-37>"
                      },
                      "contact": {
                        "type": "object",
                        "description": "Contact card shared in a message",
                        "properties": {
                          "name": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-38>"
                          },
                          "phone": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-39>"
                          },
                          "email": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-40>"
                          }
                        },
                        "x-parser-schema-id": "WSSharedContact"
                      },
                      "location": {
                        "type": "object",
                        "required": [
                          "latitude",
                          "longitude"
                        ],
                        "properties": {
                          "latitude": {
                            "type": "number",
                            "format": "double",
                            "x-parser-schema-id": "<anonymous-schema-41>"
                          },
                          "longitude": {
                            "type": "number",
                            "format": "double",
                            "x-parser-schema-id": "<anonymous-schema-42>"
                          },
                          "address": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-43>"
                          },
                          "name": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-44>"
                          }
                        },
                        "x-parser-schema-id": "WSLocation"
                      },
                      "system": {
                        "type": "object",
                        "properties": {
                          "type": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-45>"
                          },
                          "metadata": {
                            "type": "object",
                            "additionalProperties": true,
                            "x-parser-schema-id": "<anonymous-schema-46>"
                          }
                        },
                        "x-parser-schema-id": "WSSystem"
                      },
                      "metadata": {
                        "type": "object",
                        "additionalProperties": true,
                        "x-parser-schema-id": "<anonymous-schema-47>"
                      },
                      "reply_to": {
                        "type": "object",
                        "description": "Compact snapshot of the quoted (replied-to) message",
                        "properties": {
                          "message_id": {
                            "type": "string",
                            "format": "uuid",
                            "x-parser-schema-id": "<anonymous-schema-48>"
                          },
                          "sender": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                          "sender_id": {
                            "type": "string",
                            "format": "uuid",
                            "x-parser-schema-id": "<anonymous-schema-49>"
                          },
                          "type": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.type",
                          "body": {
                            "type": "string",
                            "description": "Body snippet, max 256 chars",
                            "x-parser-schema-id": "<anonymous-schema-50>"
                          },
                          "created_at": {
                            "type": "integer",
                            "format": "int64",
                            "x-parser-schema-id": "<anonymous-schema-51>"
                          },
                          "attachment_kind": {
                            "type": "string",
                            "title": "AttachmentKind",
                            "enum": [
                              "document",
                              "image"
                            ],
                            "x-parser-schema-id": "<anonymous-schema-52>"
                          },
                          "attachment_name": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-53>"
                          },
                          "attachment_mime": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-54>"
                          },
                          "attachment_address": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-55>"
                          },
                          "is_deleted": {
                            "type": "boolean",
                            "description": "The quoted message was deleted after the reply was sent",
                            "x-parser-schema-id": "<anonymous-schema-56>"
                          }
                        },
                        "x-parser-schema-id": "WSReplyTo"
                      },
                      "forward_origin": {
                        "type": "object",
                        "properties": {
                          "kind": {
                            "type": "string",
                            "title": "ForwardOriginKind",
                            "enum": [
                              "FORWARD_ORIGIN_KIND_UNSPECIFIED",
                              "FORWARD_ORIGIN_KIND_INTERNAL",
                              "FORWARD_ORIGIN_KIND_EXTERNAL_USER",
                              "FORWARD_ORIGIN_KIND_EXTERNAL_HIDDEN_USER",
                              "FORWARD_ORIGIN_KIND_EXTERNAL_CHAT"
                            ],
                            "x-parser-schema-id": "<anonymous-schema-57>"
                          },
                          "sender_id": {
                            "type": "string",
                            "format": "uuid",
                            "x-parser-schema-id": "<anonymous-schema-58>"
                          },
                          "sender_name": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-59>"
                          },
                          "original_sent_at": {
                            "type": "integer",
                            "format": "int64",
                            "x-parser-schema-id": "<anonymous-schema-60>"
                          },
                          "source_message_id": {
                            "type": "string",
                            "format": "uuid",
                            "x-parser-schema-id": "<anonymous-schema-61>"
                          }
                        },
                        "x-parser-schema-id": "WSForwardOrigin"
                      },
                      "seq": {
                        "type": "integer",
                        "format": "int64",
                        "description": "Per-thread sequence number; used to ack/read by (thread_id, seq)",
                        "x-parser-schema-id": "<anonymous-schema-62>"
                      },
                      "updates_cursor": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-63>"
                      }
                    },
                    "x-parser-schema-id": "WSMessage"
                  },
                  "message_edited_event": {
                    "type": "object",
                    "description": "Match the message by id and replace its content",
                    "properties": {
                      "id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-64>"
                      },
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-65>"
                      },
                      "edited_by": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                      "body": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-66>"
                      },
                      "type": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.type",
                      "version": {
                        "type": "integer",
                        "format": "int32",
                        "description": "Edit version; ignore an event older than the one applied",
                        "x-parser-schema-id": "<anonymous-schema-67>"
                      },
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-68>"
                      },
                      "edited_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-69>"
                      },
                      "metadata": {
                        "type": "object",
                        "additionalProperties": true,
                        "x-parser-schema-id": "<anonymous-schema-70>"
                      },
                      "updates_cursor": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-71>"
                      }
                    },
                    "x-parser-schema-id": "WSMessageEdited"
                  },
                  "message_deleted_event": {
                    "type": "object",
                    "description": "Match the message by id and remove it",
                    "properties": {
                      "id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-72>"
                      },
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-73>"
                      },
                      "deleted_by": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-74>"
                      },
                      "deleted_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-75>"
                      },
                      "updates_cursor": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-76>"
                      }
                    },
                    "x-parser-schema-id": "WSMessageDeleted"
                  },
                  "message_status_event": {
                    "type": "object",
                    "description": "Per-recipient delivery status change. delivered/read move the member's\nhorizon up to `up_to_seq`; failed names the single `message_id`.\n",
                    "properties": {
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-77>"
                      },
                      "status": {
                        "type": "string",
                        "title": "DeliveryStatus",
                        "enum": [
                          "delivered",
                          "read",
                          "failed"
                        ],
                        "x-parser-schema-id": "<anonymous-schema-78>"
                      },
                      "member": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                      "message_id": {
                        "type": "string",
                        "format": "uuid",
                        "description": "Failed message; set only when status is failed",
                        "x-parser-schema-id": "<anonymous-schema-79>"
                      },
                      "up_to_seq": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-80>"
                      },
                      "unread_count": {
                        "type": "integer",
                        "format": "int64",
                        "description": "Reader's unread count after a read; absent otherwise",
                        "x-parser-schema-id": "<anonymous-schema-81>"
                      },
                      "via": {
                        "type": "string",
                        "title": "StatusVia",
                        "enum": [
                          "ws",
                          "push",
                          "provider",
                          "bot"
                        ],
                        "x-parser-schema-id": "<anonymous-schema-82>"
                      },
                      "error": {
                        "type": "object",
                        "description": "Why the message did not reach the member",
                        "properties": {
                          "code": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-83>"
                          },
                          "message": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-84>"
                          }
                        },
                        "x-parser-schema-id": "WSStatusError"
                      },
                      "occurred_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-85>"
                      },
                      "updates_cursor": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-86>"
                      }
                    },
                    "x-parser-schema-id": "WSMessageStatus"
                  },
                  "message_reaction_event": {
                    "type": "object",
                    "description": "Authoritative reaction state of a message after the change; replace the\nwhole reaction bar. An empty `reactions` array means no reactions left.\n",
                    "required": [
                      "message_id",
                      "thread_id",
                      "reactions"
                    ],
                    "properties": {
                      "message_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-87>"
                      },
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-88>"
                      },
                      "reactions": {
                        "type": "array",
                        "items": {
                          "type": "object",
                          "properties": {
                            "emoji": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-90>"
                            },
                            "count": {
                              "type": "integer",
                              "format": "int32",
                              "x-parser-schema-id": "<anonymous-schema-91>"
                            },
                            "reacted_by_me": {
                              "type": "boolean",
                              "description": "Resolved per recipient",
                              "x-parser-schema-id": "<anonymous-schema-92>"
                            },
                            "reactor_ids": {
                              "type": "array",
                              "items": {
                                "type": "string",
                                "x-parser-schema-id": "<anonymous-schema-94>"
                              },
                              "x-parser-schema-id": "<anonymous-schema-93>"
                            },
                            "last_reacted_at": {
                              "type": "string",
                              "description": "Unix ms encoded as a string (protojson int64)",
                              "x-parser-schema-id": "<anonymous-schema-95>"
                            }
                          },
                          "x-parser-schema-id": "WSReactionAggregate"
                        },
                        "x-parser-schema-id": "<anonymous-schema-89>"
                      },
                      "updates_cursor": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-96>"
                      }
                    },
                    "x-parser-schema-id": "WSReaction"
                  },
                  "typing_event": {
                    "type": "object",
                    "description": "Ephemeral typing indicator; hide it after timeout_ms",
                    "properties": {
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-97>"
                      },
                      "timeout_ms": {
                        "type": "integer",
                        "format": "int32",
                        "x-parser-schema-id": "<anonymous-schema-98>"
                      },
                      "from": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                      "preview_text": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-99>"
                      }
                    },
                    "x-parser-schema-id": "WSTyping"
                  },
                  "thread_created_event": {
                    "type": "object",
                    "properties": {
                      "id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-100>"
                      },
                      "domain_id": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-101>"
                      },
                      "created_at": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-102>"
                      },
                      "subject": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-103>"
                      },
                      "type": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-104>"
                      },
                      "members": {
                        "type": "array",
                        "items": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
                        "x-parser-schema-id": "<anonymous-schema-105>"
                      }
                    },
                    "x-parser-schema-id": "WSThread"
                  },
                  "thread_tags_changed_event": {
                    "type": "object",
                    "description": "Recipient's own current tags on the thread (full replacement)",
                    "required": [
                      "thread_id",
                      "tags"
                    ],
                    "properties": {
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-106>"
                      },
                      "tags": {
                        "type": "array",
                        "items": {
                          "type": "string",
                          "x-parser-schema-id": "<anonymous-schema-108>"
                        },
                        "x-parser-schema-id": "<anonymous-schema-107>"
                      },
                      "updates_cursor": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-109>"
                      }
                    },
                    "x-parser-schema-id": "ThreadTagsChangedPayload"
                  },
                  "member_added_event": {
                    "type": "object",
                    "description": "Body of member_added_event and member_left_event",
                    "properties": {
                      "thread_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-110>"
                      },
                      "contact_id": {
                        "type": "string",
                        "format": "uuid",
                        "x-parser-schema-id": "<anonymous-schema-111>"
                      },
                      "metadata": {
                        "type": "object",
                        "additionalProperties": true,
                        "x-parser-schema-id": "<anonymous-schema-112>"
                      },
                      "updates_cursor": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-113>"
                      },
                      "action": {
                        "type": "string",
                        "title": "MemberAction",
                        "enum": [
                          "joined",
                          "left"
                        ],
                        "x-parser-schema-id": "<anonymous-schema-114>"
                      }
                    },
                    "x-parser-schema-id": "MemberEventPayload"
                  },
                  "member_left_event": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.member_added_event",
                  "interactive_callback": {
                    "type": "object",
                    "description": "A participant pressed an interactive button",
                    "properties": {
                      "reacted_by": {
                        "type": "object",
                        "description": "Internal peer shape (used only by interactive_callback)",
                        "properties": {
                          "id": {
                            "type": "string",
                            "format": "uuid",
                            "x-parser-schema-id": "<anonymous-schema-115>"
                          },
                          "type": {
                            "type": "integer",
                            "title": "PeerKind",
                            "enum": [
                              1,
                              2,
                              3
                            ],
                            "x-enumNames": [
                              "user",
                              "group",
                              "channel"
                            ],
                            "x-parser-schema-id": "<anonymous-schema-116>"
                          },
                          "sub": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-117>"
                          },
                          "issuer": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-118>"
                          },
                          "name": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-119>"
                          },
                          "username": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-120>"
                          },
                          "is_bot": {
                            "type": "boolean",
                            "x-parser-schema-id": "<anonymous-schema-121>"
                          },
                          "member_id": {
                            "type": "string",
                            "x-parser-schema-id": "<anonymous-schema-122>"
                          },
                          "role": {
                            "type": "integer",
                            "format": "int32",
                            "x-parser-schema-id": "<anonymous-schema-123>"
                          },
                          "domain_id": {
                            "type": "integer",
                            "x-parser-schema-id": "<anonymous-schema-124>"
                          }
                        },
                        "x-parser-schema-id": "RawPeer"
                      },
                      "in_reply_to": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-125>"
                      },
                      "button_code": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-126>"
                      },
                      "callback_data": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-127>"
                      },
                      "reacted_at": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-128>"
                      },
                      "domain_id": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-129>"
                      },
                      "thread_id": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-130>"
                      },
                      "receiver": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.interactive_callback.properties.reacted_by"
                    },
                    "x-parser-schema-id": "InteractiveCallbackPayload"
                  },
                  "variable_set_event": {
                    "type": "object",
                    "description": "Body of variable_set_event and variable_flush_event",
                    "properties": {
                      "variables": {
                        "type": "object",
                        "additionalProperties": {
                          "type": "object",
                          "properties": {
                            "value": {
                              "description": "Any JSON value",
                              "x-parser-schema-id": "<anonymous-schema-132>"
                            },
                            "set_by": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-133>"
                            },
                            "set_at": {
                              "type": "string",
                              "x-parser-schema-id": "<anonymous-schema-134>"
                            }
                          },
                          "x-parser-schema-id": "VariableEntry"
                        },
                        "x-parser-schema-id": "<anonymous-schema-131>"
                      },
                      "action": {
                        "type": "string",
                        "title": "VariableAction",
                        "enum": [
                          "set",
                          "flush"
                        ],
                        "x-parser-schema-id": "<anonymous-schema-135>"
                      }
                    },
                    "x-parser-schema-id": "WSVariables"
                  },
                  "variable_flush_event": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.variable_set_event",
                  "ack_event": {
                    "type": "object",
                    "description": "Reserved; not emitted by the current server",
                    "properties": {
                      "id": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-136>"
                      },
                      "status": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-137>"
                      }
                    },
                    "x-parser-schema-id": "AckPayload"
                  },
                  "error_event": {
                    "type": "object",
                    "description": "Reserved; not emitted by the current server",
                    "properties": {
                      "code": {
                        "type": "integer",
                        "x-parser-schema-id": "<anonymous-schema-138>"
                      },
                      "message": {
                        "type": "string",
                        "x-parser-schema-id": "<anonymous-schema-139>"
                      },
                      "details": {
                        "type": "object",
                        "additionalProperties": true,
                        "x-parser-schema-id": "<anonymous-schema-140>"
                      }
                    },
                    "x-parser-schema-id": "ErrorPayload"
                  },
                  "ping_event": {
                    "type": "object",
                    "description": "Reserved; liveness uses WebSocket ping/pong control frames",
                    "properties": {
                      "timestamp": {
                        "type": "integer",
                        "format": "int64",
                        "x-parser-schema-id": "<anonymous-schema-141>"
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
        },
        "AuthRequest": {
          "name": "AuthRequest",
          "payload": {
            "type": "object",
            "required": [
              "x-webitel-access"
            ],
            "properties": {
              "x-webitel-access": {
                "type": "string",
                "description": "Access token",
                "x-parser-schema-id": "<anonymous-schema-142>"
              },
              "x-webitel-client": {
                "type": "string",
                "description": "Optional client identifier",
                "x-parser-schema-id": "<anonymous-schema-143>"
              }
            },
            "x-parser-schema-id": "AuthRequestPayload"
          },
          "x-parser-unique-object-id": "AuthRequest"
        },
        "ClientCommand": {
          "name": "ClientCommand",
          "payload": {
            "type": "object",
            "description": "`ack` confirms delivery, `read` marks as read. A live client sends\n`event_id` (the envelope id); a reconnecting client sends\n`thread_id` + `seq` from history instead.\n",
            "required": [
              "type"
            ],
            "properties": {
              "type": {
                "type": "string",
                "title": "ClientCommandType",
                "enum": [
                  "ack",
                  "read"
                ],
                "x-parser-schema-id": "<anonymous-schema-144>"
              },
              "event_id": {
                "type": "string",
                "format": "uuid",
                "x-parser-schema-id": "<anonymous-schema-145>"
              },
              "thread_id": {
                "type": "string",
                "format": "uuid",
                "x-parser-schema-id": "<anonymous-schema-146>"
              },
              "seq": {
                "type": "integer",
                "format": "int64",
                "x-parser-schema-id": "<anonymous-schema-147>"
              }
            },
            "x-parser-schema-id": "ClientCommandPayload"
          },
          "x-parser-unique-object-id": "ClientCommand"
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
    },
    "authenticate": {
      "action": "send",
      "channel": "$ref:$.channels.ws",
      "summary": "Late-binding auth frame.",
      "description": "Sent as the first frame when the token was not supplied on the upgrade\nrequest. Must arrive within 5s, otherwise the server closes the socket\nwith `401_unauthorized`.\n",
      "messages": [
        "$ref:$.channels.ws.messages.AuthRequest"
      ],
      "x-parser-unique-object-id": "authenticate"
    },
    "sendCommands": {
      "action": "send",
      "channel": "$ref:$.channels.ws",
      "summary": "Delivery acknowledgements and read receipts.",
      "messages": [
        "$ref:$.channels.ws.messages.ClientCommand"
      ],
      "x-parser-unique-object-id": "sendCommands"
    }
  },
  "components": {
    "messages": {
      "ServerEvent": "$ref:$.channels.ws.messages.ServerEvent",
      "AuthRequest": "$ref:$.channels.ws.messages.AuthRequest",
      "ClientCommand": "$ref:$.channels.ws.messages.ClientCommand"
    },
    "schemas": {
      "ServerEventPayload": "$ref:$.channels.ws.messages.ServerEvent.payload",
      "EventPriority": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.priority",
      "EventPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload",
      "ConnectedPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.connected_event",
      "DisconnectedPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.disconnected_event",
      "ResyncPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.resync_event",
      "AckPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.ack_event",
      "ErrorPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.error_event",
      "PingPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.ping_event",
      "WSContact": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender.properties.contact",
      "WSPeer": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.sender",
      "MessageType": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.type",
      "WSMessage": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event",
      "WSImage": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.images.items",
      "WSDocument": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.documents.items",
      "WSSharedContact": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.contact",
      "WSLocation": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.location",
      "WSSystem": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.system",
      "WSForwardOrigin": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.forward_origin",
      "WSReplyTo": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_event.properties.reply_to",
      "WSMessageEdited": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_edited_event",
      "WSMessageDeleted": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_deleted_event",
      "WSMessageStatus": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_status_event",
      "WSStatusError": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_status_event.properties.error",
      "WSReaction": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_reaction_event",
      "WSReactionAggregate": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.message_reaction_event.properties.reactions.items",
      "WSTyping": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.typing_event",
      "WSThread": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.thread_created_event",
      "ThreadTagsChangedPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.thread_tags_changed_event",
      "MemberEventPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.member_added_event",
      "InteractiveCallbackPayload": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.interactive_callback",
      "RawPeer": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.interactive_callback.properties.reacted_by",
      "WSVariables": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.variable_set_event",
      "VariableEntry": "$ref:$.channels.ws.messages.ServerEvent.payload.properties.payload.properties.variable_set_event.properties.variables.additionalProperties",
      "AuthRequestPayload": "$ref:$.channels.ws.messages.AuthRequest.payload",
      "ClientCommandPayload": "$ref:$.channels.ws.messages.ClientCommand.payload"
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
  