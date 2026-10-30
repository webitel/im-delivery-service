
class WsSystem {
  private _reservedType?: string;
  private _metadata?: Map<string, any>;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    reservedType?: string,
    metadata?: Map<string, any>,
    additionalProperties?: Map<string, any>,
  }) {
    this._reservedType = input.reservedType;
    this._metadata = input.metadata;
    this._additionalProperties = input.additionalProperties;
  }

  get reservedType(): string | undefined { return this._reservedType; }
  set reservedType(reservedType: string | undefined) { this._reservedType = reservedType; }

  get metadata(): Map<string, any> | undefined { return this._metadata; }
  set metadata(metadata: Map<string, any> | undefined) { this._metadata = metadata; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsSystem;
