package test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/packet"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
)

var (
	block           interface{}
	blockHash       interface{}
	privKey         *ecdsa.PrivateKey
	account         interface{}
	contractAddress interface{}
	txHash          interface{}
	tx              interface{}
	tx2             interface{}
	txDeploy        interface{}
	privTest        string

	signData interface{}
	signMsg  interface{}
	err      error
)

const (
	PASSWORD1       = "0"
	PASSWORD1_WRONG = "x"
	PASSWORD2       = "123"
	PASSWORD3       = "456"

	KEYSTORE_FILE_PATH = "../../../release/linux/data/node-0/keystore/"
	TEST_FILE_PATH     = "test.txt"
	TEST_FILE_PATH2    = "test2.txt"
	TEST_FILE_PATH3    = "test3.txt"

	DURATION = 3

	ENODE = "enode://a24ac7c5484ef4ed0c5eb2d36620ba4e4aa13b8c84684e1b4aab0cebea2ae45cb4d375b77eab56516d34bfbd3c1a833fc51296ff084b770b94fb9028c4d25ccf@127.0.0.1:6792"

	BLOCK_NUM = "latest"

	deployContract = 1
)

func init() {
	utils.SetHttpUrl("127.0.0.1:6791")

	block, _ = utils.RpcCalls("eth_getBlockByNumber", []interface{}{BLOCK_NUM, true})
	blockHash = block.(map[string]interface{})["hash"]

	privKey, _ = ecdsa.GenerateKey(crypto.S256(), rand.Reader)

	account, _ = utils.RpcCalls("personal_newAccount", []interface{}{PASSWORD1})
	address := common.HexToAddress(account.(string))

	contractAddress = getConAddr(address)

	// tx = packet.NewTxParams(address, &address, "", "", "", "", 0)
	tx = newTx(address, "0x1")
	tx2 = newTx(address, "0x2")

	signData = hexutil.Encode([]byte("test"))
	signMsg, _ = utils.RpcCalls("personal_sign", []interface{}{signData, account, PASSWORD1})

	privTest = "18e14a7b6a307f426a94f8114701e7c8e774e7f9a47e2c2035db29a206321725"

	fmt.Printf("======================initial data============================\n")
	fmt.Printf("the account is %v\n", account)
	fmt.Printf("the transaction is %+v\n", tx)
	fmt.Printf("the transaction hash is %+v\n", txHash)
	fmt.Printf("the blockhash is %v\n", blockHash)
	fmt.Printf("the sign data is %v\n", signData)
	fmt.Printf("============================end===============================")

}

func getData() string {

	codePath := "../test/test_case/sol/sol.bin"
	abiPath := "../test/test_case/sol/sol.abi"
	vm := "evm"

	codeBytes, _ := utils.ParseFileToBytes(codePath)
	abiBytes, _ := utils.ParseFileToBytes(abiPath)

	call := packet.NewDeployCall(codeBytes, abiBytes, vm, deployContract)

	data, _, _, _ := call.CombineData()

	return data
}

func getConAddr(address common.Address) interface{} {
	_, _ = utils.RpcCalls("personal_unlockAccount", []interface{}{account, PASSWORD1, 10})
	txDeploy = packet.NewTxParams(address, nil, "", "", "", getData(), deployContract)
	txHash, _ = utils.RpcCalls("eth_sendTransaction", []interface{}{txDeploy})
	return packet.GetResponseByReceipt(txHash.(string))
}

func newTx(address common.Address, nonce string) interface{} {
	return struct {
		From     common.Address
		To       *common.Address
		Gas      string
		GasPrice string
		Value    string
		Data     string
		Nonce    string
		TxType   int
	}{address, &address, "", "", "", "", "0x1", 0}
}

/*
	Notation: the comments have the following meanings
	//?		: not test
	//x		: test failed
	///		: success, comment out temp
	///<spc>: success, not needed
*/
func TestPlatoneRPCs(t *testing.T) {
	var testCase = []struct {
		name   string
		params []interface{}
	}{
		//========================consensus.engine==================
		{"istanbul_getSnapshot", []interface{}{BLOCK_NUM}},
		{"istanbul_getSnapshotAtHash", []interface{}{blockHash}},
		{"istanbul_getValidators", []interface{}{BLOCK_NUM}},
		{"istanbul_getValidatorsAtHash", []interface{}{blockHash}},
		{"istanbul_candidates", nil},
		//?{"istanbul_propose", nil},
		//?{"istanbul_discard", nil},

		//=========================ethService: personal=============
		/*
			Namespace: personal; Struct: PrivateAccountAPI
		*/
		///{"personal_listAccounts", nil},
		///{"personal_listWallets", nil},
		///{"personal_openWallet", []interface{}{"127.0.0.1:6791",""}},
		//?{"personal_deriveAccount", []interface{}{"0x7"}},
		/// {"personal_newAccount", []interface{}{"123"}},
		///{"personal_importRawKey", []interface{}{privTest, PASSWORD2}},
		///{"personal_unlockAccount", []interface{}{account, PASSWORD1, 300}},
		///{"personal_lockAccount", []interface{}{account}},
		///{"personal_sendTransaction", []interface{}{tx, PASSWORD1}},
		/// {"personal_sendTransaction", []interface{}{tx, PASSWORD1_WRONG}},
		///{"personal_signTransaction", []interface{}{tx, PASSWORD1}},
		///{"personal_ecRecover", []interface{}{signData, signMsg}},
		/// {"personal_signAndSendTransaction", []interface{}{tx, PASSWORD1_WRONG}},	// deprecated
		/// {"personal_signAndSendTransaction", []interface{}{tx2, PASSWORD1}},

		//============================node.apis=========================
		///{"web3_clientVersion", nil},
		///{"web3_sha3", []interface{}{signData}},

		/*
			Namespace: debug; Struct: debug.HandlerT
		*/
		///{"debug_metrics", []interface{}{true}},					// TODO learn more
		///{"debug_verbosity", []interface{}{4}},					// TODO no return, 如何验证结果?
		///{"debug_vmodule", []interface{}{"p2p=6"}},				// TODO 无错误,如何验证结果?
		///{"debug_backtraceAt", []interface{}{"server.go:443"}},	// TODO 无错误,如何验证结果?
		///{"debug_memStats", nil},									// TODO learn more
		///{"debug_gcStats", nil},									// TODO learn more
		//?{"debug_cpuProfile", []interface{}{TEST_FILE_PATH,10}},
		//?{"debug_stopCPUProfile", nil},
		//?{"debug_startCPUProfile", []interface{}{TEST_FILE_PATH2}},
		//?{"debug_stopCPUProfile", nil},
		//?{"debug_goTrace", []interface{}{TEST_FILE_PATH3, DURATION}},
		//?{"debug_blockProfile", []interface{}{TEST_FILE_PATH3, DURATION}},
		/// {"debug_SetBlockProfileRate", nil},
		//?{"debug_writeBlockProfile", []interface{}{TEST_FILE_PATH3, DURATION}},
		//?{"debug_mutexProfile", []interface{}{TEST_FILE_PATH3, DURATION}},
		/// {"debug_SetMutexProfileFraction", nil},
		//?{"debug_writeMutexProfile", []interface{}{TEST_FILE_PATH3}},
		//?{"debug_writeMemProfile", []interface{}{TEST_FILE_PATH3}},
		///{"debug_stacks", nil},									// TODO learn more
		//?{"debug_FreeOSMemory", nil},
		///{"debug_setGCPercent", []interface{}{50}},				// TODO 如何验证结果?

		/*
			Namespace: admin; Struct: node.PrivateAdminAPI
		*/
		///{"admin_addPeer", []interface{}{ENODE}},			// TODO server.AddPeer(node)实际运行是否成功?
		///{"admin_removePeer", []interface{}{ENODE}},
		///{"admin_addTrustedPeer", []interface{}{ENODE}},
		///{"admin_removeTrustedPeer", []interface{}{ENODE}},
		//?{"admin_peerEvents", nil},						// TODO RPC subscription
		/// {"admin_startRPC", nil},
		//... .. no need to test

		/*
			Namespace: admin; Struct: node.PublicAdminAPI
		*/
		///{"admin_peers", nil},
		///{"admin_nodeInfo", nil},
		///{"admin_datadir", nil},

		//==============================ethapi=================================
		/*
			Namespace: txpool; Struct: ethapi.PublicTxPoolAPI
		*/
		///{"txpool_content", nil},
		///{"txpool_status", nil},
		///{"txpool_inspect", nil},

		/*
			Namespace: debug; Struct: ethapi.PublicDebugAPI
		*/
		///{"debug_getBlockRlp", []interface{}{7}},
		///{"debug_printBlock", []interface{}{7}},

		/*
			Namespace: debug; Struct: ethapi.PrivateDebugAPI
		*/
		///{"debug_chaindbProperty", []interface{}{"leveldb.stats"}},	// TODO learn more
		//?{"debug_chaindbCompact", nil},								// TODO 如何验证结果?
		//?{"debug_setHead", []interface{}{"0x7"}},						// TODO 如何验证结果?

		/*
			Namespace: eth; Struct: ethapi.PublicEthereumAPI
		*/
		///{"eth_gasPrice", nil},
		///{"eth_protocolVersion", nil},		// "0x3f"
		///{"eth_syncing", nil},

		/*
		   Namespace: eth; Struct: ethapi.PublicBlockChainAPI
		*/
		//?{"eth_monitor", []interface{}{"0x7"}},
		///{"eth_blockNumber", nil},
		///{"eth_getBalance", []interface{}{account, BLOCK_NUM}},
		///{"eth_getAccountBaseInfo", []interface{}{account, BLOCK_NUM}},
		/// {"eth_getBlockByNumber", []interface{}{BLOCK_NUM, true}},
		///{"eth_getBlockByHash", []interface{}{blockHash, false}},
		///{"eth_getCode", []interface{}{contractAddress, "latest"}},
		///{"eth_getStorageAt", []interface{}{contractAddress, "0x0", "latest"}},	// TODO "0x00...00" no result???
		/// {"eth_call", nil},						// "eth_call"
		///{"eth_estimateGas", []interface{}{txDeploy}},

		/*
			Namespace: eth; Struct: ethapi.PublicTransactionPoolAPI
		*/
		///{"eth_getBlockTransactionCountByNumber", []interface{}{BLOCK_NUM}},
		///{"eth_getBlockTransactionCountByHash", []interface{}{blockHash}},
		///{"eth_getTransactionByBlockNumberAndIndex", []interface{}{BLOCK_NUM, "0x0"}},
		///{"eth_getTransactionByBlockHashAndIndex", []interface{}{blockHash, "0x0"}},
		///{"eth_getRawTransactionByBlockNumberAndIndex", []interface{}{BLOCK_NUM, "0x0"}},
		///{"eth_getRawTransactionByBlockHashAndIndex", []interface{}{blockHash, "0x0"}},
		///{"eth_getTransactionCount", []interface{}{account, "latest"}},
		///{"eth_getTransactionByHash", []interface{}{txHash}},
		///{"eth_getRawTransactionByHash", []interface{}{txHash}},
		/// {"eth_getTransactionReceipt", nil},
		/// {"eth_sendTransaction", nil},
		/// {"eth_sendRawTransaction", nil},
		///{"eth_sign", []interface{}{account, signData}},
		///{"eth_signTransaction", []interface{}{tx}},
		///{"eth_pendingTransactions", nil},
		///{"eth_pendingTransactionsLength", nil},
		//?{"eth_resend", nil},

		/*
		   Namespace: eth; Struct: ethapi.PublicAccountAPI
		*/
		///{"eth_accounts", nil},

		//==============================eth=================================
		/*
			Namespace: net; Struct: ethapi.PublicNetAPI
		*/
		///{"net_listening", nil},
		///{"net_peerCount", nil},
		///{"net_version", nil},			// version: "1"

		/*
			Namespace: miner; Struct: eth.PrivateMinerAPI
		*/
		// not needed

		/*
		   Namespace: admin; Struct: eth.PrivateAdminAPI
		*/
		//?{"admin_importChain", nil},
		//?{"admin_exportChain", nil},

		/*
			Namespace: debug; Struct: eth.PublicDebugAPI
		*/
		///{"debug_dumpBlock", []interface{}{BLOCK_NUM}},

		/*
		   Namespace: debug; Struct: eth.PrivateDebugAPI
		*/
		//?{"debug_preimage", []interface{}{blockHash}},							// TODO
		///{"debug_getBadBlocks", nil},
		//?{"debug_storageRangeAt", []interface{}{blockHash, 0, contractAddress}}, 	// TODO
		///{"debug_getModifiedAccountsByNumber", []interface{}{0, 7}},				// TODO learn more
		///{"debug_getModifiedAccountsByHash", []interface{}{blockHash, nil}},		// TODO learn more

		/*
			Namespace: eth; Struct: eth.PublicEthereumAPI
		*/
		///{"eth_etherbase", nil},
		///{"eth_coinbase", nil},

		/*
		   Namespace: eth; Struct: eth.PublicMinerAPI
		*/
		///{"eth_mining", nil},									// TODO should be deprecated

		/*
			Namespace: eth; Struct: eth.PublicDownloaderAPI		// TODO
		*/
		///{"eth_syncing", nil},

		/*
		   Namespace: eth; Struct: eth.PublicFilterAPI			// TODO
		*/
		//?{"eth_newPendingTransactionFilter", nil},
		//?{"eth_newPendingTransactions", nil},
		//?{"eth_newBlockFilter", nil},
		//?{"eth_newHeads", nil},
		//?{"eth_logs", nil},
		//?{"eth_newFilter", nil},
		//?{"eth_getLogs", nil},
		//?{"eth_uninstallFilter", nil},
		//?{"eth_getFilterLogs", nil},
		//?{"eth_getFilterChanges", nil},

		//===================db?=======================			// TODO
	}

	for _, data := range testCase {
		t.Run(fmt.Sprintf("%s", data.name), func(t *testing.T) {

			result, err := utils.RpcCalls(data.name, data.params)

			switch {
			case err != nil:
				t.Errorf("input is %v, the error is %v\n", data.params, err.Error())
			case result == nil:
				t.Errorf("the return is null")
			default:
				t.Logf("the result is %v\n", result)
			}
		})
	}

	defer func() {
		fileName := utils.GetFileByKey(KEYSTORE_FILE_PATH, account.(string)[2:])
		_ = os.Remove(KEYSTORE_FILE_PATH + fileName)

	}()

}
