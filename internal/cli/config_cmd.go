package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Fission-AI/openspec-go/internal/core/globalconfig"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage OpenSpec global configuration",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Show global config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(globalconfig.GetGlobalConfigPath())
			return nil
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := globalconfig.GetGlobalConfig()
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(cfg)
		},
	})

	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := globalconfig.GetGlobalConfig()
			key := args[0]
			switch key {
			case "profile":
				fmt.Println(cfg.Profile)
			case "delivery":
				fmt.Println(cfg.Delivery)
			case "workflows":
				if cfg.Workflows != nil {
					fmt.Println(strings.Join(cfg.Workflows, ", "))
				}
			default:
				if cfg.FeatureFlags != nil {
					if v, ok := cfg.FeatureFlags[key]; ok {
						fmt.Println(v)
						return nil
					}
				}
				return fmt.Errorf("unknown config key: %s", key)
			}
			return nil
		},
	}
	configCmd.AddCommand(getCmd)

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := globalconfig.GetGlobalConfig()
			key, value := args[0], args[1]
			switch key {
			case "profile":
				cfg.Profile = globalconfig.Profile(value)
			case "delivery":
				cfg.Delivery = globalconfig.Delivery(value)
			default:
				return fmt.Errorf("unknown config key: %s. Valid: profile, delivery", key)
			}
			if err := globalconfig.SaveGlobalConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("Set %s = %s\n", key, value)
			return nil
		},
	}
	configCmd.AddCommand(setCmd)

	unsetCmd := &cobra.Command{
		Use:   "unset <key>",
		Short: "Remove a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := globalconfig.GetGlobalConfig()
			key := args[0]
			switch key {
			case "profile":
				cfg.Profile = globalconfig.ProfileCore
			case "delivery":
				cfg.Delivery = globalconfig.DeliveryBoth
			case "workflows":
				cfg.Workflows = nil
			default:
				if cfg.FeatureFlags != nil {
					delete(cfg.FeatureFlags, key)
				}
			}
			if err := globalconfig.SaveGlobalConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("Unset %s\n", key)
			return nil
		},
	}
	configCmd.AddCommand(unsetCmd)

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := globalconfig.GlobalConfig{
				FeatureFlags: map[string]bool{},
				Profile:      globalconfig.ProfileCore,
				Delivery:     globalconfig.DeliveryBoth,
			}
			if err := globalconfig.SaveGlobalConfig(cfg); err != nil {
				return err
			}
			fmt.Println("Configuration reset to defaults")
			return nil
		},
	}
	configCmd.AddCommand(resetCmd)

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Open config file in $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			configPath := globalconfig.GetGlobalConfigPath()
			// Ensure config file exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				cfg := globalconfig.GetGlobalConfig()
				_ = globalconfig.SaveGlobalConfig(cfg)
			}
			c := exec.Command(editor, configPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
	configCmd.AddCommand(editCmd)

	rootCmd.AddCommand(configCmd)
}
