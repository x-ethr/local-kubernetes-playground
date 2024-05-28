# `color` - ANSI Terminal Color(s)

The `color` package provides terminal escape sequences for giving life to console output.

![terminal-output-example](./.documentation/terminal-output.png)

## Overview

ANSI escape sequence processing allows capable output device(s) to display color.

While there exists other packages, `github.com/x-ethr/color` dynamically determines if the given output
device is capable of escaping. For usage with CI and production systems, `color` will automatically disable
adding color to user-provided value(s).

Additionally, users can optionally force color for their own testing purposes if necessary (for users implementing custom
loggers, for example).

## Documentation

Official `godoc` documentation (with examples) can be found at the [Package Registry](https://pkg.go.dev/github.com/x-ethr/color).

## Usage

###### Add Package Dependency

```bash
go get -u github.com/x-ethr/color
```

###### Import & Implement

`main.go`

```go
package main

import (
    "fmt"

    "github.com/x-ethr/color"
)

func main() {
    // --> Write the content "Default" out to stdout without color
    color.Color().Default("Default").Print() // Output: Default

    // --> Write the content "Red", "Blue", "Green" out to stdout with color escapes
    color.Color().Red("Red").Print() // Output: Red
    color.Color().Blue("Blue").Print() // Output: Blue
    color.Color().Green("Green").Print() // Output: Green

    // --> Wrap color(s) with bold color escapes and write to stdout
    color.Color().Bold(color.Color().Cyan("Cyan")).Print() // Output: Cyan

    // --> Customize how newlines and spaces are added to the formatted output
    color.Color().Red("Color-1").Print(func(o *color.Options) { o.Line = false; o.Space = true })
    color.Color().Red("Color-2").Print(func(o *color.Options) { o.Line = false; o.Space = true })
    color.Color().Red("Color-3").Print()

    // Output: Color-1 Color-2 Color-3

    // --> Store the ANSI-modified string to a variable and then format, write the value to stdout
    v := color.Color().Italic(color.Color().Magenta("Magenta")).String()

    fmt.Printf("Example Magenta Color Output: %s\n", v)
    // Output: Example Magenta Color Output: Magenta
}
```

- Please refer to the [code examples](./example_test.go) for additional usage and implementation details.
- See https://pkg.go.dev/github.com/x-ethr/color for additional documentation.

## Contributions

See the [**Contributing Guide**](./CONTRIBUTING.md) for additional details on getting started.
