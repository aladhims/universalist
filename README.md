----
## Universalist
Universalist is an IDE agnostic annotation-highlighter for annotations like TODO, FIXME or even your own custom annotations.

----
## Installation
From the source:

    $ go get github.com/aladhims/universalist/cmd/universalist

----
## Usage

In your current working directory: 

*annotations will be in color if specified*

    $ universalist
    TODO
      - Complete this function                    main.java:87
      - Make the comment below clearer            index.js
     
    FIXME
      - Improve the algorithm                     utils.go:25

Specify the directory:

    $ universalist --path path/to/your/workdir

Indent the output:

    $ universalist --indent

Using custom configurations (see [config-example.json](https://github.com/aladhims/universalist/blob/master/config-sample.json)):

    $ universalist --config ./mycustomconfig.json

----
## Supported Colors

* black
* red
* green
* yellow
* blue
* magenta
* cyan
* white
* 0...255 (256 colors)

For more information, see [here](https://github.com/mgutz/ansi)
