# gof
Perform gofmt for many .go files in parallel

```
$ gof

Usage: gof [options] [path ...]

Gof performs Gofmt in parallel

Options:
	-f="-l"                 Options passed to 'gofmt'
	-exclude=""             String in path to exclude
	-parallelism=10         Number of parallel executions of 'gofmt'
	-batch=10               Number of .go files fed into 'gofmt'

Path:
	folders or files

$ gof . $GOPATH/src/github.com/trung/gof
<..>/src/github.com/trung/gof/gofmt_issue.go
<..>/src/github.com/trung/gof/gofmt_issue_foo.go

$ gof -exclude="foo" $GOPATH/src/github.com/trung/gof
<..>/src/github.com/trung/gof/gofmt_issue.go
```

# Performance

```
$ cd $GOPATH/src/github.com/hashicorp/terraform
$ time gofmt -l `find . -name "*.go" | grep -v vendor`

real	0m5.652s
user	0m4.750s
sys	0m0.360s

$ time gof -exclude="vendor" .

real	0m1.786s
user	0m8.103s
sys	0m3.895s

$ time gof -exclude="vendor" -batch=100 .

real	0m1.212s
user	0m7.690s
sys	0m0.719s
```