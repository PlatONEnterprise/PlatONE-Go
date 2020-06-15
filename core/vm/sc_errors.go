package vm

import "errors"

var (
	SC_ERR_FUNC_NOT_FOUND        = errors.New("The func not found")
	SC_ERR_PARAMS_NUMBER_INVALID = errors.New("The params number invalid")
	SC_ERR_NO_PERMISSION         = errors.New("No permission")
)

//for node system contract
var (
	SC_ERR_PAMAMS_INVALID = errors.New("The params invalid")
)
