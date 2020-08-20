package abi

import (
	"reflect"
)

// typeCheck checks that the given reflection value can be assigned to the reflection
// type in t.
func typeCheckV2(t Type, value reflect.Value) error {
	if t.T == SliceTy || t.T == ArrayTy {
		return sliceTypeCheckV2(t, value)
	}

	// Check base type validity. Element types will be checked later on.
	if t.GetType().Kind() != value.Kind() {
		return typeErr(t.GetType().Kind(), value.Kind())
	} else if t.T == FixedBytesTy && t.Size != value.Len() {
		return typeErr(t.GetType(), value.Type())
	} else {
		return nil
	}

}

// sliceTypeCheck checks that the given slice can by assigned to the reflection
// type in t.
func sliceTypeCheckV2(t Type, val reflect.Value) error {
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return typeErr(formatSliceString(t.GetType().Kind(), t.Size), val.Type())
	}

	if t.T == ArrayTy && val.Len() != t.Size {
		return typeErr(formatSliceString(t.Elem.GetType().Kind(), t.Size), formatSliceString(val.Type().Elem().Kind(), val.Len()))
	}

	if t.Elem.T == SliceTy || t.Elem.T == ArrayTy {
		if val.Len() > 0 {
			return sliceTypeCheck(*t.Elem, val.Index(0))
		}
	}

	if elemKind := val.Type().Elem().Kind(); elemKind != t.Elem.GetType().Kind() {
		return typeErr(formatSliceString(t.Elem.GetType().Kind(), t.Size), val.Type())
	}
	return nil
}
