# seams

Go static analyzer that reports untestable function/method calls.

## As a command

First, install `seams` command via `go get`.

```shellsession
$ go get github.com/ichiban/seam/cmd/seams
```

Then, run `seams` command at your package directory.

```shellsession
$ seams ./...
/Users/ichiban/src/swuntestable/main.go:8:17: untestable function/method call: time.Parse
/Users/ichiban/src/swuntestable/main.go:11:7: untestable function/method call: (time.Duration).Hours
/Users/ichiban/src/swuntestable/main.go:11:7: untestable function/method call: (time.Time).Sub
/Users/ichiban/src/swuntestable/main.go:11:16: untestable function/method call: time.Now
/Users/ichiban/src/swuntestable/main.go:12:2: untestable function/method call: fmt.Printf
```

## As an `analysis.Analyzer`

First, install the package via `go get`.

```shellsession
$ go get github.com/ichiban/seams
```

Then, include `cyclomatic.Analyzer` in your checker.

```go
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"

	"github.com/ichiban/seams"
)

func main() {
	multichecker.Main(
		// other analyzers of your choice
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,

		seams.Analyzer,
	)
}
```

## What do you mean by untestable?

This static analyzer reports function/method calls that are of:

- functions/methods defined in other packages,
- not builtin functions,
- not in tests, nor
- in generated files

because those function/method calls can't be replaced by test doubles in tests.

For example, `time.Now()` isn't testable because we can't change its behaviour for our tests.
Though, `timeNow()` where `var timeNow = time.Now` is testable because we can change the behaviour in tests.

### untestable example

```go
package main

import (
        "fmt"
        "time"
)

var date = must(time.Parse(time.RFC3339, "2019-12-20T00:00:00+09:00"))

func main() {
        d := date.Sub(time.Now()).Hours() / 24
        fmt.Printf("%d days until Star Wars: The Rise of Skywalker\n", int(d))
}

func must(t time.Time, err error) time.Time {
        if err != nil {
                panic(err)
        }
        return t
}
```

### testable example

```go
package main

import (
    "fmt"
    "time"
)

var (
    timeParse         = time.Parse
    timeDurationHours = time.Duration.Hours
    timeTimeSub       = time.Time.Sub
    timeNow           = time.Now
    fmtPrintf         = fmt.Printf
)

var date = must(timeParse(time.RFC3339, "2019-12-20T00:00:00+09:00"))

func main() {
    d := timeDurationHours(timeTimeSub(date, timeNow())) / 24
    fmtPrintf("%d days until Star Wars: The Rise of Skywalker\n", int(d))
}

func must(t time.Time, err error) time.Time {
    if err != nil {
        panic(err)
    }
    return t
}
```

```go
package main

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
    t.Run("on the day", func(t *testing.T) {
        assert := assert.New(t)

        now, err := time.Parse(time.RFC3339, "2019-12-20T12:00:00+09:00")
        assert.NoError(err)

        n := timeNow
        defer func() { timeNow = n }()
        timeNow = func() time.Time {
            return now
        }

        called := false
        p := fmtPrintf
        defer func() { fmtPrintf = p }()
        fmtPrintf = func(format string, a ...interface{}) (int, error) {
            assert.Equal("%d days until Star Wars: The Rise of Skywalker\n", format)
            assert.Equal([]interface{}{0}, a)

            called = true
            return 0, nil
        }

        main()

        assert.True(called)
    })
}
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
