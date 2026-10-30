
class AuthRequestPayload {
  private _xMinusWebitelMinusAccess: string;
  private _xMinusWebitelMinusClient?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    xMinusWebitelMinusAccess: string,
    xMinusWebitelMinusClient?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._xMinusWebitelMinusAccess = input.xMinusWebitelMinusAccess;
    this._xMinusWebitelMinusClient = input.xMinusWebitelMinusClient;
    this._additionalProperties = input.additionalProperties;
  }

  get xMinusWebitelMinusAccess(): string { return this._xMinusWebitelMinusAccess; }
  set xMinusWebitelMinusAccess(xMinusWebitelMinusAccess: string) { this._xMinusWebitelMinusAccess = xMinusWebitelMinusAccess; }

  get xMinusWebitelMinusClient(): string | undefined { return this._xMinusWebitelMinusClient; }
  set xMinusWebitelMinusClient(xMinusWebitelMinusClient: string | undefined) { this._xMinusWebitelMinusClient = xMinusWebitelMinusClient; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default AuthRequestPayload;
