# gauto-refresh
Automatically refresh your web pages when saving files

![Gauto Refresh demo](gauto-refresh-demo.gif)

## Install
Install go (sometimes named "golang") and then:
```sh
go install github.com/OnitiFR/gauto-refresh@latest
```

## Help
```
  -a string
    	custom action (default "location.reload()")
  -c	display a conditional script sample
  -d	debug mode
  -f value
    	file to watch (mutliple -f accepted, default = current dir)
  -p int
    	listening port (default 8888)
  -t int
    	delay in milliseconds before reload, for double-reload prevention or build/upload time (default 50)
  -v	show version

```

- Ignores: `.git, .svn, node_modules, vendor`

## TODO
- detect new folders and automatically add them to the watch list
- allow user to specify a list of ignored folders
