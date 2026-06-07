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
	"encoding/json"
	"fmt"

	"github.com/wasp-project/yazi/pkg/client"
	"github.com/wasp-project/yazi/pkg/memory"
	"github.com/wasp-project/yazi/pkg/memory/provider"
	"github.com/wasp-project/yazi/pkg/protocol"
	"github.com/wasp-project/yazi/pkg/tenant"

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
		Short: "yazictl memory <basic|advanced|policy>",
	}

	memoryBasicCmd = &cobra.Command{
		Use:   "basic",
		Short: "yazictl memory basic <put|get|list|del>",
	}
	memoryAdvancedCmd = &cobra.Command{
		Use:   "advanced",
		Short: "yazictl memory advanced <put|get|list|del>",
	}
	memoryPolicyCmd = &cobra.Command{
		Use:   "policy",
		Short: "yazictl memory policy <put|get|list|del>",
	}

	memoryRecallCmd = &cobra.Command{
		Use:   "recall",
		Short: "yazictl memory recall --query '<text>' [--profile student|standard] [--tag T] [--top-k N]",
		Long: "Recall memories with a cost-aware composition profile. 'student' (default) reads " +
			"the durable store at zero token cost; 'standard' embeds and vector-searches. " +
			"Prints the hits plus the metered Usage and estimated cost.",
		Run: memoryRecallf,
	}

	memoryBasicPutCmd = &cobra.Command{
		Use:   "put",
		Short: "yazictl memory basic put --json '<payload>'",
		Run:   memoryBasicPutf,
	}
	memoryBasicGetCmd = &cobra.Command{
		Use:   "get <id>",
		Short: "yazictl memory basic get <id>",
		Run:   memoryBasicGetf,
	}
	memoryBasicListCmd = &cobra.Command{
		Use:   "list",
		Short: "yazictl memory basic list --filter '<filter>'",
		Run:   memoryBasicListf,
	}
	memoryBasicDelCmd = &cobra.Command{
		Use:   "del <id>",
		Short: "yazictl memory basic del <id>",
		Run:   memoryBasicDelf,
	}

	memoryAdvancedPutCmd = &cobra.Command{
		Use:   "put",
		Short: "yazictl memory advanced put --json '<payload>'",
		Run:   memoryAdvancedPutf,
	}
	memoryAdvancedGetCmd = &cobra.Command{
		Use:   "get <id>",
		Short: "yazictl memory advanced get <id>",
		Run:   memoryAdvancedGetf,
	}
	memoryAdvancedListCmd = &cobra.Command{
		Use:   "list",
		Short: "yazictl memory advanced list --filter '<filter>'",
		Run:   memoryAdvancedListf,
	}
	memoryAdvancedDelCmd = &cobra.Command{
		Use:   "del <id>",
		Short: "yazictl memory advanced del <id>",
		Run:   memoryAdvancedDelf,
	}

	memoryPolicyPutCmd = &cobra.Command{
		Use:   "put",
		Short: "yazictl memory policy put --json '<payload>'",
		Run:   memoryPolicyPutf,
	}
	memoryPolicyGetCmd = &cobra.Command{
		Use:   "get <id>",
		Short: "yazictl memory policy get <id>",
		Run:   memoryPolicyGetf,
	}
	memoryPolicyListCmd = &cobra.Command{
		Use:   "list",
		Short: "yazictl memory policy list --filter '<filter>'",
		Run:   memoryPolicyListf,
	}
	memoryPolicyDelCmd = &cobra.Command{
		Use:   "del <id>",
		Short: "yazictl memory policy del <id>",
		Run:   memoryPolicyDelf,
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

	memoryBasicPutf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		var payload memory.BasicMemory
		if err := json.Unmarshal([]byte(memoryJSON), &payload); err != nil {
			panic(err)
		}
		id, err := store.PutBasic(payload)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", id)
	}

	memoryBasicGetf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		m, err := store.GetBasic(args[0])
		if err != nil {
			panic(err)
		}
		data, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", data)
	}

	memoryBasicListf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		var filter memory.BasicFilter
		if memoryFilter != "" {
			if err := json.Unmarshal([]byte(memoryFilter), &filter); err != nil {
				panic(err)
			}
		}
		list, err := store.ListBasic(filter)
		if err != nil {
			panic(err)
		}
		data, err := json.Marshal(list)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", data)
	}

	memoryBasicDelf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		if err := store.DeleteBasic(args[0]); err != nil {
			panic(err)
		}
	}

	memoryAdvancedPutf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		var payload memory.AdvancedMemory
		if err := json.Unmarshal([]byte(memoryJSON), &payload); err != nil {
			panic(err)
		}
		id, err := store.PutAdvanced(payload)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", id)
	}

	memoryAdvancedGetf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		m, err := store.GetAdvanced(args[0])
		if err != nil {
			panic(err)
		}
		data, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", data)
	}

	memoryAdvancedListf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		var filter memory.AdvancedFilter
		if memoryFilter != "" {
			if err := json.Unmarshal([]byte(memoryFilter), &filter); err != nil {
				panic(err)
			}
		}
		list, err := store.ListAdvanced(filter)
		if err != nil {
			panic(err)
		}
		data, err := json.Marshal(list)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", data)
	}

	memoryAdvancedDelf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		if err := store.DeleteAdvanced(args[0]); err != nil {
			panic(err)
		}
	}

	memoryPolicyPutf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		var payload memory.CognitivePolicy
		if err := json.Unmarshal([]byte(memoryJSON), &payload); err != nil {
			panic(err)
		}
		id, err := store.PutPolicy(payload)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", id)
	}

	memoryPolicyGetf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		m, err := store.GetPolicy(args[0])
		if err != nil {
			panic(err)
		}
		data, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", data)
	}

	memoryPolicyListf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		var filter memory.PolicyFilter
		if memoryFilter != "" {
			if err := json.Unmarshal([]byte(memoryFilter), &filter); err != nil {
				panic(err)
			}
		}
		list, err := store.ListPolicy(filter)
		if err != nil {
			panic(err)
		}
		data, err := json.Marshal(list)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", data)
	}

	memoryPolicyDelf = func(cmd *cobra.Command, args []string) {
		store, closeFn := newMemoryStore()
		defer closeFn()
		if err := store.DeletePolicy(args[0]); err != nil {
			panic(err)
		}
	}

	memoryRecallf = func(cmd *cobra.Command, args []string) {
		q := recallQuery
		if q == "" && len(args) > 0 {
			q = args[0]
		}
		if q == "" {
			panic("recall needs a query: --query '<text>' (or a positional arg)")
		}
		store, closeFn := newMemoryStore()
		defer closeFn()

		pricing := provider.DefaultPricing()
		cfg := provider.Config{
			Profile: recallProfile,
			Pricing: pricing,
			Budget:  provider.Budget{RecallContextTokens: recallMaxContext},
		}
		meter := provider.NewMeter()
		query := provider.Query{Text: q, Tag: recallTag, TopK: recallTopK}

		hits, usage, err := memory.RecallWithProfile(store, cfg, query, meter, tenantID)
		if err != nil {
			panic(err)
		}

		type hitOut struct {
			ID    string  `json:"id"`
			Text  string  `json:"text"`
			Score float64 `json:"score"`
		}
		outHits := make([]hitOut, len(hits))
		for i, h := range hits {
			outHits[i] = hitOut{ID: h.ID, Text: h.Text, Score: h.Score}
		}
		profile := recallProfile
		if profile == "" {
			profile = "student"
		}
		out := map[string]interface{}{
			"profile": profile,
			"hits":    outHits,
			"usage":   usage,
			"costUSD": pricing.Cost(usage),
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", data)
	}
)

var (
	key          string
	value        string
	keys         []string
	values       []string
	proto        string
	port         string
	host         string
	memoryJSON   string
	memoryFilter string
	tenantID     string

	recallQuery      string
	recallProfile    string
	recallTag        string
	recallTopK       int
	recallMaxContext int
)

func init() {
	rootCmd.Flags().StringVarP(&proto, "protocol", "p", "grpc", "client server protocol")
	rootCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "server host")
	rootCmd.Flags().StringVarP(&port, "port", "P", "3456", "server port")
	// --tenant is persistent so it propagates to the memory subcommands. Empty
	// (default) routes through the tenant bypass: no namespacing, original keys.
	rootCmd.PersistentFlags().StringVar(&tenantID, "tenant", "", "tenant id for memory namespacing (empty = bypass)")
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(delCmd)
	rootCmd.AddCommand(msetCmd)
	rootCmd.AddCommand(mgetCmd)
	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(expireCmd)

	rootCmd.AddCommand(memoryCmd)
	memoryCmd.AddCommand(memoryBasicCmd)
	memoryCmd.AddCommand(memoryAdvancedCmd)
	memoryCmd.AddCommand(memoryPolicyCmd)
	memoryCmd.AddCommand(memoryRecallCmd)

	memoryBasicCmd.AddCommand(memoryBasicPutCmd)
	memoryBasicCmd.AddCommand(memoryBasicGetCmd)
	memoryBasicCmd.AddCommand(memoryBasicListCmd)
	memoryBasicCmd.AddCommand(memoryBasicDelCmd)

	memoryAdvancedCmd.AddCommand(memoryAdvancedPutCmd)
	memoryAdvancedCmd.AddCommand(memoryAdvancedGetCmd)
	memoryAdvancedCmd.AddCommand(memoryAdvancedListCmd)
	memoryAdvancedCmd.AddCommand(memoryAdvancedDelCmd)

	memoryPolicyCmd.AddCommand(memoryPolicyPutCmd)
	memoryPolicyCmd.AddCommand(memoryPolicyGetCmd)
	memoryPolicyCmd.AddCommand(memoryPolicyListCmd)
	memoryPolicyCmd.AddCommand(memoryPolicyDelCmd)

	memoryBasicPutCmd.Flags().StringVar(&memoryJSON, "json", "", "memory payload json")
	memoryAdvancedPutCmd.Flags().StringVar(&memoryJSON, "json", "", "memory payload json")
	memoryPolicyPutCmd.Flags().StringVar(&memoryJSON, "json", "", "memory payload json")
	memoryBasicListCmd.Flags().StringVar(&memoryFilter, "filter", "", "filter json")
	memoryAdvancedListCmd.Flags().StringVar(&memoryFilter, "filter", "", "filter json")
	memoryPolicyListCmd.Flags().StringVar(&memoryFilter, "filter", "", "filter json")

	memoryRecallCmd.Flags().StringVar(&recallQuery, "query", "", "recall query text")
	memoryRecallCmd.Flags().StringVar(&recallProfile, "profile", "student", "composition profile: student | standard")
	memoryRecallCmd.Flags().StringVar(&recallTag, "tag", "", "restrict candidates to this tag")
	memoryRecallCmd.Flags().IntVar(&recallTopK, "top-k", 5, "max memories to return")
	memoryRecallCmd.Flags().IntVar(&recallMaxContext, "max-context-tokens", 0, "cap recalled context tokens (0 = unbounded)")
}

func main() {
	rootCmd.Execute()
}

type memoryClientKV struct {
	cli client.Client
}

func (k *memoryClientKV) Get(key string) (string, error) {
	return k.cli.Get(key)
}

func (k *memoryClientKV) Set(key, value string) error {
	return k.cli.Set(key, value)
}

func (k *memoryClientKV) Del(key string) error {
	return k.cli.Del(key)
}

func (k *memoryClientKV) MGet(keys []string) ([]string, error) {
	return k.cli.MGet(keys)
}

func (k *memoryClientKV) MSet(keys, values []string) error {
	return k.cli.MSet(keys, values)
}

func (k *memoryClientKV) Keys() ([]string, error) {
	return k.cli.Keys()
}

func newMemoryStore() (*memory.Store, func()) {
	cli, err := client.NewYaziClient(protocol.Protocol(proto))
	if err != nil {
		panic(err)
	}
	if err := cli.Connect(host, port); err != nil {
		panic(err)
	}
	kv := &memoryClientKV{cli: cli}

	// Route through the tenant-service layer. An empty --tenant resolves via the
	// bypass service to an empty prefix, preserving the original key layout.
	svc := tenantService()
	ctx, err := svc.Resolve(tenant.Request{TenantID: tenantID})
	if err != nil {
		panic(err)
	}
	return memory.NewStoreWithTenant(kv, ctx.KeyPrefix()), func() { cli.Close() }
}

// tenantService selects the tenant service for the CLI: bypass when no --tenant
// is given, otherwise a static service pinned to the supplied tenant id.
func tenantService() tenant.Service {
	if tenantID == "" {
		return tenant.NewBypassService()
	}
	return tenant.NewStaticService(tenantID)
}
