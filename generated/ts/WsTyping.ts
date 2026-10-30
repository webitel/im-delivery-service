import WsPeer from './WsPeer';
class WsTyping {
  private _threadId?: string;
  private _timeoutMs?: number;
  private _reservedFrom?: WsPeer;
  private _previewText?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    threadId?: string,
    timeoutMs?: number,
    reservedFrom?: WsPeer,
    previewText?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._threadId = input.threadId;
    this._timeoutMs = input.timeoutMs;
    this._reservedFrom = input.reservedFrom;
    this._previewText = input.previewText;
    this._additionalProperties = input.additionalProperties;
  }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get timeoutMs(): number | undefined { return this._timeoutMs; }
  set timeoutMs(timeoutMs: number | undefined) { this._timeoutMs = timeoutMs; }

  get reservedFrom(): WsPeer | undefined { return this._reservedFrom; }
  set reservedFrom(reservedFrom: WsPeer | undefined) { this._reservedFrom = reservedFrom; }

  get previewText(): string | undefined { return this._previewText; }
  set previewText(previewText: string | undefined) { this._previewText = previewText; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsTyping;
