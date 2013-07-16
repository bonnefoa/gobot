package main

import (
        "flag"
        "os"
        "log"
        "runtime/pprof"
        "fmt"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
        flag.Parse()
        if *cpuprofile != "" {
                f, err := os.Create(*cpuprofile)
                if err != nil {
                        log.Fatal(err)
                }
                pprof.StartCPUProfile(f)
                defer pprof.StopCPUProfile()
        }
        fmt.Printf("Pi is %v\n", EstimatePi(400).FloatString(10))
}
