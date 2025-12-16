package main

import (
	"github.com/spf13/cobra"
	"github.com/vendelin8/firemage/internal/api"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/conf"
	"github.com/vendelin8/firemage/internal/firebase"
	"github.com/vendelin8/firemage/internal/frontend"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/log"
	"github.com/vendelin8/firemage/internal/util"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "firemage",
	Short: lang.ShortDesc,
	Long:  lang.LongDesc,
}

func init() {
	rootCmd.Flags().StringVarP(&conf.KeyPath, "key", "k", "service-account.json", lang.DescKey)
	rootCmd.Flags().StringVarP(&conf.ConfPath, "conf", "c", "conf.yaml", lang.DescConf)
	rootCmd.Flags().StringVarP(&log.LogPath, "log", "l", "log.txt", lang.DescLog)
	rootCmd.Flags().BoolVarP(&log.Verbose, "verbose", "v", false, lang.DescDebug)
	rootCmd.Flags().BoolVarP(&conf.UseEmu, "emulator", "e", false, lang.DescEmul)
}

func main() {
	api.InitMenu()
	log.Must("initialize timed buttons map from custom/custom.txt", util.InitializeTimedButtonsMap())
	log.Must("execute command", rootCmd.Execute())

	if logSync := log.Init(); logSync != nil {
		defer logSync()
	}
	common.Fb = firebase.New()
	common.Fe = frontend.CreateGUI()
	common.Fe.Run()
}
