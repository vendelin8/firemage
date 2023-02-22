package internal

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	confPath string
	keyPath  string
	verbose  bool
	useEmu   bool
	lgr      *zap.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "firemage",
	Short: shortDesc,
	Long:  longDesc,
}

func init() {
	rootCmd.Flags().StringVarP(&keyPath, "key", "k", "service-account.json", descKey)
	rootCmd.Flags().StringVarP(&confPath, "conf", "c", "conf.yaml", descConf)
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, descDebug)
	rootCmd.Flags().BoolVarP(&useEmu, "emulator", "e", false, descEmul)
}

func Main() {
	must("execute command", rootCmd.Execute())
	if logSync := initLogger(); logSync != nil {
		defer logSync()
	}
	fb = NewFirebase()
	fe = createGUI()
	fe.run()
	cancelF()
}

func initLogger() func() {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"log.txt"}
	if !verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	var err error
	lgr, err = cfg.Build()
	must("build logger", err)
	zap.ReplaceGlobals(lgr)
	return func() {
		must("log sync", lgr.Sync())
	}
}

// must panics if the given error is not nil, with the given description. For initializations only.
func must(descr string, err error) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", descr, err))
	}
}
