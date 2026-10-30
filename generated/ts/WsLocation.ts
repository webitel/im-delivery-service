
class WsLocation {
  private _latitude: number;
  private _longitude: number;
  private _address?: string;
  private _reservedName?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    latitude: number,
    longitude: number,
    address?: string,
    reservedName?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._latitude = input.latitude;
    this._longitude = input.longitude;
    this._address = input.address;
    this._reservedName = input.reservedName;
    this._additionalProperties = input.additionalProperties;
  }

  get latitude(): number { return this._latitude; }
  set latitude(latitude: number) { this._latitude = latitude; }

  get longitude(): number { return this._longitude; }
  set longitude(longitude: number) { this._longitude = longitude; }

  get address(): string | undefined { return this._address; }
  set address(address: string | undefined) { this._address = address; }

  get reservedName(): string | undefined { return this._reservedName; }
  set reservedName(reservedName: string | undefined) { this._reservedName = reservedName; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsLocation;
