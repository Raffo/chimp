package main

import (
	"fmt"
	"strconv"
	"strings"
)

func GetStringFromArgs(arguments map[string]interface{}, name, _default string) string {
	in := arguments[name]
	if in == nil {
		return _default
	}
	return in.(string)
}

func GetIntFromStr(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("ERR: int conversion failed for %s: err: %s", s, err)
	}
	return int(i)
}

func GetBoolFromStr(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		fmt.Printf("ERR: bool conversion failed for %s: err: %s", s, err)
	}
	return b
}

func GetIntFromArgs(arguments map[string]interface{}, name string, _default int) int {
	if arguments[name] == nil {
		return _default
	}
	val_str := arguments[name].(string)
	return GetIntFromStr(val_str)
}

func GetBoolFromArgs(arguments map[string]interface{}, name string, _default bool) bool {
	if arguments[name] == nil {
		return _default
	}
	val_str := arguments[name].(string)
	return GetBoolFromStr(val_str)
}

func ConvertMaps(input string) map[string]string {
	var labels map[string]string = make(map[string]string)
	if input != "" {
		labelsArray := strings.Split(input, " ")
		for i := 0; i < len(labelsArray); i++ {
			if strings.Contains(labelsArray[i], "=") { //just to ensure it's actually a k,v pair
				kvArr := strings.Split(labelsArray[i], "=")
				for i := 0; i < len(kvArr); i++ {
					if i%2 == 0 {
						labels[kvArr[i]] = ""
					} else {
						labels[kvArr[i-1]] = kvArr[i]
					}
				}
			}

		}
	}
	return labels
}
