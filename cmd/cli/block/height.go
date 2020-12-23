package block

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/httpx"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// heightCmd represents the height command
var heightCmd = &cobra.Command{
	Use:   "height",
	Short: "Show the height of the last block stored",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := httpx.GET(fmt.Sprintf("%s/api/blocks/tip/height", viper.GetString("host")), nil)
		if err != nil {
			logger.Error("bitgodine-cli", err, logger.Params{})
			os.Exit(1)
		}
		fmt.Println(resp)
	},
}
