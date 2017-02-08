package main

import "fmt"

func ccTxt(beforecolor int, s interface{}) string {
	return fmt.Sprintf("\033[1;%d;40m%v\033[0m", beforecolor, s)
}
func greenTxt(s interface{}) string {
	return ccTxt(32, s)
}
func redTxt(s interface{}) string {
	return ccTxt(31, s)
}
func yellowTxt(s interface{}) string {
	return ccTxt(33, s)
}
