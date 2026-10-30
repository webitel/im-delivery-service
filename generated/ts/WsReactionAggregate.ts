
class WsReactionAggregate {
  private _emoji?: string;
  private _count?: number;
  private _reactedByMe?: boolean;
  private _reactorIds?: string[];
  private _lastReactedAt?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    emoji?: string,
    count?: number,
    reactedByMe?: boolean,
    reactorIds?: string[],
    lastReactedAt?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._emoji = input.emoji;
    this._count = input.count;
    this._reactedByMe = input.reactedByMe;
    this._reactorIds = input.reactorIds;
    this._lastReactedAt = input.lastReactedAt;
    this._additionalProperties = input.additionalProperties;
  }

  get emoji(): string | undefined { return this._emoji; }
  set emoji(emoji: string | undefined) { this._emoji = emoji; }

  get count(): number | undefined { return this._count; }
  set count(count: number | undefined) { this._count = count; }

  get reactedByMe(): boolean | undefined { return this._reactedByMe; }
  set reactedByMe(reactedByMe: boolean | undefined) { this._reactedByMe = reactedByMe; }

  get reactorIds(): string[] | undefined { return this._reactorIds; }
  set reactorIds(reactorIds: string[] | undefined) { this._reactorIds = reactorIds; }

  get lastReactedAt(): string | undefined { return this._lastReactedAt; }
  set lastReactedAt(lastReactedAt: string | undefined) { this._lastReactedAt = lastReactedAt; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsReactionAggregate;
