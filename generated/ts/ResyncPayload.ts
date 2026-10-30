
class ResyncPayload {
  private _cursor?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    cursor?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._cursor = input.cursor;
    this._additionalProperties = input.additionalProperties;
  }

  get cursor(): string | undefined { return this._cursor; }
  set cursor(cursor: string | undefined) { this._cursor = cursor; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default ResyncPayload;
