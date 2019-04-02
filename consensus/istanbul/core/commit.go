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
	"github.com/BCOSnetwork/BCOS-Go/log"
	"math/big"
	"reflect"

	"github.com/BCOSnetwork/BCOS-Go/common"
	"github.com/BCOSnetwork/BCOS-Go/consensus/istanbul"
)

func (c *core) sendCommit() {
	sub := c.current.Subject()
	c.broadcastCommit(sub)
}

func (c *core) sendCommitForOldBlock(view *istanbul.View, digest common.Hash) {
	sub := &istanbul.Subject{
		View:   view,
		Digest: digest,
	}
	c.broadcastCommit(sub)
}

func (c *core) broadcastCommit(sub *istanbul.Subject) {
	logger := c.logger.New("state", c.state)
	logger.Info("*********************sendCommit*****************************")
	encodedSubject, err := Encode(sub)
	if err != nil {
		logger.Error("Failed to encode", "subject", sub)
		return
	}
	c.broadcast(&message{
		Code: msgCommit,
		Msg:  encodedSubject,
	})
}

func (c *core) handleCommit(msg *message, src istanbul.Validator) error {
	logger := c.logger.New("from", src, "state", c.state)
	logger.Debug("*********************handleCommit*****************************")
	// Decode COMMIT message
	var commit *istanbul.Subject
	err := msg.Decode(&commit)
	if err != nil {
		return errFailedDecodeCommit
	}

	if commit.View.Sequence.Cmp(c.current.sequence) != 0 {
		return errors.New("future sequence's commit")
	}

	if commit.View.Round.Cmp(c.current.round) < 0 {
		return errors.New("old round's commit")
	}

	if err := c.addCommit(commit, msg, src); err != nil {
		return err
	}

	currentMaxRound := big.NewInt(-1)

	// get the max round on which we have a +2/3 prepares
	for r, _ := range c.current.preprepareSet {
		if c.current.commitSet[r].Size() >= c.valSet.Size()-c.valSet.F() {
			if r.Cmp(currentMaxRound) > 0 {
				currentMaxRound = r
			}
		}
	}

	if currentMaxRound.Cmp(c.current.lockedRound) > 0 {
		c.current.lockedRound = currentMaxRound
		c.current.lockedHash = c.current.preprepareSet[currentMaxRound].Proposal.Hash()
		c.current.round = currentMaxRound
		c.current.Preprepare = c.current.preprepareSet[currentMaxRound]
		c.current.Prepares = c.current.prepareSet[currentMaxRound]
		c.current.Commits = c.current.commitSet[currentMaxRound]

		c.sendCommit()
		c.commit()
		return nil
	}

	if err := c.checkMessage(msgCommit, commit.View); err != nil {
		return err
	}

	if err := c.verifyCommit(commit, src); err != nil {
		return err
	}
	log.Debug("handleCommit", "c.current.Commits.Size()", c.current.Commits.Size(), "c.valSet.Size()", c.valSet.Size(), "c.valSet.F()", c.valSet.F())
	c.acceptCommit(msg, src)

	// Commit the proposal once we have enough COMMIT messages and we are not in the Committed state.
	//
	// If we already have a proposal, we may have chance to speed up the consensus process
	// by committing the proposal without PREPARE messages.
	log.Debug("handleCommit", "c.current.Commits.Size()", c.current.Commits.Size(), "c.valSet.Size()", c.valSet.Size(), "c.valSet.F()", c.valSet.F())
	if c.current.Commits.Size() >= /*2*c.valSet.F()*/ c.valSet.Size()-c.valSet.F() && c.state.Cmp(StateCommitted) < 0 {
		logger.Info("*********************get Enough 2/3 handleCommit*****************************")
		// Still need to call LockHash here since state can skip Prepared state and jump directly to the Committed state.
		c.current.LockHash()
		c.commit()
	}

	return nil
}

func (c *core) addCommit(commit *istanbul.Subject, msg *message, src istanbul.Validator) error {
	logger := c.logger.New("addNewCommit", "from", src, "state", c.state)

	msgSet, ok := c.current.commitSet[commit.View.Round]
	// Add the COMMIT message to current round state
	if ok {
		if err := msgSet.Add(msg); err != nil {
			logger.Error("Failed to record commit message", "msg", msg, "err", err)
			return err
		}
	} else {
		commitSet := newMessageSet(c.valSet)
		commitSet.view.Sequence = commit.View.Sequence
		commitSet.view.Round = commit.View.Round

		if err := commitSet.Add(msg); err != nil {
			logger.Error("Failed to record commit message", "msg", msg, "err", err)
			return err
		}
		c.current.commitSet[commit.View.Round] = commitSet
	}

	return nil
}

// verifyCommit verifies if the received COMMIT message is equivalent to our subject
func (c *core) verifyCommit(commit *istanbul.Subject, src istanbul.Validator) error {
	logger := c.logger.New("from", src, "state", c.state)

	sub := c.current.Subject()
	if !reflect.DeepEqual(commit, sub) {
		logger.Warn("Inconsistent subjects between commit and proposal", "expected", sub, "got", commit)
		return errInconsistentSubject
	}

	return nil
}

func (c *core) acceptCommit(msg *message, src istanbul.Validator) error {
	logger := c.logger.New("from", src, "state", c.state)

	// Add the COMMIT message to current round state
	if err := c.current.Commits.Add(msg); err != nil {
		logger.Error("Failed to record commit message", "msg", msg, "err", err)
		return err
	}

	return nil
}
