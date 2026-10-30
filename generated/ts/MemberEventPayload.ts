import MemberAction from './MemberAction';
class MemberEventPayload {
  private _threadId?: string;
  private _contactId?: string;
  private _metadata?: Map<string, any>;
  private _updatesCursor?: string;
  private _action?: MemberAction;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    threadId?: string,
    contactId?: string,
    metadata?: Map<string, any>,
    updatesCursor?: string,
    action?: MemberAction,
    additionalProperties?: Map<string, any>,
  }) {
    this._threadId = input.threadId;
    this._contactId = input.contactId;
    this._metadata = input.metadata;
    this._updatesCursor = input.updatesCursor;
    this._action = input.action;
    this._additionalProperties = input.additionalProperties;
  }

  get threadId(): string | undefined { return this._threadId; }
  set threadId(threadId: string | undefined) { this._threadId = threadId; }

  get contactId(): string | undefined { return this._contactId; }
  set contactId(contactId: string | undefined) { this._contactId = contactId; }

  get metadata(): Map<string, any> | undefined { return this._metadata; }
  set metadata(metadata: Map<string, any> | undefined) { this._metadata = metadata; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get action(): MemberAction | undefined { return this._action; }
  set action(action: MemberAction | undefined) { this._action = action; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default MemberEventPayload;
