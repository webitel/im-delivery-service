
class WsRecipient {
  private _id?: string;
  private _reservedName?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    reservedName?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._reservedName = input.reservedName;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get reservedName(): string | undefined { return this._reservedName; }
  set reservedName(reservedName: string | undefined) { this._reservedName = reservedName; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsRecipient;
