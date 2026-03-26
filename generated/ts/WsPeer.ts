
class WsPeer {
  private _sub?: string;
  private _iss?: string;
  private _reservedName?: string;
  private _reservedType?: string;
  private _isBot?: boolean;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    sub?: string,
    iss?: string,
    reservedName?: string,
    reservedType?: string,
    isBot?: boolean,
    additionalProperties?: Map<string, any>,
  }) {
    this._sub = input.sub;
    this._iss = input.iss;
    this._reservedName = input.reservedName;
    this._reservedType = input.reservedType;
    this._isBot = input.isBot;
    this._additionalProperties = input.additionalProperties;
  }

  get sub(): string | undefined { return this._sub; }
  set sub(sub: string | undefined) { this._sub = sub; }

  get iss(): string | undefined { return this._iss; }
  set iss(iss: string | undefined) { this._iss = iss; }

  get reservedName(): string | undefined { return this._reservedName; }
  set reservedName(reservedName: string | undefined) { this._reservedName = reservedName; }

  get reservedType(): string | undefined { return this._reservedType; }
  set reservedType(reservedType: string | undefined) { this._reservedType = reservedType; }

  get isBot(): boolean | undefined { return this._isBot; }
  set isBot(isBot: boolean | undefined) { this._isBot = isBot; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsPeer;
