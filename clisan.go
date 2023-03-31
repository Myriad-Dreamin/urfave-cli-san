package clisan

import (
	"fmt"
	"io"

	"github.com/urfave/cli"
)

const (
	AfterTag         = "$after"
	BeforeTag        = "$before"
	CurrentTag       = "$current"
	TaintPositionKey = "$taint"
)

type InstrumentOptions func(opts *instrumentState)
type InstrumentFunc = func(ctx *cli.Context, next ActionFunc) error

// Inject instrument the app.
// The returned function restores the original state of the app.
func Inject(app *cli.App, instrumentation InstrumentFunc, opts ...InstrumentOptions) (restore func()) {
	var san = &transformer{
		instrumentingApp: app,
		instrumentState: instrumentState{
			proxy:     instrumentation,
			userProxy: instrumentation,
		},
		overridenBefore: make(map[*cli.BeforeFunc]BeforeFunc),
		overriden:       make(map[*interface{}]interface{}),
		overridenAfter:  make(map[*cli.AfterFunc]AfterFunc),
	}

	for _, optFn := range opts {
		optFn(&san.instrumentState)
	}

	san.instrumentHandlersAndApp(app)

	return san.restore
}

type (
	InstrumentBeforeFunc          = func(ctx *cli.Context, next BeforeFunc) error
	InstrumentAfterFunc           = func(ctx *cli.Context, next AfterFunc) error
	InstrumentHelpPrinter         = func(w io.Writer, templ string, data interface{}, next HelpPrinter)
	InstrumentHelpPrinterCustom   = func(w io.Writer, templ string, data interface{}, customFunc map[string]interface{}, next HelpPrinterCustom)
	InstrumentCommandNotFoundFunc = func(ctx *cli.Context, command string, next CommandNotFoundFunc)
	InstrumentOnUsageErrorFunc    = func(ctx *cli.Context, err error, isSubcommand bool, next OnUsageErrorFunc) error
	InstrumentExitErrHandlerFunc  = func(ctx *cli.Context, err error, next ExitErrHandlerFunc)
	InstrumentFlagStringFunc      = func(flag cli.Flag, next FlagStringFunc) string
	InstrumentFlagNamePrefixFunc  = func(fullName, placeholder string, next FlagNamePrefixFunc) string
	InstrumentFlagEnvHintFunc     = func(envVar, str string, next FlagEnvHintFunc) string
	InstrumentFlagFileHintFunc    = func(filePath, str string, next FlagFileHintFunc) string
)

func WithHelpInstrumentation(proxy InstrumentHelpPrinter) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.helpPrinter = proxy
	}
}

func WithHelpCustomInstrumentation(proxy InstrumentHelpPrinterCustom) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.helpPrinterCustom = proxy
	}
}

func WithCommandNotFoundInstrumentation(proxy InstrumentCommandNotFoundFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.commandNotFound = proxy
	}
}

func WithOnUsageErrorInstrumentation(proxy InstrumentOnUsageErrorFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.onUsageError = proxy
	}
}

func WithExitErrHandlerInstrumentation(proxy InstrumentExitErrHandlerFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.exitErrHandler = proxy
	}
}

func WithFlagStringInstrumentation(proxy InstrumentFlagStringFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.flagString = proxy
	}
}

func WithFlagNamePrefixInstrumentation(proxy InstrumentFlagNamePrefixFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.flagNamePrefix = proxy
	}
}

func WithFlagEnvHintInstrumentation(proxy InstrumentFlagEnvHintFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.flagEnvHint = proxy
	}
}

func WithFlagFileHintInstrumentation(proxy InstrumentFlagFileHintFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.flagFileHint = proxy
	}
}

func WithBeforeInstrumentation(proxyBefore InstrumentBeforeFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.proxyBefore = proxyBefore
	}
}

func WithAfterInstrumentation(proxyAfter InstrumentAfterFunc) InstrumentOptions {
	return func(opts *instrumentState) {
		opts.proxyAfter = proxyAfter
	}
}

func WithBeforeAfterTagging(tags []string) InstrumentOptions {
	for _, s := range tags {
		switch s {
		case BeforeTag, AfterTag:
		default:
			panic(fmt.Errorf("invalid tag: %s, must be one of %q, %q", s, BeforeTag, AfterTag))
		}
	}
	return func(opts *instrumentState) {
		for _, s := range tags {
			switch s {
			case BeforeTag:
				opts.proxyBefore = func(ctx *cli.Context, next BeforeFunc) error {
					ctx.App.Metadata[TaintPositionKey] = BeforeTag
					return opts.userProxy(ctx, next)
				}
			case AfterTag:
				opts.proxyAfter = func(ctx *cli.Context, next AfterFunc) error {
					ctx.App.Metadata[TaintPositionKey] = AfterTag
					return opts.userProxy(ctx, next)
				}
			default:
				panic("invalid tag")
			}
		}
		opts.proxy = func(ctx *cli.Context, next BeforeFunc) error {
			ctx.App.Metadata[TaintPositionKey] = CurrentTag
			return opts.userProxy(ctx, next)
		}
	}
}

// GetTaintPosition returns the position tag of the current execution.
// The tag is set by WithBeforeAfterTagging.
func GetTaintPosition(app *cli.App) string {
	return app.Metadata[TaintPositionKey].(string)
}

// InjectAndRun instrument the app and runs it.
// The returned error is the same as the one returned by app.Run, i.e. https://pkg.go.dev/github.com/urfave/cli/v3#App.Run.
func InjectAndRun(app *cli.App, arguments []string, instrumentation InstrumentFunc, opts ...InstrumentOptions) error {
	restore := Inject(app, instrumentation, opts...)
	defer restore()
	return app.Run(arguments)
}

type instrumentState struct {
	userProxy   InstrumentFunc
	proxy       InstrumentFunc
	proxyBefore InstrumentBeforeFunc
	proxyAfter  InstrumentAfterFunc

	helpPrinter       InstrumentHelpPrinter
	helpPrinterCustom InstrumentHelpPrinterCustom

	commandNotFound InstrumentCommandNotFoundFunc
	onUsageError    InstrumentOnUsageErrorFunc
	exitErrHandler  InstrumentExitErrHandlerFunc

	flagString     InstrumentFlagStringFunc
	flagNamePrefix InstrumentFlagNamePrefixFunc
	flagEnvHint    InstrumentFlagEnvHintFunc
	flagFileHint   InstrumentFlagFileHintFunc
}

type overridenState struct {
	helpPrinter       HelpPrinter
	helpPrinterCustom HelpPrinterCustom

	commandNotFound CommandNotFoundFunc
	onUsageError    OnUsageErrorFunc
	exitErrHandler  ExitErrHandlerFunc

	flagString     FlagStringFunc
	flagNamePrefix FlagNamePrefixFunc
	flagEnvHint    FlagEnvHintFunc
	flagFileHint   FlagFileHintFunc
}

type transformer struct {
	instrumentingApp *cli.App
	instrumentState
	overridenState overridenState

	overridenBefore map[*cli.BeforeFunc]BeforeFunc
	overriden       map[*interface{}]interface{}
	overridenAfter  map[*cli.AfterFunc]AfterFunc
}

func (s *transformer) instrumentBefore(action *cli.BeforeFunc, proxy InstrumentFunc) {
	if proxy == nil || *action == nil {
		return
	}
	if _, ok := s.overridenBefore[action]; ok {
		return
	}

	var instrumenting ActionFunc = *action
	var instrumented = ActionFunc(func(ctx *cli.Context) error {
		return proxy(ctx, instrumenting)
	})
	s.overridenBefore[action] = instrumented
	*action = instrumented
}

func (s *transformer) instrumentAction(action *interface{}, proxy InstrumentFunc) {
	if *action == nil {
		return
	}
	if _, ok := s.overriden[action]; ok {
		return
	}

	var instrumenting ActionFunc
	switch a := (*action).(type) {
	case ActionFunc:
		instrumenting = a
	case cli.ActionFunc:
		instrumenting = a
	default:
		panic(fmt.Errorf("invalid action type: %T", *action))
	}

	var instrumented = ActionFunc(func(ctx *cli.Context) error {
		return proxy(ctx, instrumenting)
	})
	s.overriden[action] = instrumented
	*action = instrumented
}

func (s *transformer) instrumentAfter(action *cli.AfterFunc, proxy InstrumentFunc) {
	if proxy == nil || *action == nil {
		return
	}
	if _, ok := s.overridenAfter[action]; ok {
		return
	}

	var instrumenting ActionFunc = *action
	var instrumented = ActionFunc(func(ctx *cli.Context) error {
		return proxy(ctx, instrumenting)
	})
	s.overridenAfter[action] = instrumented
	*action = instrumented
}

func (s *transformer) instrumentHandlersAndApp(app *cli.App) {
	if s.instrumentState.helpPrinter != nil {
		s.overridenState.helpPrinter = (HelpPrinter)(cli.HelpPrinter)
		cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
			s.instrumentState.helpPrinter(w, templ, data, s.overridenState.helpPrinter)
		}
	}
	if s.instrumentState.helpPrinterCustom != nil {
		s.overridenState.helpPrinterCustom = (HelpPrinterCustom)(cli.HelpPrinterCustom)
		cli.HelpPrinterCustom = func(w io.Writer, templ string, data interface{}, customFuncs map[string]interface{}) {
			s.instrumentState.helpPrinterCustom(w, templ, data, customFuncs, s.overridenState.helpPrinterCustom)
		}
	}

	if s.instrumentState.flagString != nil {
		s.overridenState.flagString = (FlagStringFunc)(cli.FlagStringer)
		cli.FlagStringer = func(flag cli.Flag) string {
			return s.instrumentState.flagString(flag, s.overridenState.flagString)
		}
	}
	if s.instrumentState.flagNamePrefix != nil {
		s.overridenState.flagNamePrefix = (FlagNamePrefixFunc)(cli.FlagNamePrefixer)
		cli.FlagNamePrefixer = func(fullName, placeholder string) string {
			return s.instrumentState.flagNamePrefix(fullName, placeholder, s.overridenState.flagNamePrefix)
		}
	}
	if s.instrumentState.flagEnvHint != nil {
		s.overridenState.flagEnvHint = (FlagEnvHintFunc)(cli.FlagEnvHinter)
		cli.FlagEnvHinter = func(envVar, str string) string {
			return s.instrumentState.flagEnvHint(envVar, str, s.overridenState.flagEnvHint)
		}
	}
	if s.instrumentState.flagFileHint != nil {
		s.overridenState.flagFileHint = (FlagFileHintFunc)(cli.FlagFileHinter)
		cli.FlagFileHinter = func(file, str string) string {
			return s.instrumentState.flagFileHint(file, str, s.overridenState.flagFileHint)
		}
	}

	if s.instrumentState.commandNotFound != nil {
		s.overridenState.commandNotFound = (CommandNotFoundFunc)(app.CommandNotFound)
		app.CommandNotFound = func(c *cli.Context, command string) {
			s.instrumentState.commandNotFound(c, command, s.overridenState.commandNotFound)
		}
	}
	if s.instrumentState.onUsageError != nil {
		s.overridenState.onUsageError = (OnUsageErrorFunc)(app.OnUsageError)
		app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
			return s.instrumentState.onUsageError(c, err, isSubcommand, s.overridenState.onUsageError)
		}
	}
	if s.instrumentState.exitErrHandler != nil {
		s.overridenState.exitErrHandler = (ExitErrHandlerFunc)(app.ExitErrHandler)
		app.ExitErrHandler = func(c *cli.Context, err error) {
			s.instrumentState.exitErrHandler(c, err, s.overridenState.exitErrHandler)
		}
	}

	s.instrumentBefore(&app.Before, s.proxyBefore)
	s.instrumentAction(&app.Action, s.proxy)
	s.instrumentAfter(&app.After, s.proxyAfter)

	s.instrument(app.Commands)
}

func (s *transformer) instrument(commands []cli.Command) {
	for i := range commands {
		cmd := &commands[i]
		s.instrument(cmd.Subcommands)
		s.instrumentBefore(&cmd.Before, s.proxyBefore)
		s.instrumentAction(&cmd.Action, s.proxy)
		s.instrumentAfter(&cmd.After, s.proxyAfter)
	}
}

func (s *transformer) restore() {
	for k, v := range s.overriden {
		*k = v
	}
	for k := range s.overriden {
		delete(s.overriden, k)
	}
	for k, v := range s.overridenBefore {
		*k = v
	}
	for k := range s.overridenBefore {
		delete(s.overridenBefore, k)
	}
	for k, v := range s.overridenAfter {
		*k = v
	}
	for k := range s.overridenAfter {
		delete(s.overridenAfter, k)
	}

	if s.overridenState.helpPrinter != nil {
		cli.HelpPrinter = s.overridenState.helpPrinter
	}
	if s.overridenState.helpPrinterCustom != nil {
		cli.HelpPrinterCustom = s.overridenState.helpPrinterCustom
	}
	if s.overridenState.flagString != nil {
		cli.FlagStringer = s.overridenState.flagString
	}
	if s.overridenState.flagNamePrefix != nil {
		cli.FlagNamePrefixer = s.overridenState.flagNamePrefix
	}
	if s.overridenState.flagEnvHint != nil {
		cli.FlagEnvHinter = s.overridenState.flagEnvHint
	}
	if s.overridenState.flagFileHint != nil {
		cli.FlagFileHinter = s.overridenState.flagFileHint
	}

	if s.overridenState.commandNotFound != nil {
		s.instrumentingApp.CommandNotFound = s.overridenState.commandNotFound
	}
	if s.overridenState.onUsageError != nil {
		s.instrumentingApp.OnUsageError = s.overridenState.onUsageError
	}
	if s.overridenState.exitErrHandler != nil {
		s.instrumentingApp.ExitErrHandler = cli.ExitErrHandlerFunc(s.overridenState.exitErrHandler)
	}
}

// BeforeFunc is an action to execute before any subcommands are run, but after
// the context is ready if a non-nil error is returned, no subcommands are run
type BeforeFunc = func(*cli.Context) error

// AfterFunc is an action to execute after any subcommands are run, but after the
// subcommand has finished it is run even if Action() panics
type AfterFunc = func(*cli.Context) error

// ActionFunc is the action to execute when no subcommands are specified
type ActionFunc = func(*cli.Context) error

// Prints help for the App or Command
type HelpPrinter = func(w io.Writer, templ string, data interface{})

// Prints help for the App or Command with custom template function.
type HelpPrinterCustom = func(w io.Writer, templ string, data interface{}, customFunc map[string]interface{})

// CommandNotFoundFunc is executed if the proper command cannot be found
type CommandNotFoundFunc = func(*cli.Context, string)

// OnUsageErrorFunc is executed if an usage error occurs. This is useful for displaying
// customized usage error messages.  This function is able to replace the
// original error messages.  If this function is not set, the "Incorrect usage"
// is displayed and the execution is interrupted.
type OnUsageErrorFunc = func(context *cli.Context, err error, isSubcommand bool) error

// ExitErrHandlerFunc is executed if provided in order to handle ExitError values
// returned by Actions and Before/After functions.
type ExitErrHandlerFunc func(context *cli.Context, err error)

// FlagStringFunc is used by the help generation to display a flag, which is
// expected to be a single line.
type FlagStringFunc = func(cli.Flag) string

// FlagNamePrefixFunc is used by the default FlagStringFunc to create prefix
// text for a flag's full name.
type FlagNamePrefixFunc = func(fullName, placeholder string) string

// FlagEnvHintFunc is used by the default FlagStringFunc to annotate flag help
// with the environment variable details.
type FlagEnvHintFunc = func(envVar, str string) string

// FlagFileHintFunc is used by the default FlagStringFunc to annotate flag help
// with the file details.
type FlagFileHintFunc = func(file, str string) string
