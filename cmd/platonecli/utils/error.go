package utils

import "errors"

var (
	ErrParamTypeFormat = "incorrect param type: %s, index: %d.\n"

	ErrReadFileFormat  = "read %s file error: %s\n"
	ErrOpenFileFormat  = "open %s file error: %s\n"
	ErrFindFileFormat  = "find file error: %s\n"
	ErrWriteFileFormat = "write file error: %s"

	//----------------------------------------------------------------
	ErrSendTransacionFormat = "send Transaction through http error: %s"
	ErrPackFunctionFormat   = "packet functions err: %s\n"
	ErrPackDataFormat       = "packet data err: %s\n"

	ErrParamCheckFormat    = "the input <%s> cannot be empty.\n"
	ErrParamNumCheckFormat = "param check error, required %d inputs, recieved %d.\n"
	//ErrParamValidFormat = "%s param is not valid: %s\n"
	ErrParamInValidSyntax = "invalid %s syntax.\n"
	ErrParamParseFormat   = "parse %s param error: %s\n"

	//ErrParamTypeFormat = "incorrect param type: %s, index: %d.\n"

	ErrParseFileFormat      = "parse %s file error: %s"
	ErrUnmarshalBytesFormat = "unmarshal %s bytes error: %s"

	ErrInputNullFormat = "the %s cannot be empty.\n"

	ErrRlpEncodeFormat = "rlp encode error: %s"
	ErrRlpDecodeFormat = "%s rlp decode error: %s"

	ErrGetFromChainFormat = "get %s from chain error: %s"

	ErrHttpSendFormat           = "send http post error:\n%s"
	ErrHttpNoResponseFormat     = "no response from node: %s"
	ErrHttpResponseStatusFormat = "http response status: %s"
	ErrRpcExecuationFormat      = "execute %s rpc call error: %s\n"

	ErrTODO = "something wrong, see more details in %s"

	PanicUnexpSituation = "unexpected situation in function %s\n"
)

var (
	ErrFileNull = errors.New("file path cannot be empty")
)
