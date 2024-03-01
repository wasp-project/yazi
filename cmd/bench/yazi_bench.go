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
	"net"
	"time"

	"github.com/database-mesh/golang-sdk/pkg/random"
	"github.com/spf13/cobra"
	"github.com/wasp-project/yazi/pkg/client"
	"github.com/wasp-project/yazi/pkg/protocol"
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
	Run:   bulksetf,
}
var (
	bulksetf = func(cmd *cobra.Command, args []string) {
		client, err := client.NewYaziClient(protocol.Protocol(proto))
		if err != nil {
			panic(err)
		}

		if err := client.Connect(host, port); err != nil {
			panic(err)
		}
		defer client.Close()

		t := time.NewTicker(1 * time.Second)

		var (
			sum int
			ite int
		)
	FOR:
		for {
			select {
			case <-t.C:
				{
					fmt.Printf("current: %d/s\n", ite)
					sum += ite
					ite = 0
				}
			default:
				{
					if sum <= number {
						key := random.StringN(4)
						value := random.StringN(6)

						if err := client.Set(key, value); err != nil {
							panic(err)
						}

						ite++
					} else {
						break FOR
					}
				}
			}
		}

		fmt.Printf("%d keys set\n", number)
	}
)

var (
	number int
	proto  string
	host   string
	port   string
)

func init() {
	rootCmd.Flags().StringVarP(&proto, "protocol", "p", "grpc", "client server protocol")
	rootCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "server host")
	rootCmd.Flags().StringVarP(&port, "port", "P", "3456", "server port")
	rootCmd.AddCommand(bulksetCmd)
	bulksetCmd.Flags().IntVarP(&number, "number", "n", 1024, "batch set size n")
	_ = bulksetCmd.MarkFlagRequired("n")
}

func main() {
	_ = rootCmd.Execute()
}

func connect() net.Conn {
	conn, err := net.Dial("tcp", "localhost:3456")
	if err != nil {
		panic(err)
	}
	return conn
}
