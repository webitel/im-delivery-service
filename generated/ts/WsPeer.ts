import WsContact from './WsContact';
class WsPeer {
  private _id?: string;
  private _contact?: WsContact;
  private _role?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    contact?: WsContact,
    role?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._contact = input.contact;
    this._role = input.role;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get contact(): WsContact | undefined { return this._contact; }
  set contact(contact: WsContact | undefined) { this._contact = contact; }

  get role(): string | undefined { return this._role; }
  set role(role: string | undefined) { this._role = role; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsPeer;
