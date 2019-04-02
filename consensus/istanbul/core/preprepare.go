// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"errors"
	"time"

	"github.com/BCOSnetwork/BCOS-Go/consensus"
	"github.com/BCOSnetwork/BCOS-Go/consensus/istanbul"
)

func (c *core) sendPreprepare(request *istanbul.Request) {
	logger := c.logger.New("state", c.state)
	// If I'm the proposer and I have the same sequence with the proposal
	if c.current.Sequence().Cmp(request.Proposal.Number()) == 0 && c.IsProposer() {
		logger.Info("*****************sendPreprepare**************")
		curView := c.currentView()
		preprepare, err := Encode(&istanbul.Preprepare{
			View:     curView,
			Proposal: request.Proposal,
		})
		if err != nil {
			logger.Error("Failed to encode", "view", curView)
			return
		}

		c.broadcast(&message{
			Code: msgPreprepare,
			Msg:  preprepare,
		})
	}
}

func (c *core) handlePreprepare(msg *message, src istanbul.Validator) error {
	logger := c.logger.New("from", src, "state", c.state)
	logger.Info("*********************handlePreprepare*****************************")
	// Decode PRE-PREPARE
	var preprepare *istanbul.Preprepare
	err := msg.Decode(&preprepare)
	if err != nil {
		return errFailedDecodePreprepare
	}

	// ----parse future preprepare begin -------------------------------------------
	if preprepare.View.Sequence.Cmp(c.current.sequence) != 0 {
		logger.Warn("expected sequence ", c.current.sequence, "but get sequence ", preprepare.View.Sequence)
		return errors.New("unexpected sequence")
	}

	if preprepare.View.Round.Cmp(c.current.round) > 0 {
		return c.handleFuturePreprepare(preprepare, src)
	}
	// ----parse future preprepare end -------------------------------------------

	// Ensure we have the same view with the PRE-PREPARE message
	// If it is old message, see if we need to broadcast COMMIT
	if err := c.checkMessage(msgPreprepare, preprepare.View); err != nil {
		if err == errOldMessage {
			// Get validator set for the given proposal
			valSet := c.backend.ParentValidators(preprepare.Proposal).Copy()
			previousProposer := c.backend.GetProposer(preprepare.Proposal.Number().Uint64() - 1)
			valSet.CalcProposer(previousProposer, preprepare.View.Round.Uint64())
			// Broadcast COMMIT if it is an existing block
			// 1. The proposer needs to be a proposer matches the given (Sequence + Round)
			// 2. The given block must exist
			if valSet.IsProposer(src.Address()) && c.backend.HasPropsal(preprepare.Proposal.Hash(), preprepare.Proposal.Number()) {
				c.sendCommitForOldBlock(preprepare.View, preprepare.Proposal.Hash())
				return nil
			}
		}
		return err
	}

	// Check if the message comes from current proposer
	if !c.valSet.IsProposer(src.Address()) {
		logger.Warn("Ignore preprepare messages from non-proposer")
		return errNotFromProposer
	}

	// Verify the proposal we received
	if duration, err := c.backend.Verify(preprepare.Proposal); err != nil {
		logger.Warn("Failed to verify proposal", "err", err, "duration", duration)
		// if it's a future block, we will handle it again after the duration
		if err == consensus.ErrFutureBlock {
			c.stopFuturePreprepareTimer()
			c.futurePreprepareTimer = time.AfterFunc(duration, func() {
				c.sendEvent(backlogEvent{
					src: src,
					msg: msg,
				})
			})
		} else {
			c.sendNextRoundChange()
		}
		return err
	}

	// Here is about to accept the PRE-PREPARE
	if c.state == StateAcceptRequest {
		// Send ROUND CHANGE if the locked proposal and the received proposal are different
		if c.current.IsHashLocked() {
			if preprepare.Proposal.Hash() == c.current.GetLockedHash() {
				// Broadcast COMMIT and enters Prepared state directly
				c.acceptPreprepare(preprepare)
				c.setState(StatePrepared)
				c.sendCommit()
			} else {
				// Send round change
				c.sendNextRoundChange()
			}
		} else {
			// Either
			//   1. the locked proposal and the received proposal match
			//   2. we have no locked proposal
			c.acceptPreprepare(preprepare)
			c.setState(StatePreprepared)
			c.sendPrepare()
		}
	}

	return nil
}

func (c *core) handleFuturePreprepare(preprepare *istanbul.Preprepare, src istanbul.Validator) error {
	logger := c.logger.New("from", src, "state", c.state)
	logger.Info("*********************handleFuturePreprepare*****************************")

	// Get validator set for the given proposal
	valSet := c.backend.ParentValidators(preprepare.Proposal).Copy()
	previousProposer := c.backend.GetProposer(preprepare.Proposal.Number().Uint64() - 1)
	valSet.CalcProposer(previousProposer, preprepare.View.Round.Uint64())
	// 1. The proposer needs to be a proposer matches the given (Sequence + Round)
	// 2. The given block must exist
	if !valSet.IsProposer(src.Address()) || !c.backend.HasPropsal(preprepare.Proposal.Hash(), preprepare.Proposal.Number()) {
		logger.Warn("invalid future preprepare message")
		return errors.New("invalid future preprepare message")
	}

	if duration, err := c.backend.Verify(preprepare.Proposal); nil != err {
		logger.Warn("Failed to verify future proposal", "err", err, duration)
		return err
	}

	//把preprepare提议添加到当前preprepareSet
	if nil == c.current.preprepareSet {
		c.current.preprepareSet = make(PreprepareSet)
	}
	c.current.preprepareSet[preprepare.View.Round] = preprepare

	return nil
}

func (c *core) acceptPreprepare(preprepare *istanbul.Preprepare) {
	c.consensusTimestamp = time.Now()
	c.current.SetPreprepare(preprepare)
}
