import WsPeer from './WsPeer';
import MessageType from './MessageType';
import AttachmentKind from './AttachmentKind';
class WsReplyTo {
  private _messageId?: string;
  private _sender?: WsPeer;
  private _senderId?: string;
  private _reservedType?: MessageType;
  private _body?: string;
  private _createdAt?: number;
  private _attachmentKind?: AttachmentKind;
  private _attachmentName?: string;
  private _attachmentMime?: string;
  private _attachmentAddress?: string;
  private _isDeleted?: boolean;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    messageId?: string,
    sender?: WsPeer,
    senderId?: string,
    reservedType?: MessageType,
    body?: string,
    createdAt?: number,
    attachmentKind?: AttachmentKind,
    attachmentName?: string,
    attachmentMime?: string,
    attachmentAddress?: string,
    isDeleted?: boolean,
    additionalProperties?: Map<string, any>,
  }) {
    this._messageId = input.messageId;
    this._sender = input.sender;
    this._senderId = input.senderId;
    this._reservedType = input.reservedType;
    this._body = input.body;
    this._createdAt = input.createdAt;
    this._attachmentKind = input.attachmentKind;
    this._attachmentName = input.attachmentName;
    this._attachmentMime = input.attachmentMime;
    this._attachmentAddress = input.attachmentAddress;
    this._isDeleted = input.isDeleted;
    this._additionalProperties = input.additionalProperties;
  }

  get messageId(): string | undefined { return this._messageId; }
  set messageId(messageId: string | undefined) { this._messageId = messageId; }

  get sender(): WsPeer | undefined { return this._sender; }
  set sender(sender: WsPeer | undefined) { this._sender = sender; }

  get senderId(): string | undefined { return this._senderId; }
  set senderId(senderId: string | undefined) { this._senderId = senderId; }

  get reservedType(): MessageType | undefined { return this._reservedType; }
  set reservedType(reservedType: MessageType | undefined) { this._reservedType = reservedType; }

  get body(): string | undefined { return this._body; }
  set body(body: string | undefined) { this._body = body; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get attachmentKind(): AttachmentKind | undefined { return this._attachmentKind; }
  set attachmentKind(attachmentKind: AttachmentKind | undefined) { this._attachmentKind = attachmentKind; }

  get attachmentName(): string | undefined { return this._attachmentName; }
  set attachmentName(attachmentName: string | undefined) { this._attachmentName = attachmentName; }

  get attachmentMime(): string | undefined { return this._attachmentMime; }
  set attachmentMime(attachmentMime: string | undefined) { this._attachmentMime = attachmentMime; }

  get attachmentAddress(): string | undefined { return this._attachmentAddress; }
  set attachmentAddress(attachmentAddress: string | undefined) { this._attachmentAddress = attachmentAddress; }

  get isDeleted(): boolean | undefined { return this._isDeleted; }
  set isDeleted(isDeleted: boolean | undefined) { this._isDeleted = isDeleted; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsReplyTo;
