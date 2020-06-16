
Lemonade
========

remote...lemote...lemode...vim...Vimonade!!! :lemon:

Vimonade is a remote utility tool specifically for (neo)vim.
(copy, paste and open browser) over HTTP/2.

[![Build Status](https://travis-ci.org/jrc2139/vimonade.svg?branch=master)](https://travis-ci.org/jrc2139/vimonade)


Installation
------------

```sh
go get -d github.com/jrc2139/vimonade
cd $GOPATH/src/github.com/jrc2139/vimonade/
make install
```

Or download from [latest release](https://github.com/jrc2139/vimonade/releases/latest)


Example of use
----------------

![Example](http://f.st-hatena.com/images/fotolife/P/Pocke/20150823/20150823173041.gif)

For example, you use a Linux as a virtual machine on Windows host.
You connect to Linux by SSH client(e.g. PuTTY).
When you want to copy text of a file on Linux to Windows, what do you do?
One solution is doing `cat file.txt` and drag displayed text.
But this answer is NOT elegant! Because your hand leaves from the keyboard to use the mouse.

Another solution is using the Lemonade.
You input `cat file.txt | vimonade copy`. Then, vimonade copies text of the file to clipboard of the Windows!

In addition to the above, vimonade supports pasting and opening URL.


Usage
--------

```sh
Usage: vimonade [options]... SUB_COMMAND [arg]
Sub Commands:
  open [URL]                  Open URL by browser
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
