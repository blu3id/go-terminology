// SNOMED-CT command line utility and terminology server
//
// Copyright 2018 Mark Wardle / Eldrix Ltd
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/wardle/go-terminology/server"
	"github.com/wardle/go-terminology/terminology"
)

// automatically populated by linker flags
var version string
var build string

// commands and flags
var doVersion = flag.Bool("version", false, "show version information")
var doImport = flag.Bool("import", false, "import SNOMED-CT data files from directories specified")
var runserver = flag.Bool("server", false, "run terminology server")
var precompute = flag.Bool("precompute", false, "perform precomputations and optimisations")
var reset = flag.Bool("reset", false, "clear precomputations and optimisations")
var stats = flag.Bool("status", false, "get statistics")
var export = flag.Bool("export", false, "export expanded descriptions in delimited protobuf format to stdout")

// general flags
var database = flag.String("db", "", "filename of database to open or create (e.g. ./snomed.db).\nCan also be set using environmental variable GTS_DATABASE")
var lang = flag.String("lang", "en-GB", "language tags to be used, default 'en-GB'.")
var verbose = flag.Bool("v", false, "show verbose information")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file specified")
var port = flag.Int("port", 8081, "port to use for http server")
var grpc = flag.Int("grpc", 9091, "port to use for grpc server")

func main() {
	flag.Parse()
	if *doVersion {
		fmt.Printf("%s v%s (%s)\n", os.Args[0], version, build)
		os.Exit(1)
	}
	if *database == "" {
		*database = os.Getenv("GTS_DATABASE")
		if *database == "" {
			fmt.Printf("error: missing mandatory database file\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}
	readOnly := true
	if *doImport || *precompute || *reset {
		readOnly = false
	}
	svc, err := terminology.NewService(*database, readOnly)
	if err != nil {
		log.Fatalf("couldn't open database: %v", err)
	}
	defer svc.Close()

	// turn on CPU profiling if a profile file is specified
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	help := true // show help, unless the user gives at least one command

	// useful for user to be able to clear precomputations in case of wishing to share
	// a data file with another; the recipient can easily re-run precomputations
	if *reset {
		help = false
		svc.ClearPrecomputations()
	}
	// perform import if an import root is specified
	if *doImport {
		help = false
		if flag.NArg() == 0 {
			log.Fatalf("no input directories specified")
		}
		runImport2(svc, 50000, *verbose)
		svc.CompactStore()
		time.Sleep(60 * time.Second)
		//for _, filename := range flag.Args() {
		//	ctx := context.Background()
		//	importer := terminology.NewImporter(svc, 5000, 0, *verbose)
		//	importer.Import(ctx, filename)
		//}
	}

	// perform precomputations if requested
	if *precompute {
		help = false
		//svc.PerformPrecomputations(context.Background(), 500, *verbose)
		svc.RunIndexBuilder(50000, *lang, *verbose)
		svc.CompactStore()
		time.Sleep(60 * time.Second)
	}

	// get statistics on store
	if *stats {
		help = false
		s, err := svc.Statistics(*lang, *verbose)
		if err != nil && err != terminology.ErrDatabaseNotInitialised {
			panic(err)
		}
		fmt.Printf("%v", s)
	}

	// export descriptions data in expanded denormalised format
	if *export {
		help = false
		err := svc.Export(*lang)
		if err != nil {
			log.Fatal(err)
		}
	}

	// optionally run a terminology server
	if *runserver {
		help = false
		opts := server.DefaultOptions
		opts.RESTPort = *port
		opts.RPCPort = *port + 1
		if *grpc != 0 {
			opts.RPCPort = *grpc
		}
		opts.DefaultLanguage = *lang
		log.Fatal(server.RunServer(svc, *opts))
	}
	if help {
		flag.PrintDefaults()
	}
}
