
class PingPayload {
  private _timestamp?: number;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    timestamp?: number,
    additionalProperties?: Map<string, any>,
  }) {
    this._timestamp = input.timestamp;
    this._additionalProperties = input.additionalProperties;
  }

  get timestamp(): number | undefined { return this._timestamp; }
  set timestamp(timestamp: number | undefined) { this._timestamp = timestamp; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default PingPayload;
