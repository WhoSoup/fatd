// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

var fat0Tx fat0.Transaction

// transactFAT0Cmd represents the FAT0 command
var transactFAT0Cmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fat0",
		Aliases: []string{"fat-0", "FAT0", "FAT-0"},
		Short:   "Send or distribute FAT-0 tokens",
		Long: `
`[1:],
		Run: func(_ *cobra.Command, _ []string) {},
	}
	transactCmd.AddCommand(cmd)
	transactCmplCmd.Sub["fat0"] = transactFAT0CmplCmd
	rootCmplCmd.Sub["help"].Sub["transact"].Sub["fat0"] = complete.Command{}

	flags := cmd.Flags()
	flags.VarPF((*AddressAmountMap)(&fat0Tx.Inputs), "input", "i", "").DefValue = ""
	flags.VarPF((*AddressAmountMap)(&fat0Tx.Outputs), "output", "o", "").DefValue = ""

	generateCmplFlags(cmd, transactFAT0CmplCmd.Flags)
	return cmd
}()

var PredictFAAddressesColon = PredictAppend(PredictFAAddresses, ":")

var transactFAT0CmplCmd = complete.Command{
	Flags: mergeFlags(apiCmplFlags, tokenCmplFlags,
		ecAdrCmplFlags, complete.Flags{
			"--input":  PredictFAAddressesColon,
			"-i":       PredictFAAddressesColon,
			"--output": PredictFAAddressesColon,
			"-o":       PredictFAAddressesColon,
		}),
}

var privateAddress = map[factom.FAAddress]factom.FsAddress{}
var addressValueStrMap = map[factom.FAAddress]string{}

type AddressAmountMap fat0.AddressAmountMap

func (m *AddressAmountMap) Set(adrAmtStr string) error {
	if *m == nil {
		*m = make(AddressAmountMap)
	}
	return m.set(adrAmtStr)
}
func (m AddressAmountMap) set(data string) error {
	// Split address from amount.
	strs := strings.Split(data, ":")
	if len(strs) != 2 {
		return fmt.Errorf("invalid format")
	}
	adrStr := strs[0]
	amountStr := strs[1]

	// Parse address, which could be FA or Fs or the keyword "coinbase" or
	// "burn"
	var fa factom.FAAddress
	var fs factom.FsAddress
	switch adrStr {
	case "coinbase", "burn":
		fa = fat.Coinbase()
	default:
		// Attempt to parse as FAAddress first
		if err := fa.Set(adrStr); err != nil {
			// Not FA, try FsAddress...
			if err := fs.Set(adrStr); err != nil {
				return fmt.Errorf("invalid address: %v", err)
			}
			fa = fs.FAAddress()
			if fa != fat.Coinbase() {
				// Save private addresses for future use.
				privateAddress[fa] = fs
			}
		}
	}
	if _, ok := m[fa]; ok {
		return fmt.Errorf("duplicate address")
	}

	// Parse amount
	amount, err := parsePositiveInt(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	m[fa] = amount
	addressValueStrMap[fa] = amountStr

	return nil
}
func (m AddressAmountMap) String() string {
	return fmt.Sprintf("%v", fat0.AddressAmountMap(m))
}
func (AddressAmountMap) Type() string {
	return "<FA | Fs>:<amount>"
}

func parsePositiveInt(intStr string) (uint64, error) {
	if len(intStr) == 0 {
		return 0, fmt.Errorf("empty")
	}
	// Parse amount
	amount, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if amount == 0 {
		return 0, fmt.Errorf("zero")
	}
	if amount < 0 {
		return 0, fmt.Errorf("negative")
	}
	return uint64(amount), nil
}
