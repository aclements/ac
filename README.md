ac is a basic architectural calculator.

To install, run `go install github.com/aclements/ac@latest`.

The ac command supports the usual arithmetic operations over numbers
and dimensions of the form

    <n>' <n>"

Any number can also include a fractional part separated by a space,
such as "1 1/2". The fractional part must not include any spaces.

Examples:

    9' - 20"
    = 7' 4"

    8' 1 1/2" / 2
    = 4' 3/4"

ac uses arbitrary-precision rational numbers for all arithmetic.