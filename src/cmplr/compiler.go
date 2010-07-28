// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package compiler

import (
    "os"
    "container/vector"
    "fmt"
    "utilz/stringset"
    "utilz/handy"
    "cmplr/dag"
    "path"
)


type Compiler struct {
    root, arch, suffix, executable string
    dryrun                         bool
    includes                       []string
}

func New(root, arch string, dryrun bool, include []string) *Compiler {
    c := new(Compiler)
    c.root = root
    c.arch, c.suffix = archNsuffix(arch)
    c.executable = findCompiler(c.arch)
    c.dryrun = dryrun
    c.includes = include
    return c
}

func findCompiler(arch string) string {

    var lookingFor string
    switch arch {
    case "arm":
        lookingFor = "5g"
    case "amd64":
        lookingFor = "6g"
    case "386":
        lookingFor = "8g"
    }

    fullPath := handy.Which(lookingFor)
    if fullPath == "" {
        die("[ERROR] could not find compiler\n")
    }
    return fullPath
}

func findLinker(arch string) string {

    var lookingFor string
    switch arch {
    case "arm":
        lookingFor = "5l"
    case "amd64":
        lookingFor = "6l"
    case "386":
        lookingFor = "8l"
    }

    fullPath := handy.Which(lookingFor)
    if fullPath == "" {
        die("[ERROR] could not find linker\n")
    }
    return fullPath
}


func archNsuffix(arch string) (a, s string) {

    if arch == "" {
        a = os.Getenv("GOARCH")
    } else {
        a = arch
    }

    switch a {
    case "arm":
        s = ".5"
    case "amd64":
        s = ".6"
    case "386":
        s = ".8"
    default:
        die("[ERROR] unknown architecture: %s\n", a)
    }

    return a, s
}

func (c *Compiler) String() string {
    s := "Compiler{ root=%s, arch=%s, suffix=%s, executable=%s }"
    return fmt.Sprintf(s, c.root, c.arch, c.suffix, c.executable)
}

func (c *Compiler) CreateArgv(pkgs *vector.Vector) {

    var argv []string

    includeLen := c.extraPkgIncludes()

    for y := 0; y < pkgs.Len(); y++ {
        pkg, _ := pkgs.At(y).(*dag.Package) //safe cast, only Packages there

        argv = make([]string, 5+pkg.Files.Len()+(includeLen*2))
        i := 0
        argv[i] = c.executable
        i++
        argv[i] = "-I"
        i++
        argv[i] = c.root
        i++
        if includeLen > 0 {
            for y := 0; y < includeLen; y++ {
                argv[i] = "-I"
                i++
                argv[i] = c.includes[y]
                i++
            }
        }
        argv[i] = "-o"
        i++
        argv[i] = path.Join(c.root, pkg.Name) + c.suffix
        i++

        for z := 0; z < pkg.Files.Len(); z++ {
            argv[i] = pkg.Files.At(z)
            i++
        }

        pkg.Argv = argv
    }
}

func (c *Compiler) SerialCompile(pkgs *vector.Vector) {

    for y := 0; y < pkgs.Len(); y++ {
        pkg, _ := pkgs.At(y).(*dag.Package) //safe cast, only Packages there

        if c.dryrun {
            dryRun(pkg.Argv)
        } else {
            fmt.Println("compiling:", pkg.Name)
            handy.StdExecve(pkg.Argv, true)
        }
    }
}

func (c *Compiler) ParallelCompile(pkgs *vector.Vector) {

    var localDeps *stringset.StringSet
    var compiledDeps *stringset.StringSet
    var pkg, cpkg *dag.Package
    var y, z int
    var parallel *vector.Vector

    localDeps    = stringset.New()
    compiledDeps = stringset.New()

    for y = 0; y < pkgs.Len(); y++ {
        pkg, _ = pkgs.At(y).(*dag.Package)
        localDeps.Add( pkg.Name )
    }

    parallel = new(vector.Vector)

    for y = 0; y < pkgs.Len(); {

        pkg, _ = pkgs.At(y).(*dag.Package)

        if ! pkg.Ready( localDeps, compiledDeps ) {

            c.compileMultipe( parallel )

            for z = 0; z < parallel.Len(); z++ {
                cpkg, _ = parallel.At(z).(*dag.Package)
                compiledDeps.Add( cpkg.Name )
            }

            parallel = new(vector.Vector)

        }else{
            parallel.Push( pkg )
            y++
        }
    }

    if parallel.Len() > 0 {
        c.compileMultipe( parallel )
    }

}

func (c *Compiler) compileMultipe(pkgs *vector.Vector){

    var ok bool
    var max int = pkgs.Len()
    var pkg *dag.Package
    var trouble bool = false

    if max == 0 {
        die("[ERROR] trying to compile 0 packages in parallel\n")
    }

    if max == 1 {
        pkg, _ = pkgs.At(0).(*dag.Package)
        fmt.Println("compiling:", pkg.Name)
        handy.StdExecve(pkg.Argv, true)
    }else{

        ch := make(chan bool, pkgs.Len())

        for y := 0; y < max; y++ {
            pkg, _ := pkgs.At(y).(*dag.Package)
            fmt.Println("compiling:", pkg.Name)
            go gCompile( pkg.Argv, ch )
        }

        // drain channel (make sure all jobs are finished)
        for z := 0; z < max; z++ {
            ok = <-ch
            if !ok {
                trouble = true
            }
        }
    }

    if trouble {
        die("[ERROR] failed batch compile job\n")
    }

}

func gCompile(argv []string, c chan bool){
    ok := handy.StdExecve(argv, false) // don't exit on error
    c<-ok
}

// for removal of temoprary packages created for testing and so on..
func (c *Compiler) DeletePackages(pkgs *vector.Vector) bool {

    var ok = true
    var e os.Error

    for i := 0; i < pkgs.Len(); i++ {
        pkg, _ := pkgs.At(i).(*dag.Package) //safe cast, only Packages there

        for y := 0; y < pkg.Files.Len(); y++ {
            e = os.Remove(pkg.Files.At(y))
            if e != nil {
                ok = false
                fmt.Fprintf(os.Stderr, "[ERROR] %s\n", e)
            }
        }
        if !c.dryrun {
            pcompile := path.Join(c.root, pkg.Name) + c.suffix
            e = os.Remove(pcompile)
            if e != nil {
                ok = false
                fmt.Fprintf(os.Stderr, "[ERROR] %s\n", e)
            }
        }
    }

    return ok
}

func (c *Compiler) ForkLink(pkgs *vector.Vector, output string, static bool) {

    var mainPKG *dag.Package

    gotMain := new(vector.Vector)

    for i := 0; i < pkgs.Len(); i++ {
        pk, _ := pkgs.At(i).(*dag.Package)
        if pk.ShortName == "main" {
            gotMain.Push(pk)
        }
    }

    if gotMain.Len() == 0 {
        die("[ERROR] (linking) no main package found\n")
    }

    if gotMain.Len() > 1 {
        choice := mainChoice(gotMain)
        mainPKG, _ = gotMain.At(choice).(*dag.Package)
    } else {
        mainPKG, _ = gotMain.Pop().(*dag.Package)
    }

    includeLen := c.extraPkgIncludes()
    staticXtra := 0
    if static {
        staticXtra++
    }

    linker := findLinker(c.arch)
    compiled := path.Join(c.root, mainPKG.Name) + c.suffix

    argv := make([]string, 6+(includeLen*2)+staticXtra)
    i := 0
    argv[i] = linker
    i++
    argv[i] = "-o"
    i++
    argv[i] = output
    i++
    argv[i] = "-L"
    i++
    argv[i] = c.root
    i++
    if static {
        argv[i] = "-d"
        i++
    }
    if includeLen > 0 {
        for y := 0; y < includeLen; y++ {
            argv[i] = "-L"
            i++
            argv[i] = c.includes[y]
            i++
        }
    }
    argv[i] = compiled
    i++

    if c.dryrun {
        dryRun(argv)
    } else {
        fmt.Println("linking  :", output)
        handy.StdExecve(argv, true)
    }
}

func mainChoice(pkgs *vector.Vector) int {

    fmt.Println("\n More than one main package found\n")

    for i := 0; i < pkgs.Len(); i++ {
        pk, _ := pkgs.At(i).(*dag.Package)
        fmt.Printf(" type %2d  for: %s\n", i, pk.Name)
    }

    var choice int

    fmt.Printf("\n type your choice: ")

    n, e := fmt.Scanf("%d", &choice)

    if e != nil {
        die("%s\n", e)
    }
    if n != 1 {
        die("failed to read input\n")
    }

    if choice >= pkgs.Len() || choice < 0 {
        die(" bad choice: %d\n", choice)
    }

    fmt.Printf(" chosen main-package: %s\n\n", pkgs.At(choice).(*dag.Package).Name)

    return choice
}

func die(strfmt string, v ...interface{}) {
    fmt.Fprintf(os.Stderr, strfmt, v)
    os.Exit(1)
}


func dryRun(argv []string) {
    var cmd string

    for i := 0; i < len(argv); i++ {
        cmd = fmt.Sprintf("%s %s ", cmd, argv[i])
    }

    fmt.Printf("%s || exit 1\n", cmd)
}

func (c *Compiler) extraPkgIncludes() int {
    if c.includes != nil && len(c.includes) > 0 {
        return len(c.includes)
    }
    return 0
}
