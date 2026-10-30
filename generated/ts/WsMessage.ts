import WsPeer from './WsPeer';
import MessageType from './MessageType';
import WsImage from './WsImage';
import WsDocument from './WsDocument';
import WsSharedContact from './WsSharedContact';
import WsLocation from './WsLocation';
import WsSystem from './WsSystem';
import WsReplyTo from './WsReplyTo';
import WsForwardOrigin from './WsForwardOrigin';
class WsMessage {
  private _id?: string;
  private _sendId?: string;
  private _threadId?: string;
  private _sender?: WsPeer;
  private _to?: WsPeer[];
  private _createdAt?: number;
  private _editedAt?: number;
  private _body?: string;
  private _reservedType?: MessageType;
  private _reservedImages?: WsImage[];
  private _documents?: WsDocument[];
  private _interactive?: Map<string, any>;
  private _contact?: WsSharedContact;
  private _reservedLocation?: WsLocation;
  private _system?: WsSystem;
  private _metadata?: Map<string, any>;
  private _replyTo?: WsReplyTo;
  private _forwardOrigin?: WsForwardOrigin;
  private _seq?: number;
  private _updatesCursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    sendId?: string,
    threadId?: string,
    sender?: WsPeer,
    to?: WsPeer[],
    createdAt?: number,
    editedAt?: number,
    body?: string,
    reservedType?: MessageType,
    reservedImages?: WsImage[],
    documents?: WsDocument[],
    interactive?: Map<string, any>,
    contact?: WsSharedContact,
    reservedLocation?: WsLocation,
    system?: WsSystem,
    metadata?: Map<string, any>,
    replyTo?: WsReplyTo,
    forwardOrigin?: WsForwardOrigin,
    seq?: number,
    updatesCursor?: string,
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
    this._reservedImages = input.reservedImages;
    this._documents = input.documents;
    this._interactive = input.interactive;
    this._contact = input.contact;
    this._reservedLocation = input.reservedLocation;
    this._system = input.system;
    this._metadata = input.metadata;
    this._replyTo = input.replyTo;
    this._forwardOrigin = input.forwardOrigin;
    this._seq = input.seq;
    this._updatesCursor = input.updatesCursor;
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

  get to(): WsPeer[] | undefined { return this._to; }
  set to(to: WsPeer[] | undefined) { this._to = to; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get editedAt(): number | undefined { return this._editedAt; }
  set editedAt(editedAt: number | undefined) { this._editedAt = editedAt; }

  get body(): string | undefined { return this._body; }
  set body(body: string | undefined) { this._body = body; }

  get reservedType(): MessageType | undefined { return this._reservedType; }
  set reservedType(reservedType: MessageType | undefined) { this._reservedType = reservedType; }

  get reservedImages(): WsImage[] | undefined { return this._reservedImages; }
  set reservedImages(reservedImages: WsImage[] | undefined) { this._reservedImages = reservedImages; }

  get documents(): WsDocument[] | undefined { return this._documents; }
  set documents(documents: WsDocument[] | undefined) { this._documents = documents; }

  get interactive(): Map<string, any> | undefined { return this._interactive; }
  set interactive(interactive: Map<string, any> | undefined) { this._interactive = interactive; }

  get contact(): WsSharedContact | undefined { return this._contact; }
  set contact(contact: WsSharedContact | undefined) { this._contact = contact; }

  get reservedLocation(): WsLocation | undefined { return this._reservedLocation; }
  set reservedLocation(reservedLocation: WsLocation | undefined) { this._reservedLocation = reservedLocation; }

  get system(): WsSystem | undefined { return this._system; }
  set system(system: WsSystem | undefined) { this._system = system; }

  get metadata(): Map<string, any> | undefined { return this._metadata; }
  set metadata(metadata: Map<string, any> | undefined) { this._metadata = metadata; }

  get replyTo(): WsReplyTo | undefined { return this._replyTo; }
  set replyTo(replyTo: WsReplyTo | undefined) { this._replyTo = replyTo; }

  get forwardOrigin(): WsForwardOrigin | undefined { return this._forwardOrigin; }
  set forwardOrigin(forwardOrigin: WsForwardOrigin | undefined) { this._forwardOrigin = forwardOrigin; }

  get seq(): number | undefined { return this._seq; }
  set seq(seq: number | undefined) { this._seq = seq; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsMessage;
