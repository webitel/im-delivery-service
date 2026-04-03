import WsPeer from './WsPeer';
import AnonymousSchema_20 from './AnonymousSchema_20';
class WsMessage {
  private _id?: string;
  private _sendId?: string;
  private _threadId?: string;
  private _sender?: WsPeer;
  private _to?: WsPeer;
  private _createdAt?: number;
  private _editedAt?: number;
  private _body?: string;
  private _reservedType?: AnonymousSchema_20;
  private _content?: Map<string, any>;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    sendId?: string,
    threadId?: string,
    sender?: WsPeer,
    to?: WsPeer,
    createdAt?: number,
    editedAt?: number,
    body?: string,
    reservedType?: AnonymousSchema_20,
    content?: Map<string, any>,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._sendId = input.sendId;
    this._threadId = input.threadId;
    this._sender = input.sender;
    this._to = input.to;
    this._createdAt = input.createdAt;
    this._editedAt = input.editedAt;
    this._body = input.body;
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

  get sender(): WsPeer | undefined { return this._sender; }
  set sender(sender: WsPeer | undefined) { this._sender = sender; }

  get to(): WsPeer | undefined { return this._to; }
  set to(to: WsPeer | undefined) { this._to = to; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get editedAt(): number | undefined { return this._editedAt; }
  set editedAt(editedAt: number | undefined) { this._editedAt = editedAt; }

  get body(): string | undefined { return this._body; }
  set body(body: string | undefined) { this._body = body; }

  get reservedType(): AnonymousSchema_20 | undefined { return this._reservedType; }
  set reservedType(reservedType: AnonymousSchema_20 | undefined) { this._reservedType = reservedType; }

  get content(): Map<string, any> | undefined { return this._content; }
  set content(content: Map<string, any> | undefined) { this._content = content; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsMessage;
