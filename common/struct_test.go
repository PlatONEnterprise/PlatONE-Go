package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckRequiredFieldsIsEmpty(t *testing.T) {
	tData := struct {
		Name   string `json:"name, required"`
		Age    int    `json:"age, required"`
		Gender int8   `json:"gender "`
	}{
		Name:   "elvindu",
		Gender: 1,
	}

	err := CheckRequiredFieldsIsEmpty(tData)
	assert.Error(t, err)

	tData2 := struct {
		Name   string `json:"name, required"`
		Age    int    `json:"age, required"`
		Gender int8   `json:"gender "`
	}{
		Name: "elvindu",
		Age:  33,
	}

	err = CheckRequiredFieldsIsEmpty(tData2)
	assert.NoError(t, err)

	tData3 := &struct {
		Name   string `json:"name "`
		Age    int    `json:"age, required"`
		Gender int8   `json:"gender "`
	}{
		Age:  33,
	}

	err = CheckRequiredFieldsIsEmpty(tData3)
	assert.NoError(t, err)

	err = CheckRequiredFieldsIsEmpty("")
	assert.Error(t, err)
}
