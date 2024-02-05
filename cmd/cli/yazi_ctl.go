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

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yazictl <command>",
	Short: "The cli tool for Yazi",

	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
		HiddenDefaultCmd:  true,
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "yaizctl get <key>",

	Run: func(cmd *cobra.Command, args []string) {
		key = args[0]
		conn := connect()
		defer conn.Close()

		req := fmt.Sprintf("+get %s;", key)
		_, err := conn.Write([]byte(req))
		if err != nil {
			panic(err)
		}

		reader := bufio.NewReader(conn)
		resp, _, err := reader.ReadLine()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", resp)
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "yazictl set <key> <value>",

	Run: func(cmd *cobra.Command, args []string) {
		key = args[0]
		value = args[1]

		conn := connect()
		defer conn.Close()

		req := fmt.Sprintf("+set %s %s;", key, value)
		_, err := conn.Write([]byte(req))
		if err != nil {
			panic(err)
		}

		reader := bufio.NewReader(conn)
		resp, _, err := reader.ReadLine()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", resp)
	},
}

var expireCmd = &cobra.Command{
	Use:   "expire",
	Short: "yazictl expire <key> <ttl>",

	Run: func(cmd *cobra.Command, args []string) {

	},
}

var (
	key   string
	value string
)

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(expireCmd)
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
