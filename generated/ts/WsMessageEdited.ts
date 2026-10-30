import WsPeer from './WsPeer';
import MessageType from './MessageType';
class WsMessageEdited {
  private _id?: string;
  private _threadId?: string;
  private _editedBy?: WsPeer;
  private _body?: string;
  private _reservedType?: MessageType;
  private _version?: number;
  private _createdAt?: number;
  private _editedAt?: number;
  private _metadata?: Map<string, any>;
  private _updatesCursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    threadId?: string,
    editedBy?: WsPeer,
    body?: string,
    reservedType?: MessageType,
    version?: number,
    createdAt?: number,
    editedAt?: number,
    metadata?: Map<string, any>,
    updatesCursor?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._threadId = input.threadId;
    this._editedBy = input.editedBy;
    this._body = input.body;
    this._reservedType = input.reservedType;
    this._version = input.version;
    this._createdAt = input.createdAt;
    this._editedAt = input.editedAt;
    this._metadata = input.metadata;
    this._updatesCursor = input.updatesCursor;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get editedBy(): WsPeer | undefined { return this._editedBy; }
  set editedBy(editedBy: WsPeer | undefined) { this._editedBy = editedBy; }

  get body(): string | undefined { return this._body; }
  set body(body: string | undefined) { this._body = body; }

  get reservedType(): MessageType | undefined { return this._reservedType; }
  set reservedType(reservedType: MessageType | undefined) { this._reservedType = reservedType; }

  get version(): number | undefined { return this._version; }
  set version(version: number | undefined) { this._version = version; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get editedAt(): number | undefined { return this._editedAt; }
  set editedAt(editedAt: number | undefined) { this._editedAt = editedAt; }

  get metadata(): Map<string, any> | undefined { return this._metadata; }
  set metadata(metadata: Map<string, any> | undefined) { this._metadata = metadata; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsMessageEdited;
