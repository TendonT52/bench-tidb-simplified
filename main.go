package main

import (
	"repo.blockfint.com/bodeesorn/bench-tidb-simplified/cmd"
	_ "repo.blockfint.com/bodeesorn/bench-tidb-simplified/db"
	_ "repo.blockfint.com/bodeesorn/bench-tidb-simplified/mem_mode"
	_ "repo.blockfint.com/bodeesorn/bench-tidb-simplified/seed"
)

func main() {
	err := cmd.ReadCmdArg()
	if err != nil {
		panic(err)
	}
}
