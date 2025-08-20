# Command line arguments

Punkbot supports several command line arguments:

| Flag | Argument | Description |
|:-----|:---------|:------------|
| -h | none | Print help and quit |
| -v | none | Print version and quit |
| -f | path to configuration file | Specifies the location of configuration file, overrides the default |
| -o | path to log output file | Specifies the location of the file to use for logging, default logging is to screen |
| -j | none | Switches logging mode to JSON, default is text |
| -l | string | Sets the logging level can be set to: `err`, `warn`, `info`(default) or `debug` | 
| -p | bool | When set to `true` and logging is set to `debug`, this will log the text contained in every post processed and is very noisy. Default is `false` | 

None of the command line arguments are mandatory.

Below is an example of running punkbot with `warn` level logging with a
non-standard configuration file:

```shell
./punkbot -l warn -f /tmp/myPbConfig.yml
```