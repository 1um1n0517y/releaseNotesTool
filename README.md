## Synopsis

Release notes tool updates Confluence page with release information for given: 

1) SVN path to game or component

2) parent page name on Confluence

3) space Key on Confluence

4) version that is released
 
First check if parent page name is Icy Wilds. If so, then svn path for tags and trunk is in "/standard/", else path is in "/standard/imp/". This paths are relative to given svnPath of the game(component).

After that gets revision numbers between given version number and previous one. If version is in format x.x.0 then reads all trunk history log, else finds revision number from previos version and reads history log between given version and previous version.

Readed history log is parsed into tables and appended on the top of content of the child page of given parent page. Child page name is always equal to "parentPageName Releases".

## Code Example

```
Usage:
  releaseNotesTool [flags]

Flags:
  -p, --parentPage string   Parent page name.
  -k, --spaceKey string     Confluence space key. (default "GAMBG")
  -s, --svnPath string      Path to svn location.
  -v, --version string      New release version.
```

> releaseNotesTool -p "Icy Wilds" -s "http://svn.g2-networks.net/svn/instantgames/games/icyWilds" -v "1.0.58"

This command will update Icy Wilds Releases page on Confluence in space GAMBG, adding two new tables for release version 1.0.58 with all commits related to this version.


## Installation

Requirements for building tool:
1) Go language  (tool is written in Go)
2) Git (for getting dependencies)
3) Svn  (for getting source code)

*For newer versions of git run next command*
>*git config --global http.https://gopkg.in.followRedirects true*

First checkout source code from svn:
>svn checkout http://svn.g2-networks.net/svn/instantgames/ops/tools/releaseNotesTool/

Then, inside project folder run:
>go get

This will install all dependencies. After that run:
>go build


## Contributors

<mladen.popadic@igt.com>

<nikola.djokic@igt.com>

<milos.stamenkovic@igt.com>

