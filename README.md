
![Logo]()


# V2ray Bot

A cli tools for v2ray.

## Features

- Parse v2ray config

## Installation

1. download the last release
2. add `v2ray-bot` to your path
3. set up autocompletion (zsh)

```shell
# Linux
$ v2ray-bot completion zsh > "${fpath[1]}/_v2ray-bot"

# macOS
$ v2ray-bot completion zsh > $(brew --prefix)/share/zsh/site-functions/_v2ray-bot
```

## Usage/Examples

### download cmd

`download` cmd is just a convenience method for downloading geo resources.  
Download predefined geo resources by specifying the `--all` flag, including
[geoip.dat](https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat), 
[geosite.dat](https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat).

```shell
$ v2ray-bot download --all
```

For well-known reasons, you may fail to download, you can specify the `--proxy` flag to use http proxy.

```shell
$ v2ray-bot download --all --proxy http://127.0.0.1:7891
```

### parse cmd

`parse` is used to parse the v2ray subscription link into human-readable json format.

```shell
$ v2ray-bot parse --from-file v2ray.share -o vmess.txt
```

If you only want to select part of the vmess configuration, you can use the `--filter` flag.  
Its principle is to search for the names of those vmess configurations.  
The vmess configuration format can be referred to 
[here](https://github.com/2dust/v2rayN/wiki/%E5%88%86%E4%BA%AB%E9%93%BE%E6%8E%A5%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E(ver-2)).

```shell
$ v2ray-bot parse --from-file ~/workspace/v2ray-bot/config/v2ray.share --filter 香港 -o vmess.txt
```

### ping cmd

`ping` is a simple way to test latency. Its principle is to send an http get request  
to the specified domain name. This command needs to specify the vmess configuration  
file (which can be obtained from parse), and then the sorted results will be output.

```shell
$ v2ray-bot ping https://www.google.com --inbound-port 7893
```

I think the `--set-fastest` flag is useful, it configures the v2ray outbound to the  
fastest vmess configuration. At the same time you must set the `--external-v2ray` flag  
to use an existing v2ray instance, so that save the fastest vmess configuration to this  
v2ray instance. Then again, you have to configure `dokodemo-door` inbound protocol  
in the v2ray instance, only then v2ray-bot can control the external v2ray instance,  
you can refer to [here](https://www.v2fly.org/config/protocols/dokodemo.html#inboundconfigurationobject).

```shell
$ v2ray-bot ping https://www.google.com --external-v2ray --dokodemo-door-addr 127.0.0.1:10085 --set-fastest --vmess-file vmess.txt
```

## License

[MIT](./LICENSE)

