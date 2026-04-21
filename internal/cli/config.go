package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/smartxworks/cloudtower-go-sdk/v2/client"
	User "github.com/smartxworks/cloudtower-go-sdk/v2/client/user"
	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func buildConfigCmd(a *app) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect local smartxctl configuration",
	}

	viewCmd := &cobra.Command{
		Use:   "view",
		Short: "Show the effective config with the password redacted",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := a.loadConfig()
			if err != nil {
				return err
			}
			out := map[string]any{
				"configFile": path,
				"tower": map[string]string{
					"host":   cfg.Tower.Host,
					"scheme": cfg.Tower.Scheme,
				},
				"user": map[string]string{
					"name":     cfg.User.Name,
					"password": redactSecret(cfg.User.Password),
					"source":   cfg.User.Source,
				},
			}
			return renderJSON(cmd.OutOrStdout(), out)
		},
	}

	configCmd.AddCommand(viewCmd)
	return configCmd
}

func (a *app) getClient() (*apiclient.Cloudtower, error) {
	cfg, _, err := a.loadConfig()
	if err != nil {
		return nil, err
	}

	var userSource models.UserSource
	if strings.EqualFold(cfg.User.Source, "LDAP") {
		userSource = models.UserSourceLDAP
	} else {
		userSource = models.UserSourceLOCAL
	}

	loginParams := User.NewLoginParams()
	loginParams.RequestBody = &models.LoginInput{
		Username: new(cfg.User.Name),
		Password: new(cfg.User.Password),
		Source:   new(userSource),
	}

	tlsClient, err := httptransport.TLSClient(httptransport.TLSClientOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create TLS client: %w", err)
	}

	transport := httptransport.NewWithClient(
		cfg.Tower.Host, "/v2/api", []string{strings.ToLower(cfg.Tower.Scheme)}, tlsClient)
	client := apiclient.New(transport, strfmt.Default)
	logRes, err := client.User.Login(loginParams)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", *logRes.Payload.Data.Token)
	return client, nil
}

func (a *app) loadConfig() (*config, string, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	var cfgPath string
	if a.configFile != "" {
		cfgPath = a.configFile
		v.SetConfigFile(cfgPath)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/smartxctl")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, "", fmt.Errorf("read config: %w", err)
	}

	if cfgPath == "" {
		cfgPath = v.ConfigFileUsed()
	}
	if absPath, err := filepath.Abs(cfgPath); err == nil {
		cfgPath = absPath
	}

	var cfg config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, "", fmt.Errorf("decode config: %w", err)
	}
	if err := validateConfig(cfg); err != nil {
		return nil, "", err
	}

	return &cfg, cfgPath, nil
}

func validateConfig(cfg config) error {
	if strings.TrimSpace(cfg.Tower.Host) == "" {
		return errors.New("config error: tower.host is required")
	}
	scheme := strings.ToLower(strings.TrimSpace(cfg.Tower.Scheme))
	if scheme != "http" && scheme != "https" {
		return errors.New("config error: tower.scheme must be http or https")
	}
	if strings.TrimSpace(cfg.User.Name) == "" {
		return errors.New("config error: user.name is required")
	}
	if strings.TrimSpace(cfg.User.Password) == "" {
		return errors.New("config error: user.password is required")
	}
	source := strings.ToUpper(strings.TrimSpace(cfg.User.Source))
	if source == "" {
		return errors.New("config error: user.source is required")
	}
	if source != "LOCAL" && source != "LDAP" {
		return errors.New("config error: user.source must be LOCAL or LDAP")
	}
	return nil
}
