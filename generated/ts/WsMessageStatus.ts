import DeliveryStatus from './DeliveryStatus';
import WsPeer from './WsPeer';
import StatusVia from './StatusVia';
import WsStatusError from './WsStatusError';
class WsMessageStatus {
  private _threadId?: string;
  private _reservedStatus?: DeliveryStatus;
  private _member?: WsPeer;
  private _messageId?: string;
  private _upToSeq?: number;
  private _unreadCount?: number;
  private _via?: StatusVia;
  private _error?: WsStatusError;
  private _occurredAt?: number;
  private _updatesCursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    threadId?: string,
    reservedStatus?: DeliveryStatus,
    member?: WsPeer,
    messageId?: string,
    upToSeq?: number,
    unreadCount?: number,
    via?: StatusVia,
    error?: WsStatusError,
    occurredAt?: number,
    updatesCursor?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._threadId = input.threadId;
    this._reservedStatus = input.reservedStatus;
    this._member = input.member;
    this._messageId = input.messageId;
    this._upToSeq = input.upToSeq;
    this._unreadCount = input.unreadCount;
    this._via = input.via;
    this._error = input.error;
    this._occurredAt = input.occurredAt;
    this._updatesCursor = input.updatesCursor;
    this._additionalProperties = input.additionalProperties;
  }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get reservedStatus(): DeliveryStatus | undefined { return this._reservedStatus; }
  set reservedStatus(reservedStatus: DeliveryStatus | undefined) { this._reservedStatus = reservedStatus; }

  get member(): WsPeer | undefined { return this._member; }
  set member(member: WsPeer | undefined) { this._member = member; }

  get messageId(): string | undefined { return this._messageId; }
  set messageId(messageId: string | undefined) { this._messageId = messageId; }

  get upToSeq(): number | undefined { return this._upToSeq; }
  set upToSeq(upToSeq: number | undefined) { this._upToSeq = upToSeq; }

  get unreadCount(): number | undefined { return this._unreadCount; }
  set unreadCount(unreadCount: number | undefined) { this._unreadCount = unreadCount; }

  get via(): StatusVia | undefined { return this._via; }
  set via(via: StatusVia | undefined) { this._via = via; }

  get error(): WsStatusError | undefined { return this._error; }
  set error(error: WsStatusError | undefined) { this._error = error; }

  get occurredAt(): number | undefined { return this._occurredAt; }
  set occurredAt(occurredAt: number | undefined) { this._occurredAt = occurredAt; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsMessageStatus;
