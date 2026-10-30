import ClientCommandType from './ClientCommandType';
class ClientCommandPayload {
  private _reservedType: ClientCommandType;
  private _eventId?: string;
  private _threadId?: string;
  private _seq?: number;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    reservedType: ClientCommandType,
    eventId?: string,
    threadId?: string,
    seq?: number,
    additionalProperties?: Map<string, any>,
  }) {
    this._reservedType = input.reservedType;
    this._eventId = input.eventId;
    this._threadId = input.threadId;
    this._seq = input.seq;
    this._additionalProperties = input.additionalProperties;
  }

  get reservedType(): ClientCommandType { return this._reservedType; }
  set reservedType(reservedType: ClientCommandType) { this._reservedType = reservedType; }

  get eventId(): string | undefined { return this._eventId; }
  set eventId(eventId: string | undefined) { this._eventId = eventId; }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get seq(): number | undefined { return this._seq; }
  set seq(seq: number | undefined) { this._seq = seq; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default ClientCommandPayload;
