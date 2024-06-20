package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/cli"
	cmtflags "github.com/cometbft/cometbft/libs/cli/flags"
	"github.com/cometbft/cometbft/libs/log"
)

var (
	config = cfg.DefaultConfig()
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

func init() {
	registerFlagsRootCmd(RootCmd)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log_level", config.LogLevel, "log level")
}

// ParseConfig retrieves the default environment configuration,
// sets up the CometBFT root and ensures that the root exists
func ParseConfig(cmd *cobra.Command) (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	var home string
	if os.Getenv("CMTHOME") != "" {
		home = os.Getenv("CMTHOME")
	} else if os.Getenv("TMHOME") != "" {
		// XXX: Deprecated.
		home = os.Getenv("TMHOME")
		logger.Error("Deprecated environment variable TMHOME identified. CMTHOME should be used instead.")
	} else {
		home, err = cmd.Flags().GetString(cli.HomeFlag)
		if err != nil {
			return nil, err
		}
	}

	conf.RootDir = home

	conf.SetRoot(conf.RootDir)
	cfg.EnsureRoot(conf.RootDir)
	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}
	if warnings := conf.CheckDeprecated(); len(warnings) > 0 {
		for _, warning := range warnings {
			logger.Info("deprecated usage found in configuration file", "usage", warning)
		}
	}

	if conf.P2P.SameRegion {
		// If SameRegion is set, we need to populate our region with the region of the node
		myRegion, err := getOwnRegion()
		if err != nil {
			return nil, err
		}
		conf.P2P.MyRegion = myRegion

		// Make sure that the MaxPercentPeersInSameRegion does not exceed some hard coded value.
		// If it does, replace it with the max
		if conf.P2P.MaxPercentPeersInSameRegion > 0.9 {
			conf.P2P.MaxPercentPeersInSameRegion = cfg.DefaultMaxPercentPeersInSameRegion
		}
	}

	return conf, nil
}

type ipInfo struct {
	Status  string
	Country string
}

func getOwnRegion() (string, error) {
	// TODO: Add fallbacks
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	var ipInfo ipInfo
	json.Unmarshal(body, &ipInfo)
	fmt.Println("ipInfoOwn", ipInfo)

	if ipInfo.Status != "success" {
		return "", fmt.Errorf("failed to get own region")
	}

	return ipInfo.Country, nil
}

// RootCmd is the root command for CometBFT core.
var RootCmd = &cobra.Command{
	Use:   "cometbft",
	Short: "BFT state machine replication for applications in any programming languages",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}

		config, err = ParseConfig(cmd)
		if err != nil {
			return err
		}

		if config.LogFormat == cfg.LogFormatJSON {
			logger = log.NewTMJSONLogger(log.NewSyncWriter(os.Stdout))
		}

		logger, err = cmtflags.ParseLogLevel(config.LogLevel, logger, cfg.DefaultLogLevel)
		if err != nil {
			return err
		}

		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}

		logger = logger.With("module", "main")
		return nil
	},
}
