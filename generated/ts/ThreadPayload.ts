import WsThread from './WsThread';
class ThreadPayload {
  private _thread?: WsThread;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    thread?: WsThread,
    additionalProperties?: Map<string, any>,
  }) {
    this._thread = input.thread;
    this._additionalProperties = input.additionalProperties;
  }

  get thread(): WsThread | undefined { return this._thread; }
  set thread(thread: WsThread | undefined) { this._thread = thread; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default ThreadPayload;
