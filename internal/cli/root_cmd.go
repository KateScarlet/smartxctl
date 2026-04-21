package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	a := &app{}
	rootCmd := &cobra.Command{
		Use:           "smartxctl",
		Short:         "CLI for SmartX CloudTower",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&a.configFile, "config", "", "Path to the config file")
	rootCmd.PersistentFlags().StringVarP(&a.output, "output", "o", "table", "Output format: table or json")
	rootCmd.AddCommand(buildGetCmd(a))
	rootCmd.AddCommand(buildDescribeCmd(a))
	rootCmd.AddCommand(buildCreateCmd(a))
	rootCmd.AddCommand(buildDeleteCmd(a))
	rootCmd.AddCommand(buildStartCmd(a))
	rootCmd.AddCommand(buildStopCmd(a))
	rootCmd.AddCommand(buildRestartCmd(a))
	rootCmd.AddCommand(buildConfigCmd(a))
	rootCmd.AddCommand(buildVersionCmd())

	return rootCmd
}

func buildVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return renderJSON(cmd.OutOrStdout(), map[string]string{
				"version":   version,
				"buildTime": buildTime,
			})
		},
	}
}
