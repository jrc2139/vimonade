
Vimonade
========

remote...lemote...lemode...vim...Vimonade!!! :lemon:

Vimonade is a remote utility tool specifically for Neovim.
(copy, paste) over HTTP/2 forked from [Lemonade](https://github.com/lemonade-command/lemonade).

[![Build Status](https://travis-ci.org/jrc2139/vimonade.svg?branch=master)](https://travis-ci.org/jrc2139/vimonade)


Installation
------------

```sh
go get -d github.com/jrc2139/vimonade
cd $GOPATH/src/github.com/jrc2139/vimonade/
make install
```

Or download from [latest release](https://github.com/jrc2139/vimonade/releases/latest)


Configuration
------------


You must edit `runtime/autoload/provider/clipboard.vim` to include `vimonade` as an executable to find.


Usage
--------

```sh
Usage: vimonade [options]... SUB_COMMAND [arg]
Sub Commands:
  copy [text]                 Copy text.
  paste                       Paste text.
  server                      Start vimonade server.

Options:
  --port=2489                 TCP port number
  --line-ending               Convert Line Ending(CR/CRLF)
  --allow="0.0.0.0/0,::/0"    Allow IP Range                [Server only]
  --host="localhost"          Destination hostname          [Client only]
  --no-fallback-messages      Do not show fallback messages [Client only]
  --trans-loopback=true       Translate loopback address    [open subcommand only]
  --trans-localfile=true      Translate local file path     [open subcommand only]
  --help                      Show this message
```




Links
-------

- https://github.com/lemonade-command/lemonade
- https://speakerdeck.com/pocke/remote-utility-tool-lemonade
- [リモートのPCのブラウザやクリップボードを操作するツール Lemonade を作った - pockestrap](http://pocke.hatenablog.com/entry/2015/07/04/235118)
- [リモートユーティリティーツール、Lemonade v0.2.0 をリリースした - pockestrap](http://pocke.hatenablog.com/entry/2015/08/23/221543)
- [lemonade v1.0.0をリリースした - pockestrap](http://pocke.hatenablog.com/entry/2016/04/19/233423)
