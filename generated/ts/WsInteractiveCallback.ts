import WsPeer from './WsPeer';
class WsInteractiveCallback {
  private _threadId?: string;
  private _messageId?: string;
  private _buttonCode?: string;
  private _callbackData?: string;
  private _reactedAt?: number;
  private _reactedBy?: WsPeer;
  private _receiver?: WsPeer;
  private _updatesCursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    threadId?: string,
    messageId?: string,
    buttonCode?: string,
    callbackData?: string,
    reactedAt?: number,
    reactedBy?: WsPeer,
    receiver?: WsPeer,
    updatesCursor?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._threadId = input.threadId;
    this._messageId = input.messageId;
    this._buttonCode = input.buttonCode;
    this._callbackData = input.callbackData;
    this._reactedAt = input.reactedAt;
    this._reactedBy = input.reactedBy;
    this._receiver = input.receiver;
    this._updatesCursor = input.updatesCursor;
    this._additionalProperties = input.additionalProperties;
  }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get messageId(): string | undefined { return this._messageId; }
  set messageId(messageId: string | undefined) { this._messageId = messageId; }

  get buttonCode(): string | undefined { return this._buttonCode; }
  set buttonCode(buttonCode: string | undefined) { this._buttonCode = buttonCode; }

  get callbackData(): string | undefined { return this._callbackData; }
  set callbackData(callbackData: string | undefined) { this._callbackData = callbackData; }

  get reactedAt(): number | undefined { return this._reactedAt; }
  set reactedAt(reactedAt: number | undefined) { this._reactedAt = reactedAt; }

  get reactedBy(): WsPeer | undefined { return this._reactedBy; }
  set reactedBy(reactedBy: WsPeer | undefined) { this._reactedBy = reactedBy; }

  get receiver(): WsPeer | undefined { return this._receiver; }
  set receiver(receiver: WsPeer | undefined) { this._receiver = receiver; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsInteractiveCallback;
