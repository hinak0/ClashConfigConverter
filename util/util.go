package util

import (
	"encoding/json"
	"fmt"
)

func PrintJson(obj any) {
	json, _ := json.MarshalIndent(obj, "", "	")
	fmt.Println(string(json))
}
