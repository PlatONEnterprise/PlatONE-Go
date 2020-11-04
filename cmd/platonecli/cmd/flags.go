package cmd

import (
	"fmt"

	"gopkg.in/urfave/cli.v1"
)

var (
	// global flags
	UrlFlags = cli.StringFlag{
		Name: "url",
		Usage: `Specify the remote node IP trying to connect, 
		the url choice will be remembered util the --url flag provided next time
		url format: <ip>:<port>, eg: 127.0.0.1:6791`,
	}
	AccountFlags = cli.StringFlag{
		Name:  "account",
		Usage: "Specify the local or remote account address used to send the message calls",
	}
	GasFlags = cli.StringFlag{
		Name:  "gas",
		Usage: "Specify the gas allowance for the code execution",
	}
	GasPriceFlags = cli.StringFlag{
		Name:  "gasPrice",
		Usage: "Specify the number of <Token> to simulate paying for each unit of gas during execution", //TODO correct
	}
	LocalFlags = cli.BoolFlag{
		Name: "local",
		Usage: `Use default local account to send the message call, 
		the default local account keystore file locates at "<future feature>"`,
	}
	KeyfileFlags = cli.StringFlag{
		Name:  "keyfile",
		Usage: "Use local account to send the message call by specifying the key file",
	}
	SyncFlags = cli.BoolFlag{
		Name:  "sync",
		Usage: "Wait for the result of polling the Tx Receipt after executing the commands",
	}
	DefaultFlags = cli.BoolFlag{
		Name: "default",
		Usage: `Default the account settings by storing the
		current values of --account, --keyfile, and --local flags  to "./config/config.json"`,
	}

	// transfer
	TransferValueFlag = cli.StringFlag{
		Name:  "value",
		Value: "0x0",
		Usage: "Amount of <Token> to transfer",
	}

	// contract
	ContractParamFlag = cli.StringSliceFlag{
		Name:  "param",
		Usage: "Specify parameters of a contract method if needed, usage: --param \"p1\" --param \"p2\"",
	}
	ContractAbiFilePathFlag = cli.StringFlag{
		Name:  "abi",
		Usage: "Specify the wasm or evm abi file path",
	}
	ContractVmFlags = cli.StringFlag{
		Name:  "vm",
		Value: "wasm",
		Usage: `Choose the virtual machine interpreter for the contract execution and deployment, 
		"wasm" for WebAssembly virtual machine; "evm" for EVM virtual machine,
		The error may occur if the interpreter does not match.`,
	}
	ContractIDFlag = cli.StringFlag{
		Name:  "contract",
		Usage: "Contract name or address",
	}

	// user
	TelFlags = cli.StringFlag{
		Name:  "phone",
		Usage: "The mobile number of a user",
	}
	EmailFlags = cli.StringFlag{
		Name:  "email",
		Usage: "The email address of a user",
	}
	OrganizationFlags = cli.StringFlag{
		Name:  "organization",
		Usage: "The organization of a user",
	}
	UserRemarkFlags = cli.StringFlag{
		Name:  "remark",
		Usage: "User remark info",
	}
	UserIDFlags = cli.StringFlag{
		Name:  "user",
		Usage: "The address or name of a user",
	}
	UserRoleFlag = cli.StringFlag{
		Name:  "role",
		Usage: "A role of a user, e.g. --role <userRole>",
	}
	RolesFlag = cli.StringFlag{
		Name:  "roles",
		Usage: "Register user roles, e.g. --roles '[\"<role1>\",\"<role2>\"]'",
	}
	UserStatusFlag = cli.StringFlag{
		Name:  "status",
		Usage: "Status of a user",
	}

	// node
	NodeP2pPortFlags = cli.StringFlag{
		Name:  "p2pPort",
		Value: "16791",
		Usage: "Specify the node p2p port number",
	}
	NodeRpcPortFlags = cli.StringFlag{
		Name:  "rpcPort",
		Value: "6791",
		Usage: "Specify the node rpc port number",
	}
	NodeDelayNumFlags = cli.StringFlag{
		Name:  "delayNum",
		Usage: "Switch the node type to consensus after <delayNum> numbers of blocks generated",
	}
	NodePublicKeyFlags = cli.StringFlag{
		Name:  "publicKey",
		Usage: "Node's public key for secure p2p communication",
	}
	NodeDescFlags = cli.StringFlag{
		Name:  "desc",
		Usage: "The description of a node",
	}
	NodeTypeFlags = cli.StringFlag{
		Name:  "type",
		Usage: "The node type can be either \"observer\" or \"consensus\"",
	}
	NameFlags = cli.StringFlag{
		Name:  "name",
		Usage: "Node name, the name is unique",
	}
	NodeStatusFlags = cli.StringFlag{
		Name:  "status",
		Usage: "Status of a node: \"valid\" or \"invalid\"",
	}

	// user
	AddressFlags = cli.StringFlag{
		Name:  "addr",
		Usage: "The address of the user registered in the user platform",
	}

	// common?
	ShowAllFlags = cli.BoolFlag{
		Name:  "all",
		Usage: "List all the valid data object",
	}
	FwClearAllFlags = cli.BoolFlag{
		Name:  "all",
		Usage: "Clear the fire wall rules of all actions",
	}

	FilePathFlags = cli.StringFlag{
		Name:  "file",
		Value: defaultFwFilePath,
		Usage: "Specify the fire wall file path to be imported or exported",
	}

	// cns
	CnsVersionFlags = cli.StringFlag{
		Name:  "version",
		Value: "latest",
		Usage: `Specify the version of the cns name. 
		Usage: --version X.X.X.X, where X is number between 0 and 9`,
	}

	// admin
	AdminApproveFlags = cli.BoolFlag{
		Name:  "approve",
		Usage: "list the registration to be approved",
	}

	AdminDeleteFlags = cli.StringFlag{
		Name:  "delete",
		Usage: "list the data object can be deleted",
	}

	//fw
	FwActionFlags = cli.StringFlag{
		Name:  "action",
		Usage: "Specify the fire wall rule action, the fire wall action can be either \"accept\" or \"reject\".",
	}

	ShowContractMethodsFlag = cli.BoolFlag{
		Name:  "methods",
		Usage: "List all the contract methods",
	}

	// page
	PageNumFlags = cli.StringFlag{
		Name:  "pageNum",
		Value: "0",
		Usage: "Can be used with --pageSize, limit the output to the terminal",
	}

	PageSizeFlags = cli.StringFlag{
		Name:  "pageSize",
		Value: "0",
		Usage: "Can be used with --pageNum, limit the output to the terminal",
	}

	// system configurations
	BlockGasLimitFlags = cli.StringFlag{
		Name:  "block-gaslimit",
		Usage: "the gas limit of the block",
	}

	TxGasLimitFlags = cli.StringFlag{
		Name:  "tx-gaslimit",
		Usage: "the gas limit of transactions",
	}

	IsTxUseGasFlags = cli.StringFlag{
		Name: "tx-use-gas",
		Usage: fmt.Sprintf("if transactions use gas, "+
			"'%s' for transactions use gas, '%s' for not", txUseGas, txNotUseGas),
	}

	IsApproveDeployedContractFlags = cli.StringFlag{
		Name: "audit-con",
		Usage: fmt.Sprintf("approve the deployed contracts, "+
			"'%s' for allowing contracts audit, '%s' for not", conAudit, conNotAudit),
	}

	IsCheckContractDeployPermissionFlags = cli.StringFlag{
		Name: "check-perm",
		Usage: fmt.Sprintf("check the sender permission when deploying contracts, "+
			"'%s' for checking permission, '%s' for not", checkPerm, notCheckPerm),
	}

	IsProduceEmptyBlockFlags = cli.StringFlag{
		Name: "empty-block",
		Usage: fmt.Sprintf("consensus produces empty block, "+
			"'%s' for allowing to produce empty block, '%s' for not", prodEmp, notProdEmp),
	}

	GasContractNameFlags = cli.StringFlag{
		Name:  "gas-contract",
		Usage: "register the gas contract by contract name",
	}

	GetBlockGasLimitFlags = cli.BoolFlag{
		Name:  "block-gaslimit",
		Usage: "the gas limit of the block",
	}

	GetTxGasLimitFlags = cli.BoolFlag{
		Name:  "tx-gaslimit",
		Usage: "the gas limit of transactions",
	}

	GetIsTxUseGasFlags = cli.BoolFlag{
		Name: "tx-use-gas",
		Usage: fmt.Sprintf("if transactions use gas, "+
			"'%s' for transactions use gas, '%s' for not", txUseGas, txNotUseGas),
	}

	GetIsApproveDeployedContractFlags = cli.BoolFlag{
		Name: "audit-con",
		Usage: fmt.Sprintf("approve the deployed contracts, "+
			"'%s' for allowing contracts audit, '%s' for not", conAudit, conNotAudit),
	}

	GetIsCheckContractDeployPermissionFlags = cli.BoolFlag{
		Name: "check-perm",
		Usage: fmt.Sprintf("check the sender permission when deploying contracts, "+
			"'%s' for checking permission, '%s' for not", checkPerm, notCheckPerm),
	}

	GetIsProduceEmptyBlockFlags = cli.BoolFlag{
		Name: "empty-block",
		Usage: fmt.Sprintf("consensus produces empty block, "+
			"'%s' for allowing to produce empty block, '%s' for not", prodEmp, notProdEmp),
	}

	GetGasContractNameFlags = cli.BoolFlag{
		Name:  "gas-contract",
		Usage: "register the gas contract by contract name",
	}

	// rest
	RestPortFlags = cli.StringFlag{
		Name:  "port",
		Value: ":8000",
		Usage: "Specify the rest server listening port number, e.g. :8000",
	}

	//=============================================================================
	globalCmdFlags = []cli.Flag{
		UrlFlags,
		AccountFlags,
		GasFlags,
		GasPriceFlags,
		LocalFlags,
		KeyfileFlags,
		SyncFlags,
		DefaultFlags,
	}

	// system config
	sysConfigCmdFlags = append(
		globalCmdFlags,
		BlockGasLimitFlags,
		TxGasLimitFlags,
		IsTxUseGasFlags,
		IsApproveDeployedContractFlags,
		IsCheckContractDeployPermissionFlags,
		IsProduceEmptyBlockFlags,
		GasContractNameFlags,
	)

	getSysConfigCmdFlags = append(
		globalCmdFlags,
		GetBlockGasLimitFlags,
		GetTxGasLimitFlags,
		GetIsTxUseGasFlags,
		GetIsApproveDeployedContractFlags,
		GetIsCheckContractDeployPermissionFlags,
		GetIsProduceEmptyBlockFlags,
		GetGasContractNameFlags,
	)

	// user
	userAddCmdFlags    = append(globalCmdFlags, TelFlags, EmailFlags, OrganizationFlags)
	userUpdateCmdFlags = append(globalCmdFlags, TelFlags, EmailFlags, OrganizationFlags)
	userQueryCmdFlags  = append(globalCmdFlags, UserIDFlags, ShowAllFlags)

	// node
	nodeUpdateCmdFlags = append(globalCmdFlags, NodeDescFlags, NodeDelayNumFlags, NodeTypeFlags)
	nodeStatCmdFlags   = append(globalCmdFlags, NodeStatusFlags, NodeTypeFlags)
	nodeAddCmdFlags    = append(
		globalCmdFlags,
		NodeP2pPortFlags,
		NodeRpcPortFlags,
		NodeDelayNumFlags,
		NodeDescFlags)

	nodeQueryCmdFlasg = append(
		globalCmdFlags,
		ShowAllFlags,
		NodeTypeFlags,
		NodeStatusFlags,
		NodePublicKeyFlags,
		NameFlags)

	// contract
	contractDeployCmdFlags  = append(globalCmdFlags, ContractAbiFilePathFlag, ContractVmFlags, ContractParamFlag)
	contractExecuteCmdFlags = append(
		globalCmdFlags,
		ContractAbiFilePathFlag,
		ContractParamFlag,
		ContractVmFlags,
		TransferValueFlag,
		ShowContractMethodsFlag)
	contractMethodsCmd = append([]cli.Flag{}, ContractAbiFilePathFlag)

	// cns
	cnsResolveCmdFlags = append(globalCmdFlags, CnsVersionFlags)
	cnsQueryCmdFlags   = append(
		globalCmdFlags,
		ShowAllFlags,
		ContractIDFlag,
		AddressFlags,
		PageNumFlags,
		PageSizeFlags)

	//fw
	fwImportCmdFlags = append(globalCmdFlags, FilePathFlags)
	fwClearCmdFlags  = append(globalCmdFlags, FwActionFlags, FwClearAllFlags)

	// role
	roleCmdFlags = globalCmdFlags
)
