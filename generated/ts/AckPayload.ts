
class AckPayload {
  private _id?: string;
  private _reservedStatus?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    reservedStatus?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._reservedStatus = input.reservedStatus;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get reservedStatus(): string | undefined { return this._reservedStatus; }
  set reservedStatus(reservedStatus: string | undefined) { this._reservedStatus = reservedStatus; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default AckPayload;
