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
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"reflect"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus/istanbul"
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
	// logger.Info("*********************sendCommit*****************************")
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
	// logger.Debug("*********************handleCommit*****************************")
	// Decode COMMIT message
	var commit *istanbul.Subject
	err := msg.Decode(&commit)
	if err != nil {
		return errFailedDecodeCommit
	}

	if err := c.checkMessage(msgCommit, commit.View); err != nil {
		return err
	}

	if err := c.verifyCommit(commit, src); err != nil {
		return err
	}
	// log.Debug("handleCommit", "c.current.Commits.Size()", c.current.Commits.Size(), "c.valSet.Size()", c.valSet.Size(), "c.valSet.F()", c.valSet.F())
	c.acceptCommit(msg, src)

	// Commit the proposal once we have enough COMMIT messages and we are not in the Committed state.
	//
	// If we already have a proposal, we may have chance to speed up the consensus process
	// by committing the proposal without PREPARE messages.
	log.Debug("handleCommit", "commit_counts", c.current.Commits.Size(), "valSet_counts", c.valSet.Size(), "valSet_F", c.valSet.F())
	if c.current.Commits.Size() >= /*2*c.valSet.F()*/ c.valSet.Size()-c.valSet.F() && c.state.Cmp(StateCommitted) < 0 {
		logger.Debug("Get enough 2/3 commit messages")
		// Still need to call LockHash here since state can skip Prepared state and jump directly to the Committed state.
		c.current.LockHash()
		c.commit()
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
