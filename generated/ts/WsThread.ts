import WsPeer from './WsPeer';
class WsThread {
  private _id?: string;
  private _domainId?: number;
  private _createdAt?: number;
  private _subject?: string;
  private _reservedType?: string;
  private _members?: WsPeer[];
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    domainId?: number,
    createdAt?: number,
    subject?: string,
    reservedType?: string,
    members?: WsPeer[],
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._domainId = input.domainId;
    this._createdAt = input.createdAt;
    this._subject = input.subject;
    this._reservedType = input.reservedType;
    this._members = input.members;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get domainId(): number | undefined { return this._domainId; }
  set domainId(domainId: number | undefined) { this._domainId = domainId; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get subject(): string | undefined { return this._subject; }
  set subject(subject: string | undefined) { this._subject = subject; }

  get reservedType(): string | undefined { return this._reservedType; }
  set reservedType(reservedType: string | undefined) { this._reservedType = reservedType; }

  get members(): WsPeer[] | undefined { return this._members; }
  set members(members: WsPeer[] | undefined) { this._members = members; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsThread;
