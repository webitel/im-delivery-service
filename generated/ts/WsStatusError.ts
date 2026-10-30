
class WsStatusError {
  private _code?: string;
  private _message?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    code?: string,
    message?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._code = input.code;
    this._message = input.message;
    this._additionalProperties = input.additionalProperties;
  }

  get code(): string | undefined { return this._code; }
  set code(code: string | undefined) { this._code = code; }

  get message(): string | undefined { return this._message; }
  set message(message: string | undefined) { this._message = message; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsStatusError;
