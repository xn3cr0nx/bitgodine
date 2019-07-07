package tag

import (
	// "fmt"
	"errors"
	"os"

	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/postgres"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var address, tag string

// getCmd launch get process of address tags
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get tag from db based on address or tag name",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetString("tag.address") == "" && viper.GetString("tag.name") == "" {
			logger.Error("Tag get", errors.New("Missing tag specification, you have to indicate --address or --name to retrieve tag"), logger.Params{})
			os.Exit(-1)
		}

		pg, err := postgres.NewPg(&postgres.Config{
			Host: "localhost",
			Port: 5432,
			Pass: "bitgodine",
			User: "bitgodine",
			Name: "bitgodine",
		})
		if err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		if err := pg.Connect(); err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}
		defer pg.Close()
		if err := pg.Migration(); err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		var tags []postgres.Tag

		var step *gorm.DB
		if viper.GetString("tag.address") != "" && viper.GetString("tag.name") != "" {
			step = pg.DB.Where("address = ? AND tag = ?", viper.GetString("tag.address"), viper.GetString("tag.name"))
		} else if viper.GetString("tag.address") != "" {
			step = pg.DB.Where("address = ?", viper.GetString("tag.address"))
		} else {
			step = pg.DB.Where("tag = ?", viper.GetString("tag.name"))
		}
		if res := step.Find(&tags); res.Error != nil {
			logger.Error("Tag get", res.Error, logger.Params{})
			os.Exit(-1)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Address", "Tag", "Meta", "Verified"})
		table.SetBorder(false)

		red := color.New(color.FgRed)
		green := color.New(color.FgGreen)
		for _, tag := range tags {
			if tag.Verified == true {
				table.Append([]string{tag.Address, tag.Tag, tag.Meta, green.Sprint("✓")})
			} else {
				table.Append([]string{tag.Address, tag.Tag, tag.Meta, red.Sprint("x")})
			}
		}

		table.Render()
	},
}

func init() {

	getCmd.PersistentFlags().StringVar(&address, "address", "", "Get tags by address")
	viper.SetDefault("tag.address", "")
	viper.BindPFlag("tag.address", getCmd.PersistentFlags().Lookup("address"))

	getCmd.PersistentFlags().StringVar(&tag, "name", "", "Get tags by name")
	viper.SetDefault("tag.name", "")
	viper.BindPFlag("tag.name", getCmd.PersistentFlags().Lookup("name"))
}
