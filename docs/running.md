# Other ways to get and run punkbot

Punkbot has binary releases and can also be built from source code.

## Binaries 

Punkbot now supports binaries for Linux and MacOS for both AMD64 and ARM64
architectures. Binaries can be downloaded from
[releases](https://github.com/mr-stringer/punkbot/releases). If you require a
binary for another platform you can raise a an issue or compile it yourself.

## How to compile

To compile you'll need a working go environment (check go.mod for the correct
version) and automake. You can run (the admittedly sparse) tests with:

```shell
make test
```

Building the binary can be done with:

```shell
make
```

## Running

To run, you'll need

* A Bluesky account
* An app password for your bluesky account
* A completed configuration file

Details of these can be found in the main [readme](../README.md)