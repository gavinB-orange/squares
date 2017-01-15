/* The request package covers all details of the task to be done */
package request

import (
    "fmt"
    "math/rand"
    "runtime"
)

const verbose = true

func report(mess string) {
    if verbose {
        fmt.Printf(mess)
    }
}

type Request struct {
    owner int
    seq int
    Xsize int        // X-size of square
    Ysize int        // Y-size of square
    Found bool       // true if a valid solution
    Square [][]byte  // the square 
    Words [][]byte   // words used to make the square
    Musts []byte     // characters that must be present
    Extras []byte    // characters that appear more than once - only these should be used for padding
}

func assert(cond bool, message string) {
    if ! cond {
        panic(message)
    }
}

type Chars []byte

func (cs *Chars)PopFrom(wh int) byte {
    assert(wh < len(*cs), "Out of bounds")
    c := (*cs)[wh]
    *cs = append((*cs)[:wh], (*cs)[wh+1:]...)
    return c
}

type Coord struct {
    x int
    y int
}

type Coords []Coord

func (cds *Coords)PopFrom(wh int) Coord {
    assert(wh < len(*cds), "Out of bounds")
    c := (*cds)[wh]
    *cds = append((*cds)[:wh], (*cds)[wh+1:]...)
    return c
}

// AddWord : Add a word to the set of known words in the square
func (r *Request) Addword(w string) {
    r.Words = append(r.Words, []byte(w))
}

// SetMusts : From the known words, create the set of must-have chars
func (r *Request) SetMusts() {
    cm := make(map[byte]int)
    for _, w := range(r.Words) {
        for i := 0; i<len(w); i++ {
            cm[w[i]] ++
        }
    }
    extras := make([]byte, 0)
    musts := make([]byte, 0)
    for k, n := range cm {
        musts = append(musts, k)  // all known characters
        if n > 1 {
            for i:= 0; i<(n-1); i++ {
                extras = append(extras, k)  // add repeats to padding list reflecting how often used.
            }
        }
    }
    r.Musts = musts
    r.Extras = extras
}

// MakeSquare : Create a square containing a padded set of chars (as bytes)
func (r *Request)MakeSquare(ownerid int, seq int) Request{
    newr := Request{ownerid,
                    seq,
                    r.Xsize,
                    r.Ysize,
                    false,
                    make([][]byte, 0),
                    make([][]byte, 0),
                    make([]byte, 0), // musts not required when solving
                    make([]byte, 0)} // extras not required when solving
    for x:= 0; x<r.Xsize; x++ {
        col := make([]byte, r.Ysize)
        newr.Square = append(newr.Square, col)
    }
    for _, w := range(r.Words) {
        newr.Words = append(newr.Words, w)
    }
    chars := make(Chars, len(r.Musts))
    coords := make(Coords, 0)
    for x := 0; x < r.Xsize; x++ {
        for y := 0; y < r.Ysize; y++ {
            coords = append(coords, Coord{x, y})
        }
    }
    copy (chars, r.Musts)
    // pad with additional chars
    for i := 0; i < (r.Xsize * r.Ysize) - len(r.Musts); i++ {
        wh := int(rand.Int31n(int32(len(r.Extras))))
        chars = append(chars, r.Extras[wh])
    }
    assert(len(chars) == len(coords), "Lengths must match")
    // At this point have a slice of valid coords and a matching slice of chars
    for len(chars) > 0 {
        charwh := int(rand.Int31n(int32(len(chars))))
        coordwh := int(rand.Int31n(int32(len(coords))))
        c := chars.PopFrom(charwh)
        cd := coords.PopFrom(coordwh)
        newr.Square[cd.x][cd.y] = c
    }
    return newr // copy of original without the musts but with the square
}

func (r *Request) MakeCorrectSquare(ownerid int) Request {
    tr := r.MakeSquare(ownerid, -1)  // start with standard generate square
    tr.Square[0] = []byte("SWOT")
    tr.Square[1] = []byte("PIGU")
    tr.Square[2] = []byte("ANDB")
    tr.Square[3] = []byte("RGNU")
    return tr
}

func (r *Request) ShowSquare() {
    fmt.Println("============")
    fmt.Printf("Owner = %d, seq = %d :\n", r.owner, r.seq)
    for _, row := range(r.Square) {
        fmt.Println(row)
    }
    fmt.Println("============")
}

func (r *Request)findchar(sc Coord, c byte, debug bool) (bool, Coords) {
    nxlist := [...]int{-1, 0, 1}
    nylist := [...]int{-1, 0, 1}
    matchposlist := make(Coords, 0)
    for _, dx := range(nxlist) {
        for _, dy := range(nylist) {
            isok := dx + dy
            if ! ((isok == -1) || (isok == 1)) {
                continue
            }
            nx := sc.x + dx
            ny := sc.y + dy
            if debug {
                fmt.Println("Checking ", nx, ny)
            }
            if (nx >= 0) && (nx < r.Xsize) && (ny >= 0) && (ny < r.Ysize) {
                if r.Square[nx][ny] == c {
                    if debug {
                        fmt.Println("Found it at", nx, ny)
                    }
                    matchposlist = append(matchposlist, Coord{nx, ny})
                }
            }
        }
    }
    return len(matchposlist)>0, matchposlist
}

func (r *Request)walkword(sc Coord, target []byte, depth int, debug bool) bool {
    if len(target) <= 1 {
        return true
    }
    if debug {
        fmt.Println("SC :", sc, "depth :", depth, "target :", target)
    }
    for i := 1; i<len(target); i++ { // we know 0 is already OK
        if debug {
            fmt.Println("sc :", sc, "targetchar = ", target[i])
        }
        ok, newsclist := r.findchar(sc, target[i], debug)
        if debug {
            fmt.Println("findchar result :", ok, newsclist)
        }
        if ! ok {
            if debug {
                fmt.Println("walkword reports failure for", target)
            }
            return false
        }
        for _, newsc := range(newsclist) {
            if r.walkword(newsc, target[1:], depth + 1, debug) {
                if debug {
                    fmt.Println("walkword (depth ", depth, ") returns success for ", target)
                }
                return true  // a child reports success
            }
        }
        // all children have reported failure
        return false
    }
    // should never get here
    fmt.Println("SC = ", sc, "target = ", target, "depth = ", depth)
    r.ShowSquare()
    panic("How did I get here?")
    return false
}

// Given a target byte slide, walk the square and find it
func (r *Request) FindWord(target []byte, debug bool) bool {
    if len(target) <= 0 {
        fmt.Println("target word empty")
        return false
    }
    // Get first char and get a list of possible starting points.
    coordlist := make(Coords, 0)
    for x := 0; x<r.Xsize; x++ {
        for y := 0; y<r.Ysize; y++ {
            if r.Square[x][y] == target[0] {
                coordlist = append(coordlist, Coord{x, y})
            }
        }
    }
    if debug {
        fmt.Println("coordlist >", coordlist)
    }
    // For each starting point, walk the word and see if the sequence can be found.
    // if found, return true.
    for _, sc := range(coordlist) {
        if r.walkword(sc, target, 0, debug) {
            return true
        }
    }
    return false
}

// Try to find all the words
func Solver(id int, in chan Request, out chan Request, verbose bool) {
    mythresh := 0
    for req := range(in) {
        if verbose {
            req.ShowSquare()
        }
        special := verbose || (req.owner > runtime.NumCPU())  // flags a test case
        ok := true
        for n, wd := range(req.Words) {
            if special || (n > mythresh) {
                fmt.Println("Solver", id, "->", n)
                req.ShowSquare()
                mythresh = n
            }
            ok = ok && req.FindWord(wd, special)
            if ! ok {
                break
            }
        }
        req.Found = ok
        out <- req
    }
}
