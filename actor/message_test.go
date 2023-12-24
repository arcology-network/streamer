package actor

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/types"
	evmCommon "github.com/arcology-network/evm/common"
	"github.com/arcology-network/evm/core"
	evmTypes "github.com/arcology-network/evm/core/types"
)

func TestEncodeMessage(t *testing.T) {
	to := evmCommon.BytesToAddress([]byte{11, 8, 9, 10})
	newmsg := core.NewMessage(
		evmCommon.BytesToAddress([]byte{7, 8, 9, 10}),
		&to,
		uint64(10),
		big.NewInt(12),
		uint64(22),
		big.NewInt(34),
		[]byte{},
		nil,
		false,
	)
	msg := types.StandardMessage{
		Source: 1,
		TxHash: evmCommon.BytesToHash([]byte{1, 2, 3, 4, 5, 6}),
		Native: &newmsg,
	}

	actMessage := Message{
		Msgid:  uint64(2),
		Name:   "swdsd",
		Height: uint64(1),
		Round:  uint64(0),
		Data:   &msg,
	}

	bys, err := common.GobEncode(&actMessage)
	if err != nil {
		fmt.Printf("encode err=%v\n", err)
		return
	}
	fmt.Printf("bys=%x\n", bys)
	var receivedMessage Message
	err = common.GobDecode(bys, &receivedMessage)
	if err != nil {
		fmt.Printf("decode message err=%v\n", err)
		return
	}
	fmt.Printf("receivedMessage=%v\n", receivedMessage)
	receivedmsg := receivedMessage.Data.(*types.StandardMessage)
	fmt.Printf("receivedmsg=%v\n", receivedmsg)
	fmt.Printf("receivedmsg.msg=%v\n", receivedmsg.Native)
}

func TestEncodeTransaction(t *testing.T) {
	newtx := evmTypes.NewTransaction(
		uint64(10),
		evmCommon.BytesToAddress([]byte{7, 8, 9, 10}),
		big.NewInt(12),
		uint64(22),
		big.NewInt(34),
		[]byte{},
	)
	tx := types.StandardTransaction{
		Source: 1,
		TxHash: evmCommon.BytesToHash([]byte{1, 2, 3, 4, 5, 6}),
		Native: newtx,
	}
	bys, err := common.GobEncode(&tx)
	if err != nil {
		fmt.Printf("encode err=%v\n", err)
		return
	}
	fmt.Printf("bys=%x\n", bys)

	var receivedTx types.StandardTransaction
	err = common.GobDecode(bys, &receivedTx)
	if err != nil {
		fmt.Printf("decode err=%v\n", err)
		return
	}
	fmt.Printf("receivedTx=%v\n", receivedTx)
	fmt.Printf("receivedTx-txs=%v\n", receivedTx.Native)
}
