import ConnectedPayload from './ConnectedPayload';
import DisconnectedPayload from './DisconnectedPayload';
import ResyncPayload from './ResyncPayload';
import WsMessage from './WsMessage';
import WsMessageEdited from './WsMessageEdited';
import WsMessageDeleted from './WsMessageDeleted';
import WsMessageStatus from './WsMessageStatus';
import WsReaction from './WsReaction';
import WsTyping from './WsTyping';
import WsThread from './WsThread';
import ThreadTagsChangedPayload from './ThreadTagsChangedPayload';
import MemberEventPayload from './MemberEventPayload';
import InteractiveCallbackPayload from './InteractiveCallbackPayload';
import WsVariables from './WsVariables';
import AckPayload from './AckPayload';
import ErrorPayload from './ErrorPayload';
import PingPayload from './PingPayload';
class EventPayload {
  private _connectedEvent?: ConnectedPayload;
  private _disconnectedEvent?: DisconnectedPayload;
  private _resyncEvent?: ResyncPayload;
  private _messageEvent?: WsMessage;
  private _messageEditedEvent?: WsMessageEdited;
  private _messageDeletedEvent?: WsMessageDeleted;
  private _messageStatusEvent?: WsMessageStatus;
  private _messageReactionEvent?: WsReaction;
  private _typingEvent?: WsTyping;
  private _threadCreatedEvent?: WsThread;
  private _threadTagsChangedEvent?: ThreadTagsChangedPayload;
  private _memberAddedEvent?: MemberEventPayload;
  private _memberLeftEvent?: MemberEventPayload;
  private _interactiveCallback?: InteractiveCallbackPayload;
  private _variableSetEvent?: WsVariables;
  private _variableFlushEvent?: WsVariables;
  private _ackEvent?: AckPayload;
  private _errorEvent?: ErrorPayload;
  private _pingEvent?: PingPayload;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    connectedEvent?: ConnectedPayload,
    disconnectedEvent?: DisconnectedPayload,
    resyncEvent?: ResyncPayload,
    messageEvent?: WsMessage,
    messageEditedEvent?: WsMessageEdited,
    messageDeletedEvent?: WsMessageDeleted,
    messageStatusEvent?: WsMessageStatus,
    messageReactionEvent?: WsReaction,
    typingEvent?: WsTyping,
    threadCreatedEvent?: WsThread,
    threadTagsChangedEvent?: ThreadTagsChangedPayload,
    memberAddedEvent?: MemberEventPayload,
    memberLeftEvent?: MemberEventPayload,
    interactiveCallback?: InteractiveCallbackPayload,
    variableSetEvent?: WsVariables,
    variableFlushEvent?: WsVariables,
    ackEvent?: AckPayload,
    errorEvent?: ErrorPayload,
    pingEvent?: PingPayload,
    additionalProperties?: Map<string, any>,
  }) {
    this._connectedEvent = input.connectedEvent;
    this._disconnectedEvent = input.disconnectedEvent;
    this._resyncEvent = input.resyncEvent;
    this._messageEvent = input.messageEvent;
    this._messageEditedEvent = input.messageEditedEvent;
    this._messageDeletedEvent = input.messageDeletedEvent;
    this._messageStatusEvent = input.messageStatusEvent;
    this._messageReactionEvent = input.messageReactionEvent;
    this._typingEvent = input.typingEvent;
    this._threadCreatedEvent = input.threadCreatedEvent;
    this._threadTagsChangedEvent = input.threadTagsChangedEvent;
    this._memberAddedEvent = input.memberAddedEvent;
    this._memberLeftEvent = input.memberLeftEvent;
    this._interactiveCallback = input.interactiveCallback;
    this._variableSetEvent = input.variableSetEvent;
    this._variableFlushEvent = input.variableFlushEvent;
    this._ackEvent = input.ackEvent;
    this._errorEvent = input.errorEvent;
    this._pingEvent = input.pingEvent;
    this._additionalProperties = input.additionalProperties;
  }

  get connectedEvent(): ConnectedPayload | undefined { return this._connectedEvent; }
  set connectedEvent(connectedEvent: ConnectedPayload | undefined) { this._connectedEvent = connectedEvent; }

  get disconnectedEvent(): DisconnectedPayload | undefined { return this._disconnectedEvent; }
  set disconnectedEvent(disconnectedEvent: DisconnectedPayload | undefined) { this._disconnectedEvent = disconnectedEvent; }

  get resyncEvent(): ResyncPayload | undefined { return this._resyncEvent; }
  set resyncEvent(resyncEvent: ResyncPayload | undefined) { this._resyncEvent = resyncEvent; }

  get messageEvent(): WsMessage | undefined { return this._messageEvent; }
  set messageEvent(messageEvent: WsMessage | undefined) { this._messageEvent = messageEvent; }

  get messageEditedEvent(): WsMessageEdited | undefined { return this._messageEditedEvent; }
  set messageEditedEvent(messageEditedEvent: WsMessageEdited | undefined) { this._messageEditedEvent = messageEditedEvent; }

  get messageDeletedEvent(): WsMessageDeleted | undefined { return this._messageDeletedEvent; }
  set messageDeletedEvent(messageDeletedEvent: WsMessageDeleted | undefined) { this._messageDeletedEvent = messageDeletedEvent; }

  get messageStatusEvent(): WsMessageStatus | undefined { return this._messageStatusEvent; }
  set messageStatusEvent(messageStatusEvent: WsMessageStatus | undefined) { this._messageStatusEvent = messageStatusEvent; }

  get messageReactionEvent(): WsReaction | undefined { return this._messageReactionEvent; }
  set messageReactionEvent(messageReactionEvent: WsReaction | undefined) { this._messageReactionEvent = messageReactionEvent; }

  get typingEvent(): WsTyping | undefined { return this._typingEvent; }
  set typingEvent(typingEvent: WsTyping | undefined) { this._typingEvent = typingEvent; }

  get threadCreatedEvent(): WsThread | undefined { return this._threadCreatedEvent; }
  set threadCreatedEvent(threadCreatedEvent: WsThread | undefined) { this._threadCreatedEvent = threadCreatedEvent; }

  get threadTagsChangedEvent(): ThreadTagsChangedPayload | undefined { return this._threadTagsChangedEvent; }
  set threadTagsChangedEvent(threadTagsChangedEvent: ThreadTagsChangedPayload | undefined) { this._threadTagsChangedEvent = threadTagsChangedEvent; }

  get memberAddedEvent(): MemberEventPayload | undefined { return this._memberAddedEvent; }
  set memberAddedEvent(memberAddedEvent: MemberEventPayload | undefined) { this._memberAddedEvent = memberAddedEvent; }

  get memberLeftEvent(): MemberEventPayload | undefined { return this._memberLeftEvent; }
  set memberLeftEvent(memberLeftEvent: MemberEventPayload | undefined) { this._memberLeftEvent = memberLeftEvent; }

  get interactiveCallback(): InteractiveCallbackPayload | undefined { return this._interactiveCallback; }
  set interactiveCallback(interactiveCallback: InteractiveCallbackPayload | undefined) { this._interactiveCallback = interactiveCallback; }

  get variableSetEvent(): WsVariables | undefined { return this._variableSetEvent; }
  set variableSetEvent(variableSetEvent: WsVariables | undefined) { this._variableSetEvent = variableSetEvent; }

  get variableFlushEvent(): WsVariables | undefined { return this._variableFlushEvent; }
  set variableFlushEvent(variableFlushEvent: WsVariables | undefined) { this._variableFlushEvent = variableFlushEvent; }

  get ackEvent(): AckPayload | undefined { return this._ackEvent; }
  set ackEvent(ackEvent: AckPayload | undefined) { this._ackEvent = ackEvent; }

  get errorEvent(): ErrorPayload | undefined { return this._errorEvent; }
  set errorEvent(errorEvent: ErrorPayload | undefined) { this._errorEvent = errorEvent; }

  get pingEvent(): PingPayload | undefined { return this._pingEvent; }
  set pingEvent(pingEvent: PingPayload | undefined) { this._pingEvent = pingEvent; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default EventPayload;
