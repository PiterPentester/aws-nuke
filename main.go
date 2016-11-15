package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/fatih/color"
)

var (
	// will be overwritten on build
	version = "unknown"
)

var (
	ReasonSkip            = *color.New(color.FgYellow)
	ReasonError           = *color.New(color.FgRed)
	ReasonRemoveTriggered = *color.New(color.FgGreen)
	ReasonWaitPending     = *color.New()
	ReasonSuccess         = *color.New(color.FgGreen)
	ColorID               = *color.New(color.Bold)
)

func Log(r Resource, c color.Color, msg string) {
	fmt.Printf("[%s] ", time.Now().Format(time.RFC3339))
	fmt.Print(strings.Split(fmt.Sprintf("%T", r), ".")[1]) // hackey
	fmt.Printf(" - ")
	ColorID.Printf("'%s'", r.String())
	fmt.Printf(" - ")
	c.Printf("%s\n", msg)
}

func LogErrorf(err error) {
	out := color.New(color.FgRed)
	trace := fmt.Sprintf("%+v", err)
	out.Println(trace)
	out.Println("")
}

func main() {
	fmt.Printf("Running aws-nuke version %s.\n", version)

	var (
		profile  = flag.String("profile", "", "profile to nuke")
		region   = flag.String("region", "eu-west-1", "profile to nuke")
		noDryRun = flag.Bool("no-dry-run", false,
			"Actualy delete found resources.")
		noWait = flag.Bool("no-wait", false,
			"Do not wait for resource removal. This is faster, "+
				"but you may have to run the nuke multiple times.")
		exitOnError = flag.Bool("exit-on-error", false,
			"Immediately exit, when an error orccurs. Otherwise "+
				"it would continue with other resources.")
		retry = flag.Bool("retry", false,
			"Retry later, if an error occurs. Retries until all "+
				"resources succeeded.")
	)

	flag.Parse()

	if *profile == "" {
		fmt.Printf("You have to specify -profile.\n")
		os.Exit(1)
	}

	if strings.Contains(strings.ToLower(*profile), "prod") {
		fmt.Printf("The profile name contains the substring 'prod'. Refusing to nuke it.\n")
		os.Exit(1)
	}

	if !*noDryRun {
		fmt.Printf("Dry run: do real delete with '--no-dry-run'.\n")
	}

	fmt.Println()

	n := &Nuke{
		session: session.New(&aws.Config{
			Region:      region,
			Credentials: credentials.NewSharedCredentials("", *profile),
		}),
		dry:       !*noDryRun,
		wait:      !*noWait,
		earlyExit: *exitOnError,
		retry:     *retry,

		queue:    []Resource{},
		waiting:  []Resource{},
		skipped:  []Resource{},
		failed:   []Resource{},
		finished: []Resource{},
	}

	n.Run()
}