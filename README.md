
# UrfaveCliSan

instrument actions for testing purpose.

### Documentation

See [Documentation](https://pkg.go.dev/github.com/Myriad-Dreamin/urfave-cli-san@v1.22.12)

### Primary Usage

use `Inject` before app run or `InjectAndRun` for one shot. The example is also shown as [example-InjectAndRun](https://pkg.go.dev/github.com/Myriad-Dreamin/urfave-cli-san#example-InjectAndRun).

```go
package main

var app = &cli.App{
    Commands: []cli.Command{
        {
            Name: "exec",
            Action: func(ctx *cli.Context) error {
                fmt.Printf("  Execute(%v)\n", ctx.Command.Name)
                return nil
            },
        },
    },
}

func main() {
	err := InjectAndRun(app, []string{"awesome-app", "exec"}, func(ctx *cli.Context, next ActionFunc) error {
		fmt.Printf("clisan.BeforeHook(%v)\n", ctx.Command.Name)
		err := next(ctx)
		fmt.Printf("clisan.AfterHook(%v)\n", ctx.Command.Name)
		return err
	})
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}
}
```

The expected output:

```
clisan.BeforeHook(exec)
  Execute(exec)
clisan.AfterHook(exec)
```

### Extra Options

To handle before, after functions, you have two option:
+ use `clisan.WithBeforeInstrumentation` or `clisan.WithAfterInstrumentation`
+ use `clisan.WithBeforeAfterTagging([]{"$before", "$after"})`

For example, to use `WithBeforeAfterTagging`, which is also shown as [example-InjectAndRun](https://pkg.go.dev/github.com/Myriad-Dreamin/urfave-cli-san#example-InjectAndRun).

```go

func main() {
	err := InjectAndRun(&cli.App{
		Commands: []cli.Command{
			{
				Name: "exec",
				Before: func(ctx *cli.Context) error {
					fmt.Printf("  Before(%v)\n", ctx.Command.Name)
					return nil
				},
				Action: func(ctx *cli.Context) error {
					fmt.Printf("  Execute(%v)\n", ctx.Command.Name)
					return nil
				},
				After: func(ctx *cli.Context) error {
					fmt.Printf("  After(%v)\n", ctx.Command.Name)
					return nil
				},
			},
		},
	}, []string{"awesome-app", "exec"}, func(ctx *cli.Context, next ActionFunc) error {
		fmt.Printf("clisan.BeforeHook%s(%v)\n", GetTaintPosition(ctx.App), ctx.Command.Name)
		err := next(ctx)
		fmt.Printf("clisan.AfterHook%s(%v)\n", GetTaintPosition(ctx.App), ctx.Command.Name)
		return err
	}, WithBeforeAfterTagging([]string{BeforeTag, AfterTag}))
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}
}
```

The expected output:

```
clisan.BeforeHook$before(exec)
  Before(exec)
clisan.AfterHook$before(exec)
clisan.BeforeHook$current(exec)
  Execute(exec)
clisan.AfterHook$current(exec)
clisan.BeforeHook$after(exec)
  After(exec)
clisan.AfterHook$after(exec)
```

Note, it is meaningless to use `WithBeforeInstrumentation/WithAfterInstrumentation` and `WithBeforeAfterTagging` at the same time.
