
class ConnectedPayload {
  private _ok?: boolean;
  private _connectionId?: string;
  private _serverVersion?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    ok?: boolean,
    connectionId?: string,
    serverVersion?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._ok = input.ok;
    this._connectionId = input.connectionId;
    this._serverVersion = input.serverVersion;
    this._additionalProperties = input.additionalProperties;
  }

  get ok(): boolean | undefined { return this._ok; }
  set ok(ok: boolean | undefined) { this._ok = ok; }

  get connectionId(): string | undefined { return this._connectionId; }
  set connectionId(connectionId: string | undefined) { this._connectionId = connectionId; }

  get serverVersion(): string | undefined { return this._serverVersion; }
  set serverVersion(serverVersion: string | undefined) { this._serverVersion = serverVersion; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default ConnectedPayload;
