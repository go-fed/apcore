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
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-fed/apcore/app"
	"github.com/go-fed/apcore/framework"
	"github.com/go-fed/apcore/util"
)

var (
	// Flags for apcore
	debugFlag        = flag.Bool("debug", false, "Enable the development server on localhost & other developer quality of life features")
	systemLogFlag    = flag.Bool("syslog", false, "Also logs to system (stdout and stderr) if logging to a file")
	infoLogFileFlag  = flag.String("info_log_file", "", "Log file for info, defaults to stdout")
	errorLogFileFlag = flag.String("error_log_file", "", "Log file for errors, defaults to stderr")
	configFlag       = flag.String("config", "config.ini", "Path to the configuration file")
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
			framework.ClarkeSays("Hi, I'm Clarke the Cow! When you run certain commands, I will help guide you "+
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
	Action      func(app.Application) error
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
func serveFn(a app.Application) error {
	s, err := newServer(*configFlag, a, *debugFlag, schemeFromFlags())
	if err != nil {
		return err
	}
	interruptCh := make(chan os.Signal, 2)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interruptCh
		s.Stop()
	}()
	return s.Start()
}

// The 'new' command line action.
func guideNewFn(a app.Application) error {
	sw := a.Software()
	fmt.Println(framework.ClarkeSays(fmt.Sprintf(`
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
	fmt.Println(framework.ClarkeSays(`
Configuration wizardry complete! It is a good idea to check that configuration
file for additional options before serving traffic. You can always re-run the
wizard using the "configure" action. Now let's initialize the database!`))
	err = initDbFn(a)
	if err != nil {
		return err
	}
	fmt.Println(framework.ClarkeSays(`
Whew! That can manually be done using the "init-db" action in the future. Next,
let's initialize your first administrator account in the database.`))
	err = initAdminFn(a)
	if err != nil {
		return err
	}
	fmt.Println(framework.ClarkeSays(`
Moo~! That was the "init-admin" action. We are done, but before you run the
"serve" action, please do double check your configuration file! Bye bye!`))
	return nil
}

// The 'init-db' command line action.
func initDbFn(a app.Application) error {
	fmt.Println(framework.ClarkeSays(`
We're connecting to the database using the specs in the config file, creating
tables, and then closing all connections.`))
	err := doCreateTables(*configFlag, a, *debugFlag, schemeFromFlags())
	if err != nil {
		return err
	}
	fmt.Println(framework.ClarkeSays(`Database initialization udderly complete!`))
	return nil
}

// The 'init-admin' command line action.
func initAdminFn(a app.Application) error {
	msg := `Moo~, let's create an administrative account!`
	if *debugFlag {
		msg += "\nWARNING: Creating a user in debug mode will NOT work in production and MUST ONLY be used for development"
	}
	fmt.Println(framework.ClarkeSays(msg))
	err := doInitAdmin(*configFlag, a, *debugFlag)
	if err != nil {
		return err
	}
	fmt.Println(framework.ClarkeSays(`New admin account successfully created! Moo~`))
	return nil
}

// The 'configure' command line action.
func configureFn(a app.Application) error {
	if len(*configFlag) == 0 {
		return fmt.Errorf("config flag to new or existing file is not set")
	}
	exists := false
	if _, err := os.Stat(*configFlag); err == nil {
		exists = true
		cont, err := framework.PromptFileExistsContinue(*configFlag)
		if err != nil {
			return err
		}
		if !cont {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot modify configuration: %s", err)
	}
	cfg, err := framework.PromptNewConfig(*configFlag)
	if err != nil {
		return err
	}
	util.InfoLogger.Info("Calling application to get default config options")
	acfg := a.NewConfiguration()
	if exists {
		cont, err := framework.PromptOverwriteExistingFile(*configFlag)
		if err != nil {
			return err
		}
		if !cont {
			util.InfoLogger.Infof("Did not overwrite configuration file at %s", *configFlag)
			return nil
		}
	}
	err = framework.SaveConfigFile(*configFlag, cfg, acfg)
	if err != nil {
		return err
	}
	util.InfoLogger.Infof("Successfully wrote configuration file to %s", *configFlag)
	return nil
}

// The 'version' command line action.
func versionFn(a app.Application) error {
	fmt.Fprintf(os.Stdout, "%s; %s\n", a.Software(), apCoreSoftware())
	return nil
}

// The 'help' command line action.
func helpFn(a app.Application) error {
	flag.Usage()
	return nil
}

func schemeFromFlags() string {
	if *debugFlag {
		return "http"
	}
	return "https"
}
