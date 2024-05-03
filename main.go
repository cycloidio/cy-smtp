package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cfile    = "config-file"
	skiptls  = "email-tls-skip-verify"
	server   = "email-smtp-svr-addr"
	username = "email-smtp-username"
	password = "email-smtp-password"
	from     = "email-addr-from"
	to       = "email-addr-to"
)

var (
	cfileFlag    string
	skiptlsFlag  bool
	serverFlag   string
	usernameFlag string
	passwordFlag string
	fromFlag     string
	toFlag       string
)

var RootCmd = &cobra.Command{
	Use:   "cy-smtp",
	Short: "Send an email using Cycloid config",
	Long:  "Send an email using Cycloid config file in order to test different SMTP servers integration",
	RunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetConfigFile(cfileFlag)
		err := viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("error reading config file: %w", err)
		}
		fmt.Printf("viper: %+v\n", viper.GetString(server))
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
	RootCmd.Flags().StringVarP(&cfileFlag, cfile, "c", "config.yaml", "The configuration file PATH.")
	viper.BindPFlag(cfile, RootCmd.Flags().Lookup(cfile))

	RootCmd.Flags().StringVarP(&serverFlag, server, "s", "", "SMTP server address (host:port)")
	viper.BindPFlag(server, RootCmd.Flags().Lookup(server))

	RootCmd.Flags().StringVarP(&usernameFlag, username, "u", "", "Username for authenticating the connections to the SMTP server")
	viper.BindPFlag(username, RootCmd.Flags().Lookup(username))

	RootCmd.Flags().StringVarP(&passwordFlag, password, "p", "", "Password for authenticating the connections to the SMTP server")
	viper.BindPFlag(password, RootCmd.Flags().Lookup(password))

	RootCmd.Flags().StringVarP(&fromFlag, from, "f", "", "From email address to use for any sent email")
	viper.BindPFlag(from, RootCmd.Flags().Lookup(from))

	RootCmd.Flags().BoolVar(&skiptlsFlag, skiptls, true, "Skip client TLS certs verification")
	viper.BindPFlag(from, RootCmd.Flags().Lookup(from))

	RootCmd.Flags().StringVarP(&toFlag, to, "t", "", "send test email to this address")
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
	ret := config{
		cfgFile:  cfileFlag,
		skiptls:  skiptlsFlag,
		server:   serverFlag,
		username: usernameFlag,
		password: passwordFlag,
		from:     fromFlag,
		to:       toFlag,
	}
	if ret.cfgFile == "" {
		ret.cfgFile = viper.GetString(cfile)
	}
	if ret.skiptls == false {
		ret.skiptls = viper.GetBool(skiptls)
	}
	if ret.server == "" {
		ret.server = viper.GetString(server)
	}
	if ret.username == "" {
		ret.username = viper.GetString(username)
	}
	if ret.password == "" {
		ret.password = viper.GetString(password)
	}
	if ret.from == "" {
		ret.from = viper.GetString(from)
	}
	if ret.to == "" {
		ret.to = viper.GetString(to)
	}
	return ret
}

func sendEmail(cfg config) error {
	from, err := mail.ParseAddress(cfg.from)
	if err != nil {
		return fmt.Errorf("error sender address: %w", err)
	}
	to, err := mail.ParseAddress(cfg.to)
	if err != nil {
		return fmt.Errorf("error recipient address: %w", err)
	}

	msg, err := formatEmail(from, to)
	if err != nil {
		return fmt.Errorf("error formatting email: %w", err)
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.skiptls,
	}
	fmt.Print("STARTTLS... ")
	client, err := smtp.DialStartTLS(cfg.server, tlsCfg)
	if err != nil {
		return fmt.Errorf("error connecting to server: %w", err)
	}
	fmt.Println("done!")
	if ok, authExt := client.Extension("AUTH"); ok {
		fmt.Printf("AUTH: %s... ", authExt)
		var auth sasl.Client
		switch authExt {
		case sasl.Plain:
			auth = sasl.NewPlainClient("", cfg.username, cfg.password)
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("error authenticating to server: %w", err)
			}
		case sasl.Login:
			auth = sasl.NewLoginClient(cfg.username, cfg.password)
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("error authenticating to server: %w", err)
			}
		}
		fmt.Println("done!")
	}
	fmt.Print("Sending test email... ")
	err = client.SendMail(from.Address, []string{to.Address}, &msg)
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	fmt.Println("done!")

	return nil
}

func formatEmail(fromAddrs, toAddrs *mail.Address) (bytes.Buffer, error) {
	var b bytes.Buffer

	from := []*mail.Address{fromAddrs}
	to := []*mail.Address{toAddrs}

	// Create our mail header
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("To", to)

	// Create a new mail writer
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		return b, fmt.Errorf("error creating email writer: %w", err)
	}

	// Create a text part
	tw, err := mw.CreateInline()
	if err != nil {
		return b, fmt.Errorf("error creating email InlineWriter: %w", err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		return b, fmt.Errorf("error creating email part: %w", err)
	}
	io.WriteString(w, "Hello from cy-smtp!\nThis is a test message.")
	w.Close()
	tw.Close()

	mw.Close()

	return b, nil

}
