/* squares solves (?) the words in a square problem
   In the kind of puzzle this is addressing, letters can be used more than once.
*/

package main

import (
    "flag"
    "fmt"
    "github.com/gavinB-orange/squares/request"
    "runtime"
    "time"
)

var dFlag bool
var nSolvers int
var nMakers int
var verbose bool

func createTemplateReq() request.Request {
    var req request.Request
    req.Xsize = 4
    req.Ysize = 4
    // add words - don't include reversed or included words
    req.Addword("SWOT")
    req.Addword("PIG")
    req.Addword("AND")
    req.Addword("GNU")
    req.Addword("SPAR")
    req.Addword("WIN")
    req.Addword("TUB")
    req.Addword("TUG")
    req.Addword("GIN")
    req.Addword("WIG")
    req.Addword("GUT")
    req.Addword("GOT")
    req.Addword("PAN")
    req.Addword("SWIG")
    req.Addword("PING")
    req.Addword("GRAN")
    // add must-have chars
    req.SetMusts()
    return req
}

func main() {
    flag.BoolVar(&dFlag, "t", false, "Test mode - known good square provided")
    flag.BoolVar(&verbose, "v", false, "Verbose")
    flag.IntVar(&nSolvers, "s", 0, "Number of solvers")
    flag.IntVar(&nMakers, "m", 0, "Number of makers")
    flag.Parse()
    if nSolvers == 0 {
        nSolvers = runtime.NumCPU() * 2
    }
    if nMakers == 0 {
        nMakers = runtime.NumCPU()
    }
    fmt.Println("Running on a system with ", runtime.NumCPU()," cores.")
    // set up the template
    req := createTemplateReq()
    // get comms sorted out
    reqChan := make(chan request.Request, nSolvers)
    resChan := make(chan request.Request, nSolvers)
    // OK at this point have everything ready for an attempt
    // set off the solvers
    for i := 0; i<nSolvers; i++ {
        go request.Solver(i, reqChan, resChan, verbose)
    }
    // and makers
    for i := 0; i<nMakers; i++ {
        go func(id int) {
            seq := 0
            for {
                if dFlag {
                    time.Sleep(10 * 1000 * time.Millisecond)
                }
                // make an example square using the data given
                newr := req.MakeSquare(id, seq)
                seq++
                // queue it on the channel
                reqChan <- newr
            }
        }(i)
    }
    if dFlag {
        // set off a known-good example that should work
        testreq := req.MakeCorrectSquare(runtime.NumCPU() + 1)
        reqChan <- testreq
    }
    for {
        // wait for a res
        res := <-resChan
        if res.Found {
            fmt.Println("Found a valid square!")
            req.ShowSquare()
            return
        }
    }
}
