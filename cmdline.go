// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2019 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package apcore

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/google/logger"
)

var (
	// Flags for apcore
	debugFlag        = flag.Bool("debug", false, "Enable the development server on localhost & other developer quality of life features")
	systemLogFlag    = flag.Bool("syslog", false, "Also logs to system (stdout and stderr) if logging to a file")
	infoLogFileFlag  = flag.String("info_log_file", "", "Log file for info, defaults to stdout")
	errorLogFileFlag = flag.String("error_log_file", "", "Log file for errors, defaults to stderr")
	configFlag       = flag.String("config", "config.ini", "Path to the configuration file")
)

var (
	// These loggers will only respect the logging flags while the call to
	// Run is executing. Otherwise, they log to os.Stdout and os.Stderr.
	InfoLogger  *logger.Logger = logger.Init("apcore", false, false, os.Stdout)
	ErrorLogger *logger.Logger = logger.Init("apcore", false, false, os.Stderr)
)

// Usage is overridable so client applications can add custom additional
// help information on the command line.
//
// Override this instead of the flag.Usage global variable.
var Usage func() = func() {}

// CmdLineName is overridable so client applications can match the help text
// with the correct executable name.
var CmdLineName func() string = func() string { return "example" }

func init() {
	flag.Usage = func() {
		Usage()
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage:\n\n    %s <action> [arguments]\n\n",
			CmdLineName())
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"This executable supports different actions to facilitate easier administration,\n"+
				"and flags to modify behavior at run-time. Each action will log its behavior, by\n"+
				"default stderr and stdout. However, non-debug commands can specify writing logs\n"+
				"to file for auditing and record keeping purposes.\n")
		fmt.Fprintf(
			flag.CommandLine.Output(),
			clarkeSays("Hi, I'm Clarke the Cow! When you run certain commands, I will help guide you "+
				"and ensure you have a smooooth time. Ciao!"))
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"\nThis application is built using apcore %d.%d.%d, which is licensed under the\n",
			apcoreMajorVersion,
			apcoreMinorVersion,
			apcorePatchVersion)
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"GNU Affero General Public License. Thank you for choosing to use this software.\n\n")
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Supported actions are:\n%s\n",
			allActionsUsage())
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Supported flags:\n")
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
	help cmdAction = cmdAction{
		Name:        "help",
		Description: "Print this help dialog",
		Action:      helpFn,
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
		help,
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
	s, err := newServer(*configFlag, a, *debugFlag)
	if err != nil {
		return err
	}
	interruptCh := make(chan os.Signal, 2)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interruptCh
		s.stop()
	}()
	return s.start()
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
	fmt.Println(clarkeSays(`
We're connecting to the database using the specs in the config file, creating
tables, and then closing all connections.`))
	c, err := loadConfigFile(*configFlag, a, *debugFlag)
	if err != nil {
		return err
	}
	db, err := newDatabase(c, a, *debugFlag)
	if err != nil {
		return err
	}
	err = db.OpenCreateTablesClose()
	if err != nil {
		return err
	}
	fmt.Println(clarkeSays(`Database initialization udderly complete!`))
	return nil
}

// The 'init-admin' command line action.
func initAdminFn(a Application) error {
	fmt.Println(clarkeSays(`
Moo~, let's create an administrative account!`))
	// TODO
	return nil
}

// The 'configure' command line action.
func configureFn(a Application) error {
	if len(*configFlag) == 0 {
		return fmt.Errorf("config flag to new or existing file is not set")
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

// The 'help' command line action.
func helpFn(a Application) error {
	flag.Usage()
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
