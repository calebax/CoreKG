package foresttype

import (
	"fmt"
	"testing"
)

func TestVerify(t *testing.T) {
	// props := Properties{
	// 	{Name: "username", Type: "string", Defaults: "guest"},
	// 	{Name: "age", Type: "int64", Defaults: 18},
	// 	{Name: "height", Type: "double", Defaults: 1.70},
	// 	{Name: "isAdmin", Type: "bool", Defaults: false},
	// }
	// pvs := PropertiesValues{
	// 	{Name: "age", Value: "25"},
	// 	{Name: "username", Value: "Alice"},
	// 	{Name: "username2", Value: "Alice"},
	// }
	// res, err := pvs.ValidateAndComplete(props)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("%+v\n", logs.JSON(res))

	fmt.Println(validateValueType(-10, "double"))
}
