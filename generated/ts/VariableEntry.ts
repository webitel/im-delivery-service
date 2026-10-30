
class VariableEntry {
  private _value?: any;
  private _setBy?: string;
  private _setAt?: string;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    value?: any,
    setBy?: string,
    setAt?: string,
    additionalProperties?: Map<string, any>,
  }) {
    this._value = input.value;
    this._setBy = input.setBy;
    this._setAt = input.setAt;
    this._additionalProperties = input.additionalProperties;
  }

  get value(): any | undefined { return this._value; }
  set value(value: any | undefined) { this._value = value; }

  get setBy(): string | undefined { return this._setBy; }
  set setBy(setBy: string | undefined) { this._setBy = setBy; }

  get setAt(): string | undefined { return this._setAt; }
  set setAt(setAt: string | undefined) { this._setAt = setAt; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default VariableEntry;
