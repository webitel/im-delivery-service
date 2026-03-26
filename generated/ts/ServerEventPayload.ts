import EventPriority from './EventPriority';
import EventPayload from './EventPayload';
class ServerEventPayload {
  private _id?: string;
  private _createdAt?: number;
  private _priority?: EventPriority;
  private _payload?: EventPayload;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    id?: string,
    createdAt?: number,
    priority?: EventPriority,
    payload?: EventPayload,
    additionalProperties?: Map<string, any>,
  }) {
    this._id = input.id;
    this._createdAt = input.createdAt;
    this._priority = input.priority;
    this._payload = input.payload;
    this._additionalProperties = input.additionalProperties;
  }

  get id(): string | undefined { return this._id; }
  set id(id: string | undefined) { this._id = id; }

  get createdAt(): number | undefined { return this._createdAt; }
  set createdAt(createdAt: number | undefined) { this._createdAt = createdAt; }

  get priority(): EventPriority | undefined { return this._priority; }
  set priority(priority: EventPriority | undefined) { this._priority = priority; }

  get payload(): EventPayload | undefined { return this._payload; }
  set payload(payload: EventPayload | undefined) { this._payload = payload; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default ServerEventPayload;
