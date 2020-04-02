package core

import "gopkg.in/urfave/cli.v1"

var (
	ConfigPathFlag = cli.StringFlag{
		Name:  "config",
		Usage: "config path",
	}
	PKFilePathFlag = cli.StringFlag{
		Name:  "pkfile",
		Value: "",
		Usage: "private key file path",
	}
	StabExecTimesFlag = cli.IntFlag{
		Name:  "times",
		Value: 1000,
		Usage: "execute times",
	}
	SendTxIntervalFlag = cli.IntFlag{
		Name:  "interval",
		Value: 10,
		Usage: "Time interval for sending transactions",
	}
	AccountSizeFlag = cli.IntFlag{
		Name:  "size",
		Value: 10,
		Usage: "account size",
	}
	/*
	TxJsonDataFlag = cli.StringFlag{
		Name:  "data",
		Usage: "transaction data",
	}*/
	ContractWasmFilePathFlag = cli.StringFlag{
		Name:  "code",
		Usage: "wasm file path",
	}
	ContractAddrFlag = cli.StringFlag{
		Name:  "addr",
		Usage: "the contract address",
	}
	ContractCnsNameFlag = cli.StringFlag{
		Name:  "cns",
		Usage: "contract name and version ,eg :\"proxyContract\"",
	}
	ContractFuncNameFlag = cli.StringFlag{
		Name:  "func",
		Usage: "function and param ,eg :set(1,\"a\")",
	}
	ContractParamFlag = cli.StringSliceFlag{
		Name:  "param",
		Usage: "params trans into contract function, eg: --param \"p1\" --param \"p2\"",
	}
	TransactionTypeFlag = cli.IntFlag{
		Name:  "type",
		Value: 2,
		Usage: "tx type ,default 2",
	}
	ContractAbiFilePathFlag = cli.StringFlag{
		Name:  "abi",
		Usage: "abi file path",
	}
	TransactionHashFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "tx hash",
	}
	TxFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "transaction sender addr",
	}
	TxToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "transaction acceptor addr",
	}
	TransferValueFlag = cli.StringFlag{
		Name:  "value",
		Value: "0xDE0B6B3A7640000", //one
		Usage: "transfer value",
	}

	// newly added
	ContractFlag = cli.StringFlag{
		Name:  "contract",
		Usage: "contract address or name",
	}

	RolesFlag = cli.StringFlag{
		Name:  "roles",
		Usage: "roles to be registered",
	}

	//

	AccountCmdFlags = cli.StringFlag{
		Name:  "account",
		Usage: "account flag test",
	}

	GasCmdFlags = cli.StringFlag{
		Name:  "gas",
		//Value: "",
		Usage: "",
	}

	GasPriceCmdFlags = cli.StringFlag{
		Name:  "gasPrice",
		//Value: "",
		Usage: "",
	}

	LocalCmdFlags = cli.BoolFlag{
		Name:  "local",
		Usage: "",
	}

	KeystoreCmdFlags = cli.StringFlag{
		Name:  "keystore",
		Usage: "",
	}

	SyncCmdFlags = cli.BoolFlag{
		Name:  "sync",
		Usage: "",
		// Destination: &isSync,
	}

	// ---------------------------------------------

	tmpCmdFlags = []cli.Flag{
		ContractAddrFlag,
	}

	globalCmdFlags = []cli.Flag{
		AccountCmdFlags,
		GasCmdFlags,
		GasPriceCmdFlags,
		LocalCmdFlags,
		KeystoreCmdFlags,
		SyncCmdFlags,
	}

	//

	executecmdFlags = []cli.Flag{
		ContractFlag,
		ContractFuncNameFlag, // temporary
		ContractParamFlag,
		ContractAbiFilePathFlag,
	}

	executeCmdFlags = append(executecmdFlags, globalCmdFlags...)



	/*
	fwCmdFlags = []cli.Flag{
		AccountCmdFlags,
		GasCmdFlags,
		GasPriceCmdFlags,
		SignCmdFlags,
		SyncCmdFlags,
	}*/
// ----------------------------------------------

	deployCmdFlags = []cli.Flag{
		ContractWasmFilePathFlag,
		ContractAbiFilePathFlag,
		ConfigPathFlag,
	}

	invokeCmdFlags = []cli.Flag{
		ContractFuncNameFlag,
		ContractParamFlag,
		ContractAbiFilePathFlag,
		ContractAddrFlag,
		ConfigPathFlag,
		TransactionTypeFlag,
	}

	cnsInvokeCmdFlags = []cli.Flag{
		ContractCnsNameFlag,
		ContractFuncNameFlag,
		ContractParamFlag,
		ContractAbiFilePathFlag,
		ConfigPathFlag,
		TransactionTypeFlag,

		ContractFlag,
	}

	fwInvokeCmdFlags = []cli.Flag{
		ContractAddrFlag,
		ContractFuncNameFlag,
		ContractParamFlag,
		TransactionTypeFlag,
		ConfigPathFlag,
	}

	migInvokeCmdFlags = []cli.Flag{
		ContractAddrFlag,
		ContractFuncNameFlag,
		ContractParamFlag,
		TransactionTypeFlag,
		ConfigPathFlag,
	}

	codeGenCmdFlags = []cli.Flag{
		ContractWasmFilePathFlag,
		ContractAbiFilePathFlag,
	}
	sendTransactionCmdFlags = []cli.Flag{
		TxFromFlag,
		TxToFlag,
		TransferValueFlag,
		ConfigPathFlag,
	}
	sendRawTransactionCmdFlags = []cli.Flag{
		PKFilePathFlag,
		TxFromFlag,
		TxToFlag,
		TransferValueFlag,
		ConfigPathFlag,
	}
	getTxReceiptCmdFlags = []cli.Flag{
		TransactionHashFlag,
		ConfigPathFlag,
	}

	stabilityCmdFlags = []cli.Flag{
		PKFilePathFlag,
		StabExecTimesFlag,
		SendTxIntervalFlag,
		ConfigPathFlag,
	}
	stabPrepareCmdFlags = []cli.Flag{
		PKFilePathFlag,
		AccountSizeFlag,
		TransferValueFlag,
		ConfigPathFlag,
	}
)
