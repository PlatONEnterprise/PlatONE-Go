package packet

import (
	"encoding/json"
	"fmt"
	"strings"
)

// todo: handle error
// ExtractContractData extract the role info from the contract return result
func ExtractContractData(result, role string) string {
	var inter = make([]interface{}, 0)
	var count int

	r, _ := ParseSysContractResult([]byte(result))
	data := r.Data.([]interface{})

	length := len(data)
	for i := 0; i < length; i++ {
		temp, _ := json.Marshal(data[0])
		if strings.Contains(string(temp), role) {
			inter = append(inter, data[i])
			count++
		}
	}

	if count == 0 {
		return fmt.Sprintf("no %s in registration\n", role)
	} else {
		r.Data = inter
		newContractData, _ := json.Marshal(r)
		return string(newContractData)
	}
}
