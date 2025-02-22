use crate::sdk;
use crate::sdk::FungibleTokenResponse;
use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, QueryRequest, Reply, ReplyOn, StdResult, SubMsg,
};
use cosmwasm_std::{DepsMut, Env, MessageInfo, Response, StdError, Uint128};
use cw2::set_contract_version;
use cw_storage_plus::Item;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use thiserror::Error;

// Flow of the smart contract:
// - `ExecuteMsg::Issue` call is sent to smart contract
// - smart contract creates two fungible tokens by executing native message delivered by the `asset` module
// - after creation of each fungible token, `reply` callback is executed
// - inside `reply` counter is incremented
// - caller queries the smart contract to verify the correct value of the counter

// version info for migration info
const CONTRACT_NAME: &str = "creates.io:issue-fungible-token";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct State {
    pub count: i32,
}

pub const STATE: Item<State> = Item::new("state");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response<sdk::Messages>, ContractError> {
    match msg {
        ExecuteMsg::Issue {
            symbol,
            subunit,
            precision,
            amount,
        } => issue_tokens(deps, symbol, subunit, precision, amount),
    }
}

fn issue_tokens(
    deps: DepsMut,
    symbol: String,
    subunit: String,
    precision: u32,
    amount: Uint128,
) -> Result<Response<sdk::Messages>, ContractError> {
    if amount == Uint128::zero() {
        return Err(ContractError::InvalidZeroAmount {});
    }

    let state = State { count: 0 };
    STATE.save(deps.storage, &state)?;

    // Send two submessages handled by the asset module to create two fungible tokens.
    // ReplyOn::Always means that we want `reply` to be called after each submessage execution.
    let mut msg1 = SubMsg::new(sdk::Messages::AssetFTMsgIssue {
        symbol: symbol.clone() + "1",
        subunit: subunit.clone() + "1",
        precision,
        initial_amount: amount,
    });
    msg1.reply_on = ReplyOn::Always;

    let mut msg2 = SubMsg::new(sdk::Messages::AssetFTMsgIssue {
        symbol: symbol.clone() + "2",
        subunit: subunit.clone() + "2",
        precision,
        initial_amount: amount,
    });
    msg2.reply_on = ReplyOn::Always;

    // As a part of the response we send two submessages which are then forwarded to the parser
    // in go.

    let res: Response<sdk::Messages> = Response::new()
        .add_attribute("method", "issue_token")
        .add_attribute("symbol", symbol)
        .add_attribute("amount", amount)
        .add_submessages([msg1, msg2]);
    Ok(res)
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn reply(deps: DepsMut, _env: Env, _msg: Reply) -> Result<Response, ContractError> {
    // After execution of each submessage this function is called.
    // Counter is incremented to confirm that callback is received.

    STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
        state.count += 1;
        Ok(state)
    })?;

    Ok(Response::new())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps<sdk::Queries>, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetCount {} => to_binary(&query_count(deps)?),
        QueryMsg::GetInfo { denom } => to_binary(&query_info(deps, denom)?),
    }
}

fn query_count(deps: Deps<sdk::Queries>) -> StdResult<CountResponse> {
    let state = STATE.load(deps.storage)?;
    Ok(CountResponse { count: state.count })
}

fn query_info(deps: Deps<sdk::Queries>, denom: String) -> StdResult<InfoResponse> {
    let request: QueryRequest<sdk::Queries> =
        sdk::Queries::AssetFTGetToken { denom: denom }.into();
    let res: FungibleTokenResponse = deps.querier.query(&request)?;
    Ok(InfoResponse { issuer: res.issuer })
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Issue {
        symbol: String,
        subunit: String,
        precision: u32,
        amount: Uint128,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    // GetCount returns the current count as a json-encoded number
    GetCount {},
    // GetInfo returns information about fungible token
    GetInfo { denom: String },
}

// We define a custom struct for each query response
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct CountResponse {
    pub count: i32,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InfoResponse {
    pub issuer: String,
}

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("Invalid zero amount")]
    InvalidZeroAmount {},

    #[error("Custom Error val: {val:?}")]
    CustomError { val: String },
    // Add any other custom errors you like here.
    // Look at https://docs.rs/thiserror/1.0.21/thiserror/ for details.
}
