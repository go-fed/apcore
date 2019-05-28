package apcore

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/logger"
)

var (
	// Flags for apcore
	debugFlag        = flag.Bool("debug", false, "Enable the development server on localhost & other developer quality of life features")
	systemLogFlag    = flag.Bool("syslog", false, "Also enable logging to system")
	infoLogFileFlag  = flag.String("info_log_file", "", "Log file for info, defaults to os.Stdout")
	errorLogFileFlag = flag.String("error_log_file", "", "Log file for errors, defaults to os.Stderr")
	configFlag       = flag.String("config", "config.ini", "Path to the configuration file")
)

var (
	// These loggers will only respect the logging flags while the call to
	// Run is executing. Otherwise, they log to os.Stdout and os.Stderr.
	InfoLogger  *logger.Logger = logger.Init("apcore", false, false, os.Stdout)
	ErrorLogger *logger.Logger = logger.Init("apcore", false, false, os.Stderr)
)

// Usage is overridable so client applications can add custom additional
// information on the command line.
var Usage func() = func() {}

func init() {
	flag.Usage = func() {
		Usage()
		fmt.Fprintf(flag.CommandLine.Output(), "Actions are:\n%s\n", allActionsUsage())
		flag.PrintDefaults()
	}
}

type cmdAction struct {
	Name        string
	Description string
	Action      func(Application) error
}

// String formats the command line action similarly to the standard library
// flag package.
func (c cmdAction) String() string {
	return fmt.Sprintf("  %s\n    \t%s",
		c.Name,
		strings.ReplaceAll(c.Description, "\n", "\n    \t"))
}

var (
	// apcore actions supported
	serve cmdAction = cmdAction{
		Name:        "serve",
		Description: "Launch the application server.",
		Action:      serveFn,
	}
	guideNew cmdAction = cmdAction{
		Name:        "new",
		Description: "Launch the guided application setup process guided by Clarke the Cow.",
		Action:      guideNewFn,
	}
	initDb cmdAction = cmdAction{
		Name:        "init-db",
		Description: "Initializes a new, empty database with the required tables if no existing database tables are detected. Requires a configuration.",
		Action:      initDbFn,
	}
	initAdmin cmdAction = cmdAction{
		Name:        "init-admin",
		Description: "Initializes a new administrator user account. Requires a database.",
		Action:      initAdminFn,
	}
	configure cmdAction = cmdAction{
		Name:        "configure",
		Description: "Create or overwrite the server configuration in a guided flow.",
		Action:      configureFn,
	}
	version cmdAction = cmdAction{
		Name:        "version",
		Description: "List the current software and version.",
		Action:      versionFn,
	}
	allActions []cmdAction
)

func init() {
	allActions = []cmdAction{
		serve,
		guideNew,
		initDb,
		initAdmin,
		configure,
		version,
	}
}

func allActionsUsage() string {
	var b bytes.Buffer
	for _, v := range allActions {
		b.WriteString(v.String())
		b.WriteString("\n")
	}
	return b.String()
}

// The 'serve' command line action.
func serveFn(a Application) error {
	_, err := newServer(*configFlag, a, *debugFlag)
	if err != nil {
		return err
	}
	// TODO
	return nil
}

// The 'new' command line action.
func guideNewFn(a Application) error {
	sw := a.Software()
	fmt.Println(clarkeSays(fmt.Sprintf(`
Hi, I'm Clarke the Cow! I am here to help you set up your ActivityPub
software. It is called %q. This is version %d.%d.%d, but I don't know what
that means. I'm a cow! First off, let's create a configuration file. Let's get
mooving!`,
		sw.Name,
		sw.MajorVersion,
		sw.MinorVersion,
		sw.PatchVersion)))
	err := configureFn(a)
	if err != nil {
		return err
	}
	fmt.Println(clarkeSays(`
Configuration wizardry complete! It is a good idea to check that configuration
file for additional options before serving traffic. You can always re-run the
wizard using the "configure" action. Now let's initialize the database!`))
	err = initDbFn(a)
	if err != nil {
		return err
	}
	fmt.Println(clarkeSays(`
Whew! That can manually be done using the "init-db" action in the future. Next,
let's initialize your first administrator account in the database.`))
	err = initAdminFn(a)
	if err != nil {
		return err
	}
	fmt.Println(clarkeSays(`
Moo~! That was the "init-admin" action. We are done, but before you run the
"serve" action, please do double check your configuration file! Bye bye!`))
	return nil
}

// The 'init-db' command line action.
func initDbFn(a Application) error {
	// TODO
	return nil
}

// The 'init-admin' command line action.
func initAdminFn(a Application) error {
	// TODO
	return nil
}

// The 'configure' command line action.
func configureFn(a Application) error {
	if len(*configFlag) == 0 {
		return fmt.Errorf("config flag not set")
	}
	exists := false
	if _, err := os.Stat(*configFlag); err == nil {
		exists = true
		cont, err := promptFileExistsContinue(*configFlag)
		if err != nil {
			return err
		}
		if !cont {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot modify configuration: %s", err)
	}
	cfg, err := promptNewConfig(*configFlag)
	if err != nil {
		return err
	}
	InfoLogger.Info("Calling application to get default config options")
	acfg := a.NewConfiguration()
	if exists {
		cont, err := promptOverwriteExistingFile(*configFlag)
		if err != nil {
			return err
		}
		if !cont {
			InfoLogger.Infof("Did not overwrite configuration file at %s", *configFlag)
			return nil
		}
	}
	err = saveConfigFile(*configFlag, cfg, acfg)
	if err != nil {
		return err
	}
	InfoLogger.Infof("Successfully wrote configuration file to %s", *configFlag)
	return nil
}

// The 'version' command line action.
func versionFn(a Application) error {
	fmt.Fprintf(os.Stdout, "%s; %s\n", a.Software(), apCoreSoftware())
	return nil
}

// Run will launch the apcore server.
func Run(a Application) {
	if !flag.Parsed() {
		flag.Parse()
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Check and prepare debug mode
	if *debugFlag {
		InfoLogger.Info("Debug mode enabled")
		if len(*infoLogFileFlag) > 0 {
			InfoLogger.Warning("info_log_file flag ignored in debug mode")
		}
		if len(*errorLogFileFlag) > 0 {
			InfoLogger.Warning("error_log_file flag ignored in debug mode")
		}
	} else {
		// Prepare production logging
		var il, el io.Writer = os.Stdout, os.Stderr
		var err error
		if len(*infoLogFileFlag) > 0 {
			il, err = os.OpenFile(
				*infoLogFileFlag,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0660)
			if err != nil {
				ErrorLogger.Errorf("cannot open %s: %s", *infoLogFileFlag, err)
				os.Exit(1)
			}
		}
		if len(*errorLogFileFlag) > 0 {
			el, err = os.OpenFile(
				*errorLogFileFlag,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0660)
			if err != nil {
				ErrorLogger.Errorf("cannot open %s: %s", *infoLogFileFlag, err)
				os.Exit(1)
			}
		}
		InfoLogger = logger.Init("apcore", false, *systemLogFlag, il)
		defer func() {
			InfoLogger.Close()
			InfoLogger = logger.Init("apcore", false, false, os.Stdout)
		}()
		ErrorLogger = logger.Init("apcore", false, *systemLogFlag, el)
		defer func() {
			ErrorLogger.Close()
			ErrorLogger = logger.Init("apcore", false, false, os.Stderr)
		}()
	}

	// Conduct the action
	var action cmdAction
	found := false
	for _, v := range allActions {
		if v.Name == flag.Arg(0) {
			action = v
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", flag.Arg(0))
		fmt.Fprintf(os.Stderr, "Available actions:\n%s", allActionsUsage())
		os.Exit(1)
	} else if err := action.Action(a); err != nil {
		ErrorLogger.Errorf("error running %s: %s", flag.Arg(0), err)
		os.Exit(1)
	}
}
