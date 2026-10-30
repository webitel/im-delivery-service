import WsReactionAggregate from './WsReactionAggregate';
class WsReaction {
  private _messageId: string;
  private _threadId: string;
  private _reactions: WsReactionAggregate[];
  private _updatesCursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    messageId: string,
    threadId: string,
    reactions: WsReactionAggregate[],
    updatesCursor?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._messageId = input.messageId;
    this._threadId = input.threadId;
    this._reactions = input.reactions;
    this._updatesCursor = input.updatesCursor;
    this._additionalProperties = input.additionalProperties;
  }

  get messageId(): string { return this._messageId; }
  set messageId(messageId: string) { this._messageId = messageId; }

  get threadId(): string { return this._threadId; }
  set threadId(threadId: string) { this._threadId = threadId; }

  get reactions(): WsReactionAggregate[] { return this._reactions; }
  set reactions(reactions: WsReactionAggregate[]) { this._reactions = reactions; }

  get updatesCursor(): string | undefined { return this._updatesCursor; }
  set updatesCursor(updatesCursor: string | undefined) { this._updatesCursor = updatesCursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsReaction;
