import RawPeer from './RawPeer';
class InteractiveCallbackPayload {
  private _reactedBy?: RawPeer;
  private _inReplyTo?: string;
  private _buttonCode?: string;
  private _callbackData?: string;
  private _reactedAt?: string;
  private _domainId?: number;
  private _threadId?: string;
  private _receiver?: RawPeer;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    reactedBy?: RawPeer,
    inReplyTo?: string,
    buttonCode?: string,
    callbackData?: string,
    reactedAt?: string,
    domainId?: number,
    threadId?: string,
    receiver?: RawPeer,
    additionalProperties?: Map<string, any>,
  }) {
    this._reactedBy = input.reactedBy;
    this._inReplyTo = input.inReplyTo;
    this._buttonCode = input.buttonCode;
    this._callbackData = input.callbackData;
    this._reactedAt = input.reactedAt;
    this._domainId = input.domainId;
    this._threadId = input.threadId;
    this._receiver = input.receiver;
    this._additionalProperties = input.additionalProperties;
  }

  get reactedBy(): RawPeer | undefined { return this._reactedBy; }
  set reactedBy(reactedBy: RawPeer | undefined) { this._reactedBy = reactedBy; }

  get inReplyTo(): string | undefined { return this._inReplyTo; }
  set inReplyTo(inReplyTo: string | undefined) { this._inReplyTo = inReplyTo; }

  get buttonCode(): string | undefined { return this._buttonCode; }
  set buttonCode(buttonCode: string | undefined) { this._buttonCode = buttonCode; }

  get callbackData(): string | undefined { return this._callbackData; }
  set callbackData(callbackData: string | undefined) { this._callbackData = callbackData; }

  get reactedAt(): string | undefined { return this._reactedAt; }
  set reactedAt(reactedAt: string | undefined) { this._reactedAt = reactedAt; }

  get domainId(): number | undefined { return this._domainId; }
  set domainId(domainId: number | undefined) { this._domainId = domainId; }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get receiver(): RawPeer | undefined { return this._receiver; }
  set receiver(receiver: RawPeer | undefined) { this._receiver = receiver; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default InteractiveCallbackPayload;
