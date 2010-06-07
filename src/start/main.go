// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package main

import (
    "os"
    "fmt"
    "path"
    "utilz/walker"
    "cmplr/compiler"
    "cmplr/dag"
    "container/vector"
    "parse/gopt"
    "strings"
    "utilz/handy"
    "io/ioutil"
    "regexp"
)


// option parser object (struct)
var getopt *gopt.GetOpt

// list of files to compile
var files *vector.StringVector

// variables for the different options
var arch, gdtest, output, srcdir, bmatch, match, rewRule, tabWidth string
var dryrun, test, testVerbose, static, noComments, tabIndent, listing bool
var gofmt, printInfo, sortInfo, cleanTree, needsHelp, needsVersion bool
var includes []string = nil


func init() {

    // initialize option parser
    getopt = gopt.New()

    // some string defaults
    gdtest = "gdtest"
    srcdir = "src"

    // add all options (bool/string)
    getopt.BoolOption("-h -help --help help")
    getopt.BoolOption("-c -clean --clean clean")
    getopt.BoolOption("-S -static --static")
    getopt.BoolOption("-v -version --version version")
    getopt.BoolOption("-s -sort --sort sort")
    getopt.BoolOption("-p -print --print")
    getopt.BoolOption("-d -dryrun --dryrun")
    getopt.BoolOption("-t -test --test")
    getopt.BoolOption("-l -list --list")
    getopt.BoolOption("-V -verbose --verbose")
    getopt.BoolOption("-f -fmt --fmt")
    getopt.BoolOption("-no-comments --no-comments")
    getopt.BoolOption("-tab --tab")
    getopt.StringOption("-a -arch --arch -arch= --arch=")
    getopt.StringOption("-I")
    getopt.StringOption("-tabwidth --tabwidth -tabwidth= --tabwidth=")
    getopt.StringOption("-rew-rule --rew-rule -rew-rule= --rew-rule=")
    getopt.StringOption("-o -output --output -output= --output=")
    getopt.StringOption("-b -benchmarks --benchmarks -benchmarks= --benchmarks=")
    getopt.StringOption("-m -match --match -match= --match=")
    getopt.StringOption("-test-bin --test-bin -test-bin= --test-bin=")

    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".go") &&
            !strings.HasSuffix(s, "_test.go")
    }

    // override IncludeDir to make walker ignore 'hidden' directories
    walker.IncludeDir = func(s string) bool {
        _, dirname := path.Split(s)
        return dirname[0] != '.'
    }

}

func gotRoot() {
    if os.Getenv("GOROOT") == "" {
        fmt.Fprintf(os.Stderr, "[ERROR] missing GOROOT\n")
        os.Exit(1)
    }
}

func findTestFilesAlso() {
    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".go")
    }
}

func main() {

    var ok bool
    var argv, args []string

    // default config location 1 $HOME/.gdrc
    argv, ok = getConfigArgv("HOME")

    if ok {
        args = parseArgv(argv)
        if len(args) > 0 {
            fmt.Fprintf(os.Stderr, "[WARNING] non-option arguments in config file\n")
        }
    }

    // default config location 2 $PWD/.gdrc
    argv, ok = getConfigArgv("PWD")

    if ok {
        args = parseArgv(argv)
        if len(args) > 0 {
            fmt.Fprintf(os.Stderr, "[WARNING] non-option arguments in config file\n")
        }
    }

    // command line arguments overrides/appends config
    args = parseArgv(os.Args[1:])

    if len(args) > 0 {
        if len(args) > 1 {
            fmt.Fprintf(os.Stderr, "[WARNING] len(input directories) > 1\n")
        }
        srcdir = args[0]
    }


    // stuff that can be done without $GOROOT
    if listing {
        printListing();
        os.Exit(0);
    }

    if needsHelp {
        printHelp();
        os.Exit(0);
    }

    if needsVersion {
        printVersion();
        os.Exit(0);
    }

    // delete all object/archive files
    if cleanTree {
        rm865(srcdir, dryrun)
        os.Exit(0)
    }

    files = findFiles(srcdir)

    // gofmt on all files gathered
    if gofmt {
        formatFiles(files, dryrun, tabIndent, noComments, rewRule, tabWidth)
        os.Exit(0)
    }

    // parse the source code, look for dependencies
    dgrph := dag.New()
    dgrph.Parse(srcdir, files)

    // print collected dependency info
    if printInfo {
        dgrph.PrintInfo()
        os.Exit(0)
    }


    gotRoot();//?


    // sort graph based on dependencies
    dgrph.GraphBuilder(includes)
    sorted := dgrph.Topsort()

    // print packages sorted, if that's what's desired
    if sortInfo {
        for pkg := range sorted.Iter() {
            rpkg, _ := pkg.(*dag.Package)
            fmt.Printf("%s\n", rpkg.Name)
        }
        os.Exit(0)
    }

    // compile
    kompiler := compiler.New(srcdir, arch, dryrun, includes)
    kompiler.ForkCompile(sorted)

    // test
    if test {
        testMain, testDir := dgrph.MakeMainTest(srcdir)
        kompiler.ForkCompile(testMain)
        kompiler.ForkLink(testMain, gdtest, false)
        kompiler.DeletePackages(testMain)
        rmError := os.Remove(testDir)
        if rmError != nil {
            fmt.Fprintf(os.Stderr, "[ERROR] failed to remove testdir: %s\n", testDir)
        }
        testArgv := createTestArgv(gdtest, bmatch, match, testVerbose)
        if !dryrun {
            tstring := "testing  : "
            if testVerbose {
                tstring += "\n"
            }
            fmt.Printf(tstring)
            ok := handy.StdExecve(testArgv, false)
            e := os.Remove(gdtest)
            if e != nil {
                fmt.Fprintf(os.Stderr, "[ERROR] %s\n", e)
            }
            if !ok {
                os.Exit(1)
            }
        }
    }

    if output != "" {
        kompiler.ForkLink(sorted, output, static)
    }

}


// syntax for config files are identical to command line
// options, i.e., write command line options to the config file 
// and everything is fine, comments start with a '#' sign.
func getConfigArgv(where string) (argv []string, ok bool) {

    location := os.Getenv(where)

    if location == "" {
        return nil, false
    }

    configFile := path.Join(location, ".gdrc")
    configDir, e := os.Stat(configFile)

    if e != nil {
        return nil, false
    }

    if !configDir.IsRegular() {
        return nil, false
    }

    b, e := ioutil.ReadFile(configFile)

    if e != nil {
        fmt.Fprintf(os.Stderr, "[WARNING] failed to read config file\n")
        fmt.Fprintf(os.Stderr, "[WARNING] %s \n", e)
        return nil, false
    }

    comStripRegex := regexp.MustCompile("#[^\n]*\n?")
    blankRegex := regexp.MustCompile("[\n\t \r]+")

    rmComments  := comStripRegex.ReplaceAllString(string(b), "")
    rmNewLine   := blankRegex.ReplaceAllString(rmComments, " ")

    pureOptions := strings.TrimSpace( rmNewLine );

    if pureOptions == "" {
        return nil, false;
    }

    argv = strings.Split(pureOptions, " ", -1)

    return argv, true
}


func parseArgv(argv []string) (args []string) {

    args = getopt.Parse(argv)

    if getopt.IsSet("-help") {
        needsHelp = true
    }

    if getopt.IsSet("-list") {
        listing = true;
    }

    if getopt.IsSet("-version") {
        needsVersion = true
    }

    if getopt.IsSet("-dryrun") {
        dryrun = true
    }

    if getopt.IsSet("-print") {
        printInfo = true
    }

    if getopt.IsSet("-sort") {
        sortInfo = true
    }

    if getopt.IsSet("-static") {
        static = true
    }

    if getopt.IsSet("-clean") {
        cleanTree = true
    }

    if getopt.IsSet("-arch") {
        arch = getopt.Get("-a")
    }

    if getopt.IsSet("-I") {
        if includes == nil {
            includes = getopt.GetMultiple("-I")
        }else{
            tmp    := getopt.GetMultiple("-I");
            joined := make([]string, ( len(includes) + len(tmp) ));

            var i, j int;

            for i = 0; i < len(includes); i++ {
                joined[i] = includes[i];
            }
            for j = 0; j < len(tmp); j++ {
                joined[i+j] = tmp[j];
            }

            includes = joined;
        }
    }

    if getopt.IsSet("-output") {
        output = getopt.Get("-o")
    }

    // for gotest
    if getopt.IsSet("-test") {
        test = true
        findTestFilesAlso()
    }

    if getopt.IsSet("-benchmarks") {
        bmatch = getopt.Get("-b")
    }

    if getopt.IsSet("-match") {
        match = getopt.Get("-m")
    }

    if getopt.IsSet("-verbose") {
        testVerbose = true
    }

    if getopt.IsSet("-test-bin") {
        gdtest = getopt.Get("-test-bin")
    }

    // for gofmt
    if getopt.IsSet("-fmt") {
        gofmt = true
    }

    if getopt.IsSet("-no-comments") {
        noComments = true
    }

    if getopt.IsSet("-rew-rule") {
        rewRule = getopt.Get("-rew-rule")
    }

    if getopt.IsSet("-tab") {
        tabIndent = true
    }

    if getopt.IsSet("-tabwidth") {
        tabWidth = getopt.Get("-tabwidth")
    }

    getopt.Reset()
    return args
}

func createTestArgv(prg, bmatch, match string, tverb bool) []string {
    var numArgs int = 1
    pwd, e := os.Getwd()
    if e != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] could not locate working directory\n")
        os.Exit(1)
    }
    arg0 := path.Join(pwd, prg)
    if bmatch != "" {
        numArgs += 2
    }
    if match != "" {
        numArgs += 2
    }
    if tverb {
        numArgs++
    }

    var i = 1
    argv := make([]string, numArgs)
    argv[0] = arg0
    if bmatch != "" {
        argv[i] = "-benchmarks"
        i++
        argv[i] = bmatch
        i++
    }
    if match != "" {
        argv[i] = "-match"
        i++
        argv[i] = match
        i++
    }
    if tverb {
        argv[i] = "-v"
    }
    return argv
}

func findFiles(pathname string) *vector.StringVector {
    okDirOrDie(pathname)
    return walker.PathWalk(path.Clean(pathname))
}

func okDirOrDie(pathname string) {

    var dir *os.FileInfo
    var staterr os.Error

    dir, staterr = os.Stat(pathname)

    if staterr != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] %s\n", staterr)
        os.Exit(1)
    } else if !dir.IsDirectory() {
        fmt.Fprintf(os.Stderr, "[ERROR] %s: is not a directory\n", pathname)
        os.Exit(1)
    }
}

func formatFiles(files *vector.StringVector, dryrun, tab, noC bool, rew, tw string) {

    var i int = 0
    var argvLen int = 0
    var argv []string
    var tabWidth string = "-tabwidth=4"
    var useTabs string = "-tabindent=false"
    var comments string = "-comments=true"
    var rewRule string = ""
    var fmtexec string = handy.Which("gofmt")

    if tw != "" {
        tabWidth = "-tabwidth=" + tw
    }
    if noC {
        comments = "-comments=false"
    }
    if rew != "" {
        rewRule = rew
        argvLen++
    }
    if tab {
        useTabs = "-tabindent=true"
    }

    argv = make([]string, 6+argvLen)

    if fmtexec == "" {
        fmt.Fprintf(os.Stderr, "[ERROR] could not find: gofmt\n")
        os.Exit(1)
    }

    argv[i] = fmtexec
    i++
    argv[i] = "-w=true"
    i++
    argv[i] = tabWidth
    i++
    argv[i] = useTabs
    i++
    argv[i] = comments
    i++

    if rewRule != "" {
        argv[i] = "-r=" + rewRule
        i++
    }

    for fileName := range files.Iter() {
        argv[i] = fileName
        if !dryrun {
            fmt.Printf("gofmt : %s\n", fileName)
            _ = handy.StdExecve(argv, true)
        } else {
            fmt.Printf(" %s\n", strings.Join(argv, " "))
        }
    }

}

func rm865(srcdir string, dryrun bool) {

    // override IncludeFile to make walker pick up only .[865] files
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".8") ||
            strings.HasSuffix(s, ".6") ||
            strings.HasSuffix(s, ".5") ||
            strings.HasSuffix(s, ".a")

    }

    okDirOrDie(srcdir)

    compiled := walker.PathWalk(path.Clean(srcdir))

    for s := range compiled.Iter() {
        if !dryrun {
            fmt.Printf("rm: %s\n", s)
            e := os.Remove(s)
            if e != nil {
                fmt.Fprintf(os.Stderr, "[ERROR] could not delete file: %s\n", s)
            }
        } else {
            fmt.Printf("[dryrun] rm: %s\n", s)
        }
    }
}

func printHelp() {
    var helpMSG string = `
  Godag is a compiler front-end for golang,
  its main purpose is to help build projects
  which are pure Go-code without Makefiles.
  Hopefully it simplifies testing as well.

  usage: gd [OPTIONS] src-directory

  options:

  -h --help            print this message and quit
  -v --version         print version and quit
  -l --list            list option values and quit
  -p --print           print package info collected
  -s --sort            print legal compile order
  -o --output          link main package -> output
  -S --static          statically link binary
  -a --arch            architecture (amd64,arm,386)
  -d --dryrun          print what gd would do (stdout)
  -c --clean           rm *.[a865] from src-directory
  -I                   import package directories
  -t --test            run all unit-tests
  -b --benchmarks      pass argument to unit-test
  -m --match           pass argument to unit-test
  -V --verbose         pass argument '-v' to unit-test
  --test-bin           name of test-binary (default: gdtest)
  -f --fmt             run gofmt on src and exit
  --rew-rule           pass rewrite rule to gofmt
  --tab                pass -tabindent=true to gofmt
  --tabwidth           pass -tabwidth to gofmt (default:4)
  --no-comments        pass -comments=false to gofmt
    `

    fmt.Println(helpMSG)
}

func printVersion() {
    fmt.Println("godag 0.1")
}

func printListing(){
    var listMSG string = `
  Listing of options and their content:

  -h --help            =>   %t
  -v --version         =>   %t
  -p --print           =>   %t
  -s --sort            =>   %t
  -o --output          =>   '%s'
  -S --static          =>   %t
  -a --arch            =>   %v
  -d --dryrun          =>   %t
  -c --clean           =>   %t
  -I                   =>   %v
  -t --test            =>   %t
  -b --benchmarks      =>   '%s'
  -m --match           =>   '%s'
  -V --verbose         =>   %t
  --test-bin           =>   '%s'
  -f --fmt             =>   %t
  --rew-rule           =>   '%s'
  --tab                =>   %t
  --tabwidth           =>   %s
  --no-comments        =>   %t

`;
    tabRepr := "4";
    if tabWidth != "" {
        tabRepr = tabWidth;
    }

    archRepr := "$GOARCH";
    if arch != "" {
        archRepr = arch;
    }

    fmt.Printf(listMSG, needsHelp, needsVersion, printInfo,
              sortInfo, output, static, archRepr, dryrun, cleanTree,
              includes, test, bmatch, match, testVerbose, gdtest,
              gofmt, rewRule, tabIndent, tabRepr, noComments);
}

