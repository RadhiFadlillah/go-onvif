package onvif

import (
	"encoding/json"
	"strconv"
	"strings"
)

var testDevice = Device{
	XAddr:    "http://10.10.2.32/onvif/device_service",
	User:     "wtsonvif",
	Password: "watchtower1",
}

func interfaceToString(src interface{}) string {
	str, _ := src.(string)
	return str
}

func interfaceToBool(src interface{}) bool {
	strBool := interfaceToString(src)
	return strings.ToLower(strBool) == "true"
}

func interfaceToInt(src interface{}) int {
	strNumber := interfaceToString(src)
	number, _ := strconv.Atoi(strNumber)
	return number
}

func prettyJSON(src interface{}) string {
	result, _ := json.MarshalIndent(&src, "", "    ")
	return string(result)
}
