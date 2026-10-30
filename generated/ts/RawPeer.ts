import PeerKind from './PeerKind';
class RawPeer {
  private _id?: string;
  private _reservedType?: PeerKind;
  private _sub?: string;
  private _issuer?: string;
  private _reservedName?: string;
  private _username?: string;
  private _isBot?: boolean;
  private _memberId?: string;
  private _role?: number;
  private _domainId?: number;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    reservedType?: PeerKind,
    sub?: string,
    issuer?: string,
    reservedName?: string,
    username?: string,
    isBot?: boolean,
    memberId?: string,
    role?: number,
    domainId?: number,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._reservedType = input.reservedType;
    this._sub = input.sub;
    this._issuer = input.issuer;
    this._reservedName = input.reservedName;
    this._username = input.username;
    this._isBot = input.isBot;
    this._memberId = input.memberId;
    this._role = input.role;
    this._domainId = input.domainId;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get reservedType(): PeerKind | undefined { return this._reservedType; }
  set reservedType(reservedType: PeerKind | undefined) { this._reservedType = reservedType; }

  get sub(): string | undefined { return this._sub; }
  set sub(sub: string | undefined) { this._sub = sub; }

  get issuer(): string | undefined { return this._issuer; }
  set issuer(issuer: string | undefined) { this._issuer = issuer; }

  get reservedName(): string | undefined { return this._reservedName; }
  set reservedName(reservedName: string | undefined) { this._reservedName = reservedName; }

  get username(): string | undefined { return this._username; }
  set username(username: string | undefined) { this._username = username; }

  get isBot(): boolean | undefined { return this._isBot; }
  set isBot(isBot: boolean | undefined) { this._isBot = isBot; }

  get memberId(): string | undefined { return this._memberId; }
  set memberId(memberId: string | undefined) { this._memberId = memberId; }

  get role(): number | undefined { return this._role; }
  set role(role: number | undefined) { this._role = role; }

  get domainId(): number | undefined { return this._domainId; }
  set domainId(domainId: number | undefined) { this._domainId = domainId; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default RawPeer;
