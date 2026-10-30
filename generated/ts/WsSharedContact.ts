
class WsSharedContact {
  private _reservedName?: string;
  private _phone?: string;
  private _email?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    reservedName?: string,
    phone?: string,
    email?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._reservedName = input.reservedName;
    this._phone = input.phone;
    this._email = input.email;
    this._additionalProperties = input.additionalProperties;
  }

  get reservedName(): string | undefined { return this._reservedName; }
  set reservedName(reservedName: string | undefined) { this._reservedName = reservedName; }

  get phone(): string | undefined { return this._phone; }
  set phone(phone: string | undefined) { this._phone = phone; }

  get email(): string | undefined { return this._email; }
  set email(email: string | undefined) { this._email = email; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsSharedContact;
