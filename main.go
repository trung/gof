package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var (
	wg              sync.WaitGroup
	semaphore       chan int
	flagOptions     string
	flagParallelism int
	flagBatch       int
	flagExclude     string
)

func main() {
	os.Exit(doGof())
}

func doGof() int {
	flags := flag.NewFlagSet("gof", flag.ExitOnError)
	flags.Usage = printUsage
	flags.StringVar(&flagOptions, "f", "-l", "")
	flags.StringVar(&flagExclude, "exclude", "", "")
	flags.IntVar(&flagParallelism, "parallelism", 10, "")
	flags.IntVar(&flagBatch, "batch", 10, "")
	if err := flags.Parse(os.Args[1:]); err != nil {
		flags.Usage()
		return 1
	}
	paths := flags.Args()
	if len(paths) == 0 {
		flags.Usage()
		return 1
	}
	semaphore = make(chan int, flagParallelism)
	collectorChanel := make(chan string)
	done := make(chan bool)
	go collect(collectorChanel, done)
	// reading all .go files from paths
	for _, p := range paths {
		abs_p, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading path %s: %s\n", p, err)
			return 1
		}
		pathInfo, err := os.Stat(abs_p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading path %s: %s\n", abs_p, err)
			return 1
		}
		if pathInfo.IsDir() {
			//fmt.Printf("Reading *.go files from %s\n", abs_p)
			filepath.Walk(abs_p, func(path string, info os.FileInfo, err error) error {
				if strings.HasSuffix(path, ".go") {
					if flagExclude != "" && strings.Contains(path, flagExclude) {
						return nil
					}
					collectorChanel <- path
				}
				return nil
			})
		} else {
			if strings.HasSuffix(abs_p, ".go") {
				collectorChanel <- abs_p
			}
		}
	}
	collectorChanel <- "END"
	<-done
	wg.Wait()
	return 0
}
func collect(ch chan string, quit chan bool) {
	defer func() {
		quit <- true
	}()
	var count int
	paths := make([]string, 0, flagBatch)
	for {
		p := <-ch
		switch p {
		case "END":
			if count > 0 {
				execGofmt(paths)
			}
			return
		default:
			paths = append(paths, p)
			count++
			if count == flagBatch {
				execGofmt(paths)
				count = 0
				paths = make([]string, 0, flagBatch)
			}
		}
	}
}
func execGofmt(_paths []string) {
	wg.Add(1)
	go func(paths []string) {
		defer wg.Done()
		semaphore <- 1
		var stderr, stdout bytes.Buffer
		options := strings.Split(flagOptions, " ")
		arg := make([]string, len(paths)+len(options))
		copy(arg[:len(options)], options)
		copy(arg[len(options):], paths)
		// fmt.Printf("gofmt %s\n", strings.Join(arg, " "))
		cmd := exec.Command("gofmt", arg...)
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\nStderr: %s\n", err, stderr)
			return
		}
		files := stdout.String()
		if files != "" {
			fmt.Println(files)
		}
		<-semaphore
	}(_paths)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, helpText)
}

const helpText = `
Usage: gof [options] [path ...]

Gof performs Gofmt in parallel

Options:
	-f="-l"                 Options passed to 'gofmt'
	-exclude=""             String in path to exclude
	-parallelism=10         Number of parallel executions of 'gofmt'
	-batch=10               Number of .go files fed into 'gofmt'

Path:
	folders or files
`
