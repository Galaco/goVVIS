package vvis

import (
	"flag"
)

type CmdArgs struct {
	File string
	Threads int				// -threads X
	Fast bool				// -fast
	Verbose bool			// -verbose
	UseRadius bool			// -radius_override X
	VisRadius int			// -radius_override X
	TraceClusterStart int	// -trace //not yet supported
	TraceClusterStop int	// -trace //not yet supported
	NoSort bool				// -nosort
	TmpIn string			// -tmpin
	LowPriority bool		// -low
	FullMiniDumps bool		// -fullMinidumps
}

// Read commandline arguments.
// Used to build the basic configuration for the vvis process.
func ParseCmdArguments() CmdArgs {
	file := flag.String("file", "", "Target bsp")
	threads := flag.Int("threads", 1, "Number of threads")
	fast := flag.Bool("fast", false, "Fast VVIS")
	verbose := flag.Bool("verbose", false, "Verbose output")
	useRadius := flag.Int("radius_override", 0, "Override vvis radius")
	nosort := flag.Bool("nosort", false, "No sorting")
	tmpin := flag.Bool("tmpin", false, "Use existing portalfiles")
	low := flag.Bool("low", false, "Low priority")
	fullMinidumps := flag.Bool("fullMinidumps", false, "Create full minidumps")

	flag.Parse()

	args := CmdArgs{}
	args.File = *file
	args.Threads = *threads
	args.Fast = *fast
	args.Verbose = *verbose

	if *useRadius > 0 {
		args.UseRadius = true
		args.VisRadius = *useRadius * *useRadius
	}

	args.NoSort = *nosort
	if *tmpin == true {
		args.TmpIn = "/tmp"
	}
	args.LowPriority = *low
	args.FullMiniDumps = *fullMinidumps

	//Not currently supported
	args.TraceClusterStart = -1
	args.TraceClusterStop = -1

	return args
}