package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cfg      = "config-file"
	skiptls  = "email-tls-skip-verify"
	server   = "email-smtp-svr-addr"
	username = "email-smtp-username"
	password = "email-smtp-password"
	from     = "email-addr-from"
	to       = "email-addr-to"
)

var RootCmd = &cobra.Command{
	Use:   "cy-smtp",
	Short: "Send an email using Cycloid config",
	Long:  "Send an email using Cycloid config file in order to test different SMTP servers integration",
	RunE: func(_ *cobra.Command, _ []string) error {
		viper.SetConfigFile(viper.GetString(cfg))
		err := viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
		cfg := getConfig()
		fmt.Printf("CONFIG: %+v\n", cfg)
		err = sendEmail(cfg)
		if err != nil {
			return err
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.Flags().StringP(cfg, "c", "config.yaml", "The configuration file PATH.")
	viper.BindPFlag(cfg, RootCmd.Flags().Lookup(cfg))

	RootCmd.Flags().StringP(server, "s", "", "SMTP server address (host:port)")
	viper.BindPFlag(server, RootCmd.Flags().Lookup(server))

	RootCmd.Flags().StringP(username, "u", "", "Username for authenticating the connections to the SMTP server")
	viper.BindPFlag(username, RootCmd.Flags().Lookup(username))

	RootCmd.Flags().StringP(password, "p", "", "Password for authenticating the connections to the SMTP server")
	viper.BindPFlag(password, RootCmd.Flags().Lookup(password))

	RootCmd.Flags().StringP(from, "f", "", "From email address to use for any sent email")
	viper.BindPFlag(from, RootCmd.Flags().Lookup(from))

	RootCmd.Flags().Bool(skiptls, true, "Skip client TLS certs verification")
	viper.BindPFlag(from, RootCmd.Flags().Lookup(from))

	RootCmd.Flags().StringP(to, "t", "", "send test email to this address")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

type config struct {
	cfgFile  string
	skiptls  bool
	server   string
	username string
	password string
	from     string
	to       string
}

func getConfig() config {
	return config{
		cfgFile:  viper.GetString(cfg),
		skiptls:  viper.GetBool(skiptls),
		server:   viper.GetString(server),
		username: viper.GetString(username),
		password: viper.GetString(password),
		from:     viper.GetString(from),
		to:       viper.GetString(to),
	}
}

func sendEmail(cfg config) error {
	msg := strings.NewReader("Hello from cy-smtp!\nThis is a test message.")

	log.Println("Sending test email ")
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.skiptls,
	}
	client, err := smtp.DialStartTLS(cfg.server, tlsCfg)
	if err != nil {
		return fmt.Errorf("error connecting to server: %w", err)
	}
	auth := sasl.NewPlainClient("", cfg.username, cfg.password)
	if ok, _ := client.Extension("AUTH"); ok {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("error authenticating to server: %w", err)
		}
	}
	err = client.SendMail(cfg.from, []string{cfg.to}, msg)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}
