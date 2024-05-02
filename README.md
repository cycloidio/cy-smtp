# cy-smtp

This is a small binary to test the connection to a SMTP server.
If all works well, running `cy-smtp` will send to the specified recipient this test email:

```
Hello from cy-smtp!
This is a test message.
```

You can pass all the needed parameters on the command line, in this case check the help of the command, `cy-smtp --help`.
Or the `config.yaml` file to configure the behavior, the parameters in the YAML file are mostly compatible with the ones used by Cycloid.
There are two addition: `email-tls-skip-verify` and `email-addr-to`. If you are using the Cycloid config file these two parameters must be passed on the cmd line or added int the config file.


## Quick start

```
wget https://github.com/cycloidio/cy-smtp/releases/download/v0.2.1/cy-smtp-v0.2.1-linux-amd64.tar.gz
tar xf cy-smtp-v*-linux-amd64.tar.gz

./cy-smtp --config-file /opt/config.yml
```
