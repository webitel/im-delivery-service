import ForwardOriginKind from './ForwardOriginKind';
class WsForwardOrigin {
  private _kind?: ForwardOriginKind;
  private _senderId?: string;
  private _senderName?: string;
  private _originalSentAt?: number;
  private _sourceMessageId?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    kind?: ForwardOriginKind,
    senderId?: string,
    senderName?: string,
    originalSentAt?: number,
    sourceMessageId?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._kind = input.kind;
    this._senderId = input.senderId;
    this._senderName = input.senderName;
    this._originalSentAt = input.originalSentAt;
    this._sourceMessageId = input.sourceMessageId;
    this._additionalProperties = input.additionalProperties;
  }

  get kind(): ForwardOriginKind | undefined { return this._kind; }
  set kind(kind: ForwardOriginKind | undefined) { this._kind = kind; }

  get senderId(): string | undefined { return this._senderId; }
  set senderId(senderId: string | undefined) { this._senderId = senderId; }

  get senderName(): string | undefined { return this._senderName; }
  set senderName(senderName: string | undefined) { this._senderName = senderName; }

  get originalSentAt(): number | undefined { return this._originalSentAt; }
  set originalSentAt(originalSentAt: number | undefined) { this._originalSentAt = originalSentAt; }

  get sourceMessageId(): string | undefined { return this._sourceMessageId; }
  set sourceMessageId(sourceMessageId: string | undefined) { this._sourceMessageId = sourceMessageId; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsForwardOrigin;
