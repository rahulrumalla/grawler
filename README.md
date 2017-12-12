# Grawler
A concurrent web crawler that crawls the website for internal links and static assets

## Makefile
Use the makefile

`makefile help` 
This will show you the available options to use it. If on windows, use cygwin command-line

NOTE: This library uses [cobra](https://github.com/spf13/cobra) for CLI wire-up

## Usage
1. Run `make install-dev-deps` to install all the dependencies
2. Run `make install` to install grawler CLI
3. Run `gralwer --help` to see for flags and usage
4. Run `grawler https://monzo.com -i -w 4` to run the program against a website (see --help for info on option)

## TODO
1. Incease test coverage
2. Include checking robots.txt
3. Include checking sitemap.xml when present
4. Implement trie data structure for quicker traversal and keeping track of navigation length between pages
5. Configure-drive the verbose level logging