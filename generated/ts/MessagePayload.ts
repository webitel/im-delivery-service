import WsMessage from './WsMessage';
class MessagePayload {
  private _message?: WsMessage;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    message?: WsMessage,
    additionalProperties?: Map<string, any>,
  }) {
    this._message = input.message;
    this._additionalProperties = input.additionalProperties;
  }

  get message(): WsMessage | undefined { return this._message; }
  set message(message: WsMessage | undefined) { this._message = message; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default MessagePayload;
