package commander

import (
	"flag"
	"fmt"
	"reflect"
)

func DerefStruct(obj interface{}) (reflect.Type, bool) {
	st := reflect.TypeOf(obj)
	for st.Kind() == reflect.Ptr || st.Kind() == reflect.Interface {
		st = st.Elem()
	}

	return st, st.Kind() == reflect.Struct
}

func SetFlag(flagset *flag.FlagSet, obj interface{}, field reflect.StructField, directive string) error {
	fmt.Println("HERE", field.Name, directive)
	v := reflect.ValueOf(obj).Elem().FieldByName(field.Name)
	ptr := v.Addr().Interface().(*bool)
	flagset.BoolVar(ptr, directive, false, "asd")
	return nil
}
