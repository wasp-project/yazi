// Copyright 2024 mlycore. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/database-mesh/golang-sdk/pkg/random"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yazibench <command>",
	Short: "The bench tool for Yazi",

	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
		HiddenDefaultCmd:  true,
	},
}

var bulksetCmd = &cobra.Command{
	Use:   "bulk",
	Short: "yazibench bulk",

	Run: func(cmd *cobra.Command, args []string) {
		conn := connect()
		defer conn.Close()

		for i := 0; i < n; i++ {
			key := random.StringN(4)
			value := random.StringN(6)

			req := fmt.Sprintf("+set %s %s;", key, value)
			_, err := conn.Write([]byte(req))
			if err != nil {
				panic(err)
			}

			reader := bufio.NewReader(conn)
			_, _, err = reader.ReadLine()
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("%d keys set\n", n)
	},
}

var (
	n int
)

func init() {
	rootCmd.AddCommand(bulksetCmd)
	bulksetCmd.Flags().IntVarP(&n, "number", "n", 1024, "batch set size n")
	_ = bulksetCmd.MarkFlagRequired("n")
}

func main() {
	rootCmd.Execute()
}

func connect() net.Conn {
	conn, err := net.Dial("tcp", "localhost:3456")
	if err != nil {
		panic(err)
	}
	return conn
}
