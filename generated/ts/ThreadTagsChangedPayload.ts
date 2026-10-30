
class ThreadTagsChangedPayload {
  private _threadId: string;
  private _tags: string[];
  private _updatesCursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    threadId: string,
    tags: string[],
    updatesCursor?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._threadId = input.threadId;
    this._tags = input.tags;
    this._updatesCursor = input.updatesCursor;
    this._additionalProperties = input.additionalProperties;
  }

  get threadId(): string { return this._threadId; }
  set threadId(threadId: string) { this._threadId = threadId; }

  get tags(): string[] { return this._tags; }
  set tags(tags: string[]) { this._tags = tags; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default ThreadTagsChangedPayload;
