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
	"fmt"

	"github.com/wasp-project/yazi/pkg/client"
	"github.com/wasp-project/yazi/pkg/protocol"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "yazictl <command>",
		Short: "The cli tool for Yazi",

		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
			HiddenDefaultCmd:  true,
		},
	}

	getCmd = &cobra.Command{
		Use:   "get",
		Short: "yaizctl get <key>",
		Run:   getf,
	}

	setCmd = &cobra.Command{
		Use:   "set",
		Short: "yazictl set <key> <value>",
		Run:   setf,
	}

	delCmd = &cobra.Command{
		Use:   "del",
		Short: "yazictl del <key>",
		Run:   delf,
	}

	mgetCmd = &cobra.Command{
		Use:   "mget",
		Short: "yazictl mget <key1> <key2>...",
		Run:   mgetf,
	}

	msetCmd = &cobra.Command{
		Use: "mset",
		// TODO: this doesn't make sense. Need to be redesigned
		Short: "yazictl mset <key1> <key2>... <value1> <value2>...",
		Run:   msetf,
	}

	keysCmd = &cobra.Command{
		Use:   "keys",
		Short: "yazictl keys",
		Run:   keysf,
	}

	expireCmd = &cobra.Command{
		Use:   "expire",
		Short: "yazictl expire <key> <ttl>",
		Run:   expiref,
	}
)

var (
	getf = func(cmd *cobra.Command, args []string) {
		client, err := client.NewYaziClient(protocol.Protocol(proto))
		if err != nil {
			panic(err)
		}

		if err := client.Connect(host, port); err != nil {
			panic(err)
		}
		defer client.Close()

		key = args[0]

		if val, err := client.Get(key); err != nil {
			panic(err)
		} else {
			fmt.Printf("%s", val)
		}
	}

	setf = func(cmd *cobra.Command, args []string) {
		client, err := client.NewYaziClient(protocol.Protocol(proto))
		if err != nil {
			panic(err)
		}
		if err := client.Connect(host, port); err != nil {
			panic(err)
		}
		defer client.Close()

		key = args[0]
		value = args[1]

		if err := client.Set(key, value); err != nil {
			panic(err)
		} else {
			fmt.Println("ok")
		}
	}

	delf = func(cmd *cobra.Command, args []string) {
		client, err := client.NewYaziClient(protocol.Protocol(proto))
		if err != nil {
			panic(err)
		}

		if err := client.Connect(host, port); err != nil {
			panic(err)
		}
		defer client.Close()

		key = args[0]

		if err := client.Del(key); err != nil {
			panic(err)
		}
	}

	mgetf = func(cmd *cobra.Command, args []string) {
		client, err := client.NewYaziClient(protocol.Protocol(proto))
		if err != nil {
			panic(err)
		}

		if err := client.Connect(host, port); err != nil {
			panic(err)
		}
		defer client.Close()

		keys = args

		if val, err := client.MGet(keys); err != nil {
			panic(err)
		} else {
			fmt.Printf("%v", val)
		}
	}

	msetf = func(cmd *cobra.Command, args []string) {
		client, err := client.NewYaziClient(protocol.Protocol(proto))
		if err != nil {
			panic(err)
		}

		if err := client.Connect(host, port); err != nil {
			panic(err)
		}
		defer client.Close()

		n := len(args)
		keys = args[:n/2-1]
		values = args[n/2:]

		if err := client.MSet(keys, values); err != nil {
			panic(err)
		}
	}

	keysf = func(cmd *cobra.Command, args []string) {
		client, err := client.NewYaziClient(protocol.Protocol(proto))
		if err != nil {
			panic(err)
		}

		if err := client.Connect(host, port); err != nil {
			panic(err)
		}
		defer client.Close()

		if val, err := client.Keys(); err != nil {
			panic(err)
		} else {
			fmt.Printf("%v", val)
		}
	}

	expiref = func(cmd *cobra.Command, args []string) {

	}
)

var (
	key    string
	value  string
	keys   []string
	values []string
	proto  string
	port   string
	host   string
)

func init() {
	rootCmd.Flags().StringVarP(&proto, "protocol", "p", "grpc", "client server protocol")
	rootCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "server host")
	rootCmd.Flags().StringVarP(&port, "port", "P", "3456", "server port")
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(delCmd)
	rootCmd.AddCommand(msetCmd)
	rootCmd.AddCommand(mgetCmd)
	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(expireCmd)
}

func main() {
	rootCmd.Execute()
}
