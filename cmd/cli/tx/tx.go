package tx

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/httpx"

	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// TxCmd represents the Tx command
var TxCmd = &cobra.Command{
	Use:   "tx",
	Short: "Get tx",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || args[0] == "" {
			logger.Error("Tx", errorx.ErrInvalidArgument, logger.Params{})
			os.Exit(1)
		}

		if !tx.IsID(args[0]) {
			logger.Error("Block", errorx.ErrInvalidArgument, logger.Params{})
			os.Exit(1)
		}

		resp, err := httpx.GET(fmt.Sprintf("%s/api/tx/%s", viper.GetString("host"), args[0]), nil)
		if err != nil {
			logger.Error("bitgodine-cli", err, logger.Params{})
			os.Exit(1)
		}

		var t tx.Tx
		if err := json.Unmarshal([]byte(resp), &t); err != nil {
			logger.Error("bitgodine-cli", err, logger.Params{})
			os.Exit(1)
		}

		fmt.Println("##################################")
		fmt.Println("Header:")
		table := tablewriter.NewWriter(os.Stdout)
		table.Append([]string{"ID", t.TxID})
		table.Append([]string{"Version", fmt.Sprint(t.Version)})
		table.Append([]string{"Locktime", fmt.Sprint(t.Locktime)})
		table.Append([]string{"Size", fmt.Sprint(t.Size)})
		table.Render()

		fmt.Println("##################################")
		fmt.Println("Inputs:")
		for _, in := range t.Vin {
			table := tablewriter.NewWriter(os.Stdout)
			table.Append([]string{"ID", in.TxID})
			table.Append([]string{"Vout", fmt.Sprint(in.Vout)})
			table.Append([]string{"Is coinbase", fmt.Sprint(in.IsCoinbase)})
			scriptsig := in.Scriptsig[:len(in.Scriptsig)/2] + "\n" + in.Scriptsig[len(in.Scriptsig)/2:]
			table.Append([]string{"Scriptsig", scriptsig})
			table.Render()
		}

		fmt.Println("##################################")
		fmt.Println("Outputs:")
		for _, out := range t.Vout {
			table := tablewriter.NewWriter(os.Stdout)
			table.Append([]string{"Value", fmt.Sprint(out.Value)})
			table.Append([]string{"Index", fmt.Sprint(out.Index)})
			table.Append([]string{"Address", out.ScriptpubkeyAddress})
			table.Append([]string{"Type", out.ScriptpubkeyType})
			scriptpubkey := out.Scriptpubkey[:len(out.Scriptpubkey)/2] + "\n" + out.Scriptpubkey[len(out.Scriptpubkey)/2:]
			table.Append([]string{"Scriptpubkey", scriptpubkey})
			table.Render()
		}
	},
}

func init() {
	TxCmd.AddCommand(lsCmd)
	TxCmd.AddCommand(rmCmd)
}
