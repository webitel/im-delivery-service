import WsPeer from './WsPeer';
class WsMessageDeleted {
  private _id?: string;
  private _threadId?: string;
  private _deletedBy?: WsPeer;
  private _createdAt?: number;
  private _deletedAt?: number;
  private _updatesCursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    threadId?: string,
    deletedBy?: WsPeer,
    createdAt?: number,
    deletedAt?: number,
    updatesCursor?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._threadId = input.threadId;
    this._deletedBy = input.deletedBy;
    this._createdAt = input.createdAt;
    this._deletedAt = input.deletedAt;
    this._updatesCursor = input.updatesCursor;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get deletedBy(): WsPeer | undefined { return this._deletedBy; }
  set deletedBy(deletedBy: WsPeer | undefined) { this._deletedBy = deletedBy; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get deletedAt(): number | undefined { return this._deletedAt; }
  set deletedAt(deletedAt: number | undefined) { this._deletedAt = deletedAt; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsMessageDeleted;
