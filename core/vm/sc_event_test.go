package vm

import (
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/stretchr/testify/assert"
)

func Test_emitEvent(t *testing.T) {
	stateDB := newMockStateDB()
	topic := "Notify"
	msg := "success"
	code := uint64(1)

	emitEvent(common.Address{}, stateDB, 1, topic, code, msg)
	topicH := common.BytesToHash(crypto.Keccak256([]byte(topic)))
	log := stateDB.eLogs[topicH.String()]
	assert.NotEqual(t, nil, log)

	var data []rlp.RawValue
	err := rlp.DecodeBytes(log.Data, &data)
	assert.NoError(t, err)

	var code2 uint64
	err = rlp.DecodeBytes(data[0], &code2)
	assert.NoError(t, err)
	assert.Equal(t, code, code2)

	var msg2 string
	err = rlp.DecodeBytes(data[1], &msg2)
	assert.NoError(t, err)
	assert.Equal(t, msg, msg2)
}
