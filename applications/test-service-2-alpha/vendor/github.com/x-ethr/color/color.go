package color

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	"golang.org/x/term"

	"github.com/x-ethr/color/internal"
)

// ANSI represents a string that contains ANSI escape codes for terminal colors.
//
// ANSI provides methods for applying different colors to the input string(s) and returning it as an ANSI string.
// It uses functions from the ANSI package to convert the input to the desired color.
// If the current operating system is not Windows and the CI variable is false, it adds the color code before and the reset code after each input string.
type ANSI string

// Black applies the black color to the input string(s) and returns it as an ANSI string.
// It uses the Black() function from the ANSI package to convert the input to black color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Black(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, black(input...)))

	*c = ANSI(v)

	return c
}

// red applies the red color to the input string(s) and returns it as an ANSI string.
// It uses the red() function from the ANSI package to convert the input to red color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Red(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, red(input...)))

	*c = ANSI(v)

	return c
}

// green applies the green color to the input string(s) and returns it as an ANSI string.
// It uses the green() function from the ANSI package to convert the input to green color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Green(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, green(input...)))

	*c = ANSI(v)

	return c
}

// yellow applies the yellow color to the input string(s) and returns it as an ANSI string.
// It uses the yellow() function from the ANSI package to convert the input to yellow color.
// If the current operating system is not Windows and the CI variable is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Yellow(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, yellow(input...)))

	*c = ANSI(v)

	return c
}

// blue applies the blue color to the input string(s) and returns it as an ANSI string.
// It uses the blue() function from the ANSI package to convert the input to blue color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Blue(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, blue(input...)))

	*c = ANSI(v)

	return c
}

// magenta applies the purple color to the input string(s) and returns it as an ANSI string.
// It uses the magenta() function from the ANSI package to convert the input to purple color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Magenta(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, magenta(input...)))

	*c = ANSI(v)

	return c
}

// cyan applies the cyan color to the input string(s) and returns it as an ANSI string.
// It uses the cyan() function from the ANSI package to convert the input to cyan color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Cyan(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, cyan(input...)))

	*c = ANSI(v)

	return c
}

// gray applies the gray color to the input string(s) and returns it as an ANSI string.
// It uses the gray() function from the ANSI package to convert the input to gray color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Gray(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, gray(input...)))

	*c = ANSI(v)

	return c
}

// white applies the white color to the input string(s) and returns it as an ANSI string.
// It uses the white() function from the ANSI package to convert the input to white color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) White(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, white(input...)))

	*c = ANSI(v)

	return c
}

// Default applies the default color to the input string(s) and returns it as an ANSI string.
// It uses the Default() function from the ANSI package to convert the input to default color.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
func (c *ANSI) Default(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, defaults(input...)))

	*c = ANSI(v)

	return c
}

// bold applies the bold style to the input string(s) and returns it as an ANSI string.
// It uses the bold() function from the ANSI package to convert the input to bold style.
// If the current operating system is not Windows and CI is false, it adds the style code before and the reset code after each input string.
func (c *ANSI) Bold(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, bold(input...)))

	*c = ANSI(v)

	return c
}

// dim applies a dimmed style to the input string(s) and returns it as an ANSI string.
// If the current operating system is not Windows and CI is false, it adds the style code before and the reset code after each input string.
func (c *ANSI) Dim(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, dim(input...)))

	*c = ANSI(v)

	return c
}

// italic applies the italic style to the input string(s) and returns it as an ANSI string.
// It uses the italic() function from the ANSI package to convert the input to italic style.
// If the current operating system is not Windows and CI is false, it adds the style code before and the reset code after each input string.
func (c *ANSI) Italic(input ...any) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, italic(input...)))

	*c = ANSI(v)

	return c
}

// Print outputs the ANSI string to the standard output based on the specified configuration settings.
//
// By default, it does not add any extra characters at the end.
//
//   - If the Line option is enabled, it adds a newline character after the ANSI string.
//   - If the Space option is enabled, it adds a space character after the ANSI string.
//   - It uses the os.Stdout and fmt.Fprintf methods to write the ANSI string to the standard output with the specified format.
//
// For customizing the [io.Writer], instead of using [os.Stdout], see [ANSI.Write].
func (c *ANSI) Print(settings ...Variadic) {
	var v string

	o := options()
	for _, configuration := range settings {
		configuration(o)
	}

	if c != nil {
		v = string(*c)
	}

	switch {
	case o.Line: // --> add newline
		fmt.Fprintf(os.Stdout, "%s\n", v)
	case o.Space:
		fmt.Fprintf(os.Stdout, "%s ", v)
	default:
		fmt.Fprintf(os.Stdout, "%s", v)
	}
}

// Write writes the ANSI string to the provided io.Writer.
// If c is not nil, it converts c to a string and trims any leading/trailing spaces.
// It then formats the string with a newline character and writes it to the io.Writer using fmt.Fprintf.
func (c *ANSI) Write(w io.Writer, settings ...Variadic) {
	var v string

	o := options()
	for _, configuration := range settings {
		configuration(o)
	}

	if c != nil {
		v = string(*c)
	}

	switch {
	case o.Line: // --> add newline
		fmt.Fprintf(w, "%s\n", v)
	case o.Space:
		fmt.Fprintf(w, "%s ", v)
	default:
		fmt.Fprintf(w, "%s", v)
	}
}

// Overload allows passing ANSI Escape characters directly.
// It constructs an ANSI string with the provided ANSI Escape characters and the input string.
func (c *ANSI) Overload(ansi []byte, input string) *ANSI {
	v := strings.TrimSpace(fmt.Sprintf("%s %s", *c, Overload(ansi, input)))

	*c = ANSI(v)

	return c
}

// String returns the ANSI string as a raw string.
func (c *ANSI) String() string {
	if c == nil {
		return ""
	}

	return string(*c)
}

// Color initializes and returns a new ANSI string.
func Color() *ANSI {
	return new(ANSI)
}

// typecast converts the provided entity to a string. Currently only supports simple and [ANSI] types.
func typecast(entity interface{}) string {
	var partial string

	switch entity.(type) {
	case string:
		partial = fmt.Sprintf("%s", entity.(string))
	case ANSI, *ANSI:
		partial = fmt.Sprintf("%s", entity.(*ANSI).String())
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, complex64, complex128:
		partial = fmt.Sprintf("%d", entity.(int))
	case bool:
		partial = fmt.Sprintf("%t", entity.(bool))
	}

	return partial
}

// Overload constructs an ANSI string with the provided ANSI Escape characters and the input string.
func Overload(ansi []byte, input string) string {
	output := make([]string, 0)

	var partials []string

	color := ansi

	if Available() {
		partials = []string{string(color), input, internal.Reset}
	} else {
		partials = []string{input}
	}

	output = append(output, strings.Join(partials, ""))

	return strings.Join(output, " ")
}

// defaults applies default color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func defaults(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Reset

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, color}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// black applies the black color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func black(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Black

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// red applies the red color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func red(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Red

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// green applies the green color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func green(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Green

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// yellow applies the yellow color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func yellow(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Yellow

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// blue applies the blue color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func blue(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Blue

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// magenta applies the purple color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func magenta(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Magenta

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// cyan applies the cyan color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func cyan(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Cyan

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// gray applies the gray color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func gray(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Gray

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// white applies the white color to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the color code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func white(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.White

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// bold applies the bold style to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the style code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func bold(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Bold

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// italic applies the italic style to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the style code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func italic(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Italic

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// dim applies a dimmed style to the input string(s) and returns it as a raw string.
// If the current operating system is not Windows and CI is false, it adds the style code before and the reset code after each input string.
//
//   - Currently only supports [ANSI] and simple types.
func dim(input ...any) string {
	output := make([]string, 0)

	var partials []string

	for index := range input {
		color := internal.Dim

		var partial = typecast(input[index])

		if Available() {
			partials = []string{color, partial, internal.Reset}
		} else {
			partials = []string{partial}
		}

		output = append(output, strings.Join(partials, ""))
	}

	return strings.Join(output, " ")
}

// Available checks if the terminal has a TTY (teletypewriter) available and returns a boolean value indicating
// if the system's output buffer is capable of color output.
func Available() bool {
	return !(CI()) && runtime.GOOS != "windows"
}

// force is a variable of type atomic.Value. It is used to store a boolean value that determines whether to force color output.
// It is initialized with a default value of false in the init() function and can be changed using the Force() function.
var force atomic.Value

// Force will force the runtime to color its output.
func Force() {
	force.Store(true)
}

// Forcing provides a sanity-check of the current, global [atomic.Value] assignment.
func Forcing() bool {
	if v := force.Load(); v != nil && v.(bool) {
		return true
	}

	return false
}

// Unset provides the ability to globally, explicitly set the force [atomic.Value] value to false.
func Unset() {
	force.Store(false)
}

// CI checks if the terminal supports color output by checking the value of the "CI" environment variable.
//   - If the "CI" environment variable is set to "true", "yes", "on", or "1", CI returns true.
//   - If the "CI" environment variable is set to "false", "no", "off", or "0", CI returns false.
//   - If the "CI" environment variable is not set and terminal is not a TTY, CI returns true.
//
// Default return value is true.
func CI() bool {
	// --> force color => CI = false
	if f := force.Load(); f != nil && ((f).(bool)) == true {
		return false
	}

	switch v := strings.ToLower(os.Getenv("CI")); v {
	case "true", "yes", "on", "1":
		return true
	case "false", "no", "off", "0":
		return false
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	return true
}

func init() {
	force.Store(false)
}
