package integrationtests

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// FundedAccount represents a requirement of a test to get some funds for an account
type FundedAccount struct {
	Address sdk.AccAddress
	Amount  sdk.Coin
}

// NewFundedAccount is the constructor of FundedAccount
func NewFundedAccount(address sdk.AccAddress, amount sdk.Coin) FundedAccount {
	return FundedAccount{
		Address: address,
		Amount:  amount,
	}
}

// Faucet is the test chain faucet.
type Faucet struct {
	chainCtx ChainContext
	queue    chan fundingRequest

	// muCh is used to serve the same purpose as `sync.Mutex` to protect `fundingWallet` against being used
	// to broadcast many transactions in parallel by different integration tests. The difference between this and `sync.Mutex`
	// is that test may exit immediately when `ctx` is canceled, without waiting for mutex to be unlocked.
	muCh chan struct{}
}

// NewFaucet creates a new instance of the Faucet.
func NewFaucet(chainCtx ChainContext) Faucet {
	faucet := Faucet{
		chainCtx: chainCtx,
		queue:    make(chan fundingRequest),
		muCh:     make(chan struct{}, 1),
	}
	faucet.muCh <- struct{}{}
	return faucet
}

type fundingRequest struct {
	AccountsToFund []FundedAccount
	FundedCh       chan error
}

// FundAccounts funds the list of the received wallets.
func (f Faucet) FundAccounts(ctx context.Context, accountsToFund ...FundedAccount) (retErr error) {
	const maxAccountsPerRequest = 20

	if len(accountsToFund) > maxAccountsPerRequest {
		return errors.Errorf("the number of accounts to fund (%d) is greater than the allowed maximum (%d)", len(accountsToFund), maxAccountsPerRequest)
	}

	req := fundingRequest{
		AccountsToFund: accountsToFund,
		FundedCh:       make(chan error, 1),
	}

	// This `select` block is essential for understanding how the algorithm works.
	// It decides if the caller of the function is the leader of the transaction or just a regular participant.
	// There are 3 possible scenarios:
	// - `<-tf.muCh` succeeds - the caller becomes a leader of the transaction. Its responsibility is to collect requests from
	//    other participants, broadcast transaction and await it.
	// - `tf.queue <- req` succeeds - the caller becomes a participant and his request was accepted by the leader, accounts will be funded in coming block
	//   Caller waits until `<-req.FundedCh` succeeds, meaning that accounts were successfully funded or process failed.
	// - none of the above - meaning that current leader finished the process of collecting requests from participants and now
	//   transaction is broadcasted or awaited. Once it is finished `muCh` is unlocked and another caller will become a new leader
	//   accepting requests from other participants again.

	select {
	case <-ctx.Done():
		return ctx.Err()
	case f.queue <- req:
		// There is a leader who accepted this request. Now we must wait for transaction to be included in a block.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-req.FundedCh:
			return err
		}
	case <-f.muCh:
		// This call is a leader, it will collect requests from participants and broadcast transaction.
	}

	// Code below is executed by the leader.

	// This call may fail only because of cancelled context, so we don't need to propagate it to
	// other participants
	requests, err := f.collectRequests(ctx, req)
	if err != nil {
		return err
	}

	defer func() {
		// After transaction is broadcasted we unlock `muCh` so another leader for next transaction might be selected
		f.muCh <- struct{}{}

		// If leader got an error during broadcasting, that error is propagated to all the other participants.
		for _, req := range requests {
			req.FundedCh <- retErr
		}
	}()

	// All requests are collected, let's create messages and broadcast tx
	return f.broadcastTx(ctx, f.collectMessages(requests))
}

func (f Faucet) collectRequests(ctx context.Context, leaderReq fundingRequest) ([]fundingRequest, error) {
	const (
		requestsPerTx   = 20
		timeoutDuration = 100 * time.Millisecond
	)

	requests := make([]fundingRequest, 0, requestsPerTx)

	// Leader adds his own request to the batch
	requests = append(requests, leaderReq)

	// In the loop, we wait a moment to give other participants to join.
	timeout := time.After(timeoutDuration)
	for len(requests) < requestsPerTx {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			// We close the window when other participants might join the batch.
			// If someone comes after timeout they must wait for next leader.
			return requests, nil
		case req := <-f.queue:
			// Request from other participant is accepted and added to the batch.
			requests = append(requests, req)
		}
	}
	return requests, nil
}

func (f Faucet) collectMessages(requests []fundingRequest) []sdk.Msg {
	var messages []sdk.Msg
	for _, req := range requests {
		for _, acc := range req.AccountsToFund {
			messages = append(messages, &banktypes.MsgSend{
				FromAddress: f.chainCtx.ClientContext.FromAddress().String(),
				ToAddress:   acc.Address.String(),
				Amount:      sdk.NewCoins(acc.Amount),
			})
		}
	}
	return messages
}

func (f Faucet) broadcastTx(ctx context.Context, msgs []sdk.Msg) error {
	log := logger.Get(ctx)
	log.Info("Funding accounts for tests, it might take a while...")
	// FIXME (wojtek): use estimation once it is available in `tx` package
	gasLimit := uint64(len(msgs)) * f.chainCtx.GasLimitByMsgs(&banktypes.MsgSend{})

	// Transaction is broadcasted and awaited
	_, err := tx.BroadcastTx(
		ctx,
		f.chainCtx.ClientContext,
		f.chainCtx.TxFactory().WithGas(gasLimit),
		msgs...,
	)
	if err != nil {
		return err
	}
	log.Info("Test accounts funded")

	return nil
}

// FundAccountsWithOptions computes the needed balances and fund account with it.
func (f Faucet) FundAccountsWithOptions(ctx context.Context, address sdk.AccAddress, options BalancesOptions) error {
	amount := f.chainCtx.ComputeNeededBalanceFromOptions(options)
	return f.FundAccounts(ctx, FundedAccount{
		// TODO(dhil) refactor to use multiple accounts
		Address: address,
		Amount:  f.chainCtx.NewCoin(amount),
	})
}
