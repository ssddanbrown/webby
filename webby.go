package main

import (
	"fmt"
)

var manager *managerServer

func main() {

	fmt.Println("test")

	manager = new(managerServer)

	err := manager.addFileServer("./")
	checkErr(err)
	err = manager.addFileServer("./test/")
	checkErr(err)

	err = manager.listen(8080)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func intInSlice(integer int, list []int) bool {
	for _, v := range list {
		if v == integer {
			return true
		}
	}
	return false
}
