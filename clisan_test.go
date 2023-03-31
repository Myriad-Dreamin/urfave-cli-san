package clisan

import (
	"fmt"

	"github.com/urfave/cli"
)

// ExampleInjectAndRun is an example of how to use InjectAndRun
// to inject a BeforeHook and AfterHook into a cli.App
// and run the app.
// The output of this example is:
// + clisan.BeforeHook(exec)
// +   Execute(exec)
// + clisan.AfterHook(exec)
func ExampleInjectAndRun() {
	err := InjectAndRun(&cli.App{
		Commands: []cli.Command{
			{
				Name: "exec",
				Action: func(ctx *cli.Context) error {
					fmt.Printf("  Execute(%v)\n", ctx.Command.Name)
					return nil
				},
			},
		},
	}, []string{"awesome-app", "exec"}, func(ctx *cli.Context, next ActionFunc) error {
		fmt.Printf("clisan.BeforeHook(%v)\n", ctx.Command.Name)
		err := next(ctx)
		fmt.Printf("clisan.AfterHook(%v)\n", ctx.Command.Name)
		return err
	})
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}
}

// ExampleWithBeforeAfterTagging is an example of how to use WithBeforeAfterTagging
// to inject a BeforeHook and AfterHook into a cli.App
// and run the app.
// The output of this example is:
// + clisan.BeforeHook$before(exec)
// +   Before(exec)
// + clisan.AfterHook$before(exec)
// + clisan.BeforeHook$current(exec)
// +   Execute(exec)
// + clisan.AfterHook$current(exec)
// + clisan.BeforeHook$after(exec)
// +   After(exec)
// + clisan.AfterHook$after(exec)
func ExampleWithBeforeAfterTagging() {
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
