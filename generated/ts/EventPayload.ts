import ConnectedPayload from './ConnectedPayload';
import DisconnectedPayload from './DisconnectedPayload';
import WsMessage from './WsMessage';
import WsThread from './WsThread';
import AckPayload from './AckPayload';
import ErrorPayload from './ErrorPayload';
import PingPayload from './PingPayload';
class EventPayload {
  private _connectedEvent?: ConnectedPayload;
  private _disconnectedEvent?: DisconnectedPayload;
  private _messageEvent?: WsMessage;
  private _threadCreatedEvent?: WsThread;
  private _ackEvent?: AckPayload;
  private _errorEvent?: ErrorPayload;
  private _pingEvent?: PingPayload;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    connectedEvent?: ConnectedPayload,
    disconnectedEvent?: DisconnectedPayload,
    messageEvent?: WsMessage,
    threadCreatedEvent?: WsThread,
    ackEvent?: AckPayload,
    errorEvent?: ErrorPayload,
    pingEvent?: PingPayload,
    additionalProperties?: Map<string, any>,
  }) {
    this._connectedEvent = input.connectedEvent;
    this._disconnectedEvent = input.disconnectedEvent;
    this._messageEvent = input.messageEvent;
    this._threadCreatedEvent = input.threadCreatedEvent;
    this._ackEvent = input.ackEvent;
    this._errorEvent = input.errorEvent;
    this._pingEvent = input.pingEvent;
    this._additionalProperties = input.additionalProperties;
  }

  get connectedEvent(): ConnectedPayload | undefined { return this._connectedEvent; }
  set connectedEvent(connectedEvent: ConnectedPayload | undefined) { this._connectedEvent = connectedEvent; }

  get disconnectedEvent(): DisconnectedPayload | undefined { return this._disconnectedEvent; }
  set disconnectedEvent(disconnectedEvent: DisconnectedPayload | undefined) { this._disconnectedEvent = disconnectedEvent; }

  get messageEvent(): WsMessage | undefined { return this._messageEvent; }
  set messageEvent(messageEvent: WsMessage | undefined) { this._messageEvent = messageEvent; }

  get threadCreatedEvent(): WsThread | undefined { return this._threadCreatedEvent; }
  set threadCreatedEvent(threadCreatedEvent: WsThread | undefined) { this._threadCreatedEvent = threadCreatedEvent; }

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
