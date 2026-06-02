package main

import "github.com/crawlab-team/crawlab-core/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
