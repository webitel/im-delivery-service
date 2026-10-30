import VariableEntry from './VariableEntry';
import VariableAction from './VariableAction';
class WsVariables {
  private _variables?: Map<string, VariableEntry>;
  private _action?: VariableAction;
  private _additionalProperties?: Map<string, any>;

  constructor(input: {
    variables?: Map<string, VariableEntry>,
    action?: VariableAction,
    additionalProperties?: Map<string, any>,
  }) {
    this._variables = input.variables;
    this._action = input.action;
    this._additionalProperties = input.additionalProperties;
  }

  get variables(): Map<string, VariableEntry> | undefined { return this._variables; }
  set variables(variables: Map<string, VariableEntry> | undefined) { this._variables = variables; }

  get action(): VariableAction | undefined { return this._action; }
  set action(action: VariableAction | undefined) { this._action = action; }

  get additionalProperties(): Map<string, any> | undefined { return this._additionalProperties; }
  set additionalProperties(additionalProperties: Map<string, any> | undefined) { this._additionalProperties = additionalProperties; }
}
export default WsVariables;
