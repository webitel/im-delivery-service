import WsPeer from './WsPeer';
class WsMessage {
  private _id?: string;
  private _sendId?: string;
  private _threadId?: string;
  private _reservedFrom?: WsPeer;
  private _to?: WsPeer;
  private _createdAt?: number;
  private _editedAt?: number;
  private _reservedText?: string;
  private _reservedType?: string;
  private _content?: Map<string, any>;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    sendId?: string,
    threadId?: string,
    reservedFrom?: WsPeer,
    to?: WsPeer,
    createdAt?: number,
    editedAt?: number,
    reservedText?: string,
    reservedType?: string,
    content?: Map<string, any>,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._sendId = input.sendId;
    this._threadId = input.threadId;
    this._reservedFrom = input.reservedFrom;
    this._to = input.to;
    this._createdAt = input.createdAt;
    this._editedAt = input.editedAt;
    this._reservedText = input.reservedText;
    this._reservedType = input.reservedType;
    this._content = input.content;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get sendId(): string | undefined { return this._sendId; }
  set sendId(sendId: string | undefined) { this._sendId = sendId; }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get reservedFrom(): WsPeer | undefined { return this._reservedFrom; }
  set reservedFrom(reservedFrom: WsPeer | undefined) { this._reservedFrom = reservedFrom; }

  get to(): WsPeer | undefined { return this._to; }
  set to(to: WsPeer | undefined) { this._to = to; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get editedAt(): number | undefined { return this._editedAt; }
  set editedAt(editedAt: number | undefined) { this._editedAt = editedAt; }

  get reservedText(): string | undefined { return this._reservedText; }
  set reservedText(reservedText: string | undefined) { this._reservedText = reservedText; }

  get reservedType(): string | undefined { return this._reservedType; }
  set reservedType(reservedType: string | undefined) { this._reservedType = reservedType; }

  get content(): Map<string, any> | undefined { return this._content; }
  set content(content: Map<string, any> | undefined) { this._content = content; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsMessage;
