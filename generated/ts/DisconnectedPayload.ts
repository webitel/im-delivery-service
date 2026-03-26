
class DisconnectedPayload {
  private _reason?: string;
  private _code?: number;
  private _reservedStatus?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    reason?: string,
    code?: number,
    reservedStatus?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._reason = input.reason;
    this._code = input.code;
    this._reservedStatus = input.reservedStatus;
    this._additionalProperties = input.additionalProperties;
  }

  get reason(): string | undefined { return this._reason; }
  set reason(reason: string | undefined) { this._reason = reason; }

  get code(): number | undefined { return this._code; }
  set code(code: number | undefined) { this._code = code; }

  get reservedStatus(): string | undefined { return this._reservedStatus; }
  set reservedStatus(reservedStatus: string | undefined) { this._reservedStatus = reservedStatus; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default DisconnectedPayload;
