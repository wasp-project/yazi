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
        "os"

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

	memoryCmd = &cobra.Command{
		Use:   "memory",
		Short: "yazictl memory operations for OpenClaw",
	}

	memorySaveCmd = &cobra.Command{
		Use:   "save <key> <file_path>",
		Short: "save memory from file to yazi",
		Args:  cobra.ExactArgs(2),
		Run:   memorySavef,
	}

	memoryLoadCmd = &cobra.Command{
		Use:   "load <key> [file_path]",
		Short: "load memory from yazi to file or stdout",
		Args:  cobra.RangeArgs(1, 2),
		Run:   memoryLoadf,
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
		if n%2 != 0 {
			panic("mset expects even number of arguments")
		}
		keys = args[:n/2]
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

        memorySavef = func(cmd *cobra.Command, args []string) {
                client, err := client.NewYaziClient(protocol.Protocol(proto))
                if err != nil {
                        panic(err)
                }
                if err := client.Connect(host, port); err != nil {
                        panic(err)
                }
                defer client.Close()

                key = args[0]
                filePath := args[1]
                data, err := os.ReadFile(filePath)
                if err != nil {
                        panic(err)
                }

                if err := client.Set(key, string(data)); err != nil {
                        panic(err)
                } else {
                        fmt.Println("memory saved successfully")
                }
        }

        memoryLoadf = func(cmd *cobra.Command, args []string) {
                client, err := client.NewYaziClient(protocol.Protocol(proto))
                if err != nil {
                        panic(err)
                }
                if err := client.Connect(host, port); err != nil {
                        panic(err)
                }
                defer client.Close()

                key = args[0]

                val, err := client.Get(key)
                if err != nil {
                        panic(err)
                }

                if len(args) == 2 {
                        filePath := args[1]
                        if err := os.WriteFile(filePath, []byte(val), 0644); err != nil {
                                panic(err)
                        }
                        fmt.Printf("memory loaded to %s successfully\n", filePath)
                } else {
                        fmt.Printf("%s\n", val)
                }
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
        
        memoryCmd.AddCommand(memorySaveCmd)
        memoryCmd.AddCommand(memoryLoadCmd)
        rootCmd.AddCommand(memoryCmd)
}

func main() {
	rootCmd.Execute()
}
