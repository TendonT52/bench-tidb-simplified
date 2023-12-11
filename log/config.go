package log

import "repo.blockfint.com/bodeesorn/bench-tidb-simplified/cmd"

var logLevel *string

func init() {
	logLevel = cmd.RootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (default is info)")
}
