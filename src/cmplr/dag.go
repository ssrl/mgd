// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package dag

import (
    "container/vector"
    "utilz/stringset"
    "utilz/stringbuffer"
    "go/parser"
    "path"
    "go/ast"
    "os"
    "fmt"
    "time"
    "log"
)


type Dag map[string]*Package // package-name -> Package object

type Package struct {
    indegree        int
    Name, ShortName string               // absolute path, basename
    Argv            []string             // command needed to compile package
    Files           *vector.StringVector // relative path of files
    dependencies    *stringset.StringSet
    children        *vector.Vector // packages that depend on this
}

type TestCollector struct {
    Names *vector.StringVector
}

func New() Dag {
    return make(map[string]*Package)
}

func newPackage() *Package {
    p := new(Package)
    p.indegree = 0
    p.Files = new(vector.StringVector)
    p.dependencies = stringset.New()
    p.children = new(vector.Vector)
    return p
}

func newTestCollector() *TestCollector {
    t := new(TestCollector)
    t.Names = new(vector.StringVector)
    return t
}


func (d Dag) Parse(root string, sv *vector.StringVector) {

    root = addSeparatorPath(root)

    var e string
    var max int = sv.Len()

    for i := 0; i < max; i++ {
        e = sv.At(i)
        tree := getSyntaxTreeOrDie(e, parser.ImportsOnly)
        dir, _ := path.Split(e)
        unroot := dir[len(root):len(dir)]
        shortname := tree.Name.String()
        pkgname := path.Join(unroot, shortname)

        _, ok := d[pkgname]
        if !ok {
            d[pkgname] = newPackage()
            d[pkgname].Name = pkgname
            d[pkgname].ShortName = shortname
        }

        ast.Walk(d[pkgname], tree)
        d[pkgname].Files.Push(e)
    }
}

func (d Dag) addEdge(from, to string) {
    fromNode := d[from]
    toNode := d[to]
    fromNode.children.Push(toNode)
    toNode.indegree++
}
// note that nothing is done in order to check if dependencies
// are valid if they are not part of the actual source-tree,
// i.e., stdlib dependencies and included (-I) dependencies
// are not investigated for validity..

func (d Dag) GraphBuilder(includes []string) {

    goRoot := path.Join(os.Getenv("GOROOT"), "src", "pkg")

    for k, v := range d {

        for dep := range v.dependencies.Iter() {

            if d.localDependency(dep) {
                d.addEdge(dep, k)
                /// fmt.Printf("local:  %s \n", dep);
            } else if !d.stdlibDependency(goRoot, dep) {
                if includes == nil || len(includes) == 0 {
                    log.Printf("[ERROR] Dependency: '%s' not found\n", dep)
                    log.Exit("[ERROR] Did you use actual src-root?\n")
                }
            }
        }
    }
}

func (d Dag) MakeDotGraph(filename string) {

    var rw_r__r__ uint32 = 420
    var file *os.File
    var fileinfo *os.FileInfo
    var e os.Error
    var sb *stringbuffer.StringBuffer

    fileinfo, e = os.Stat(filename)

    if e == nil {
        if fileinfo.IsRegular() {
            e = os.Remove(fileinfo.Name)
            if e != nil {
                log.Exitf("[ERROR] failed to remove: %s\n", filename)
            }
        }
    }

    sb = stringbuffer.NewSize(500)
    file, e = os.Open(filename, os.O_WRONLY|os.O_CREAT, rw_r__r__)

    if e != nil {
        log.Exitf("[ERROR] %s\n", e)
    }

    sb.Add("digraph depgraph {\n\trankdir=LR;\n")

    for _, v := range d {
        v.DotGraph(sb)
    }

    sb.Add("}\n")

    file.WriteString(sb.String())

    file.Close()

}

func (d Dag) MakeMainTest(root string) (*vector.Vector, string) {

    var max, i int
    var rwxr_xr_x uint32 = 493
    var isTest bool
    var sname, tmpdir, tmpstub, tmpfile string

    sbImports := stringbuffer.NewSize(300)
    sbTests := stringbuffer.NewSize(1000)
    sbBench := stringbuffer.NewSize(1000)

    sbImports.Add("\n// autogenerated code\n\n")
    sbImports.Add("package main\n\n")
    sbImports.Add("import \"regexp\";\n");
    sbImports.Add("import \"testing\";\n")


    sbTests.Add("\n\nvar tests = []testing.Test{\n")
    sbBench.Add("\n\nvar benchmarks = []testing.InternalBenchmark{\n")

    for _, v := range d {

        isTest = false
        sname = v.ShortName
        max = len(v.ShortName)

        if max > 5 && sname[max-5:] == "_test" {
            collector := newTestCollector()
            for i = 0; i < v.Files.Len(); i++ {
                tree := getSyntaxTreeOrDie(v.Files.At(i), 0)
                ast.Walk(collector, tree)
            }

            if collector.Names.Len() > 0 {
                isTest = true
                sbImports.Add(fmt.Sprintf("import \"%s\";\n", v.Name))
                for i = 0; i < collector.Names.Len(); i++ {
                    testFunc := collector.Names.At(i)
                    if len(testFunc) > 4 && testFunc[0:4] == "Test" {
                        sbTests.Add(fmt.Sprintf("testing.Test{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))
                    } else if len(testFunc) > 9 && testFunc[0:9] == "Benchmark" {
                        sbBench.Add(fmt.Sprintf("testing.InternalBenchmark{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))

                    }
                }
            }
        }

        if !isTest {

            collector := newTestCollector()

            for i = 0; i < v.Files.Len(); i++ {
                fname := v.Files.At(i)
                if len(fname) > 8 && fname[len(fname)-8:] == "_test.go" {
                    tree := getSyntaxTreeOrDie(fname, 0)
                    ast.Walk(collector, tree)
                }
            }

            if collector.Names.Len() > 0 {
                sbImports.Add(fmt.Sprintf("import \"%s\";\n", v.Name))
                for i = 0; i < collector.Names.Len(); i++ {
                    testFunc := collector.Names.At(i)
                    if len(testFunc) > 4 && testFunc[0:4] == "Test" {
                        sbTests.Add(fmt.Sprintf("testing.Test{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))
                    } else if len(testFunc) > 9 && testFunc[0:9] == "Benchmark" {
                        sbBench.Add(fmt.Sprintf("testing.InternalBenchmark{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))
                    }
                }
            }
        }
    }

    sbTests.Add("};\n")
    sbBench.Add("};\n\n")

    sbTotal := stringbuffer.NewSize(sbImports.Len() +
        sbTests.Len() +
        sbBench.Len() + 5)
    sbTotal.Add(sbImports.String())
    sbTotal.Add(sbTests.String())
    sbTotal.Add(sbBench.String())

    sbTotal.Add("func main(){\n")
    sbTotal.Add("testing.Main(regexp.MatchString, tests);\n")
    sbTotal.Add("testing.RunBenchmarks(regexp.MatchString, benchmarks);\n}\n\n")

    tmpstub = fmt.Sprintf("tmp%d", time.Seconds())
    tmpdir = fmt.Sprintf("%s%s", addSeparatorPath(root), tmpstub)

    dir, e1 := os.Stat(tmpdir)

    if e1 == nil && dir.IsDirectory() {
        log.Printf("[ERROR] directory: %s already exists\n", tmpdir)
    } else {
        e_mk := os.Mkdir(tmpdir, rwxr_xr_x)
        if e_mk != nil {
            log.Exit("[ERROR] failed to create directory for testing")
        }
    }

    tmpfile = path.Join(tmpdir, "main.go")

    fil, e2 := os.Open(tmpfile, os.O_WRONLY|os.O_CREAT, rwxr_xr_x)

    if e2 != nil {
        log.Exitf("[ERROR] %s\n", e2)
    }

    n, e3 := fil.WriteString(sbTotal.String())

    if e3 != nil {
        log.Exitf("[ERROR] %s\n", e3)
    } else if n != sbTotal.Len() {
        log.Exit("[ERROR] failed to write test")
    }

    fil.Close()

    p := newPackage()
    p.Name = path.Join(tmpstub, "main")
    p.ShortName = "main"
    p.Files.Push(tmpfile)

    vec := new(vector.Vector)
    vec.Push(p)
    return vec, tmpdir
}

func (d Dag) Topsort() *vector.Vector {

    var node, child *Package
    var cnt int = 0

    zero := new(vector.Vector)
    done := new(vector.Vector)

    for _, v := range d {
        if v.indegree == 0 {
            zero.Push(v)
        }
    }

    for zero.Len() > 0 {

        node, _ = zero.Pop().(*Package)

        for i := 0; i < node.children.Len(); i++ {
            child = node.children.At(i).(*Package)
            child.indegree--
            if child.indegree == 0 {
                zero.Push(child)
            }
        }
        cnt++
        done.Push(node)
    }

    if cnt < len(d) {
        log.Exit("[ERROR] loop in dependency graph")
    }

    return done
}

func (d Dag) localDependency(dep string) bool {
    _, ok := d[dep]
    return ok
}

func (d Dag) stdlibDependency(root, dep string) bool {
    dir, staterr := os.Stat(path.Join(root, dep))
    if staterr != nil {
        return false
    }
    return dir.IsDirectory()
}

func (d Dag) PrintInfo() {

    var i int

    fmt.Println("--------------------------------------")
    fmt.Println("Packages and Dependencies")
    fmt.Println("p = package, f = file, d = dependency ")
    fmt.Println("--------------------------------------\n")

    for k, v := range d {
        fmt.Println("p ", k)
        for i = 0; i < v.Files.Len(); i++ {
            fmt.Println("f ", v.Files.At(i))
        }
        for ds := range v.dependencies.Iter() {
            fmt.Println("d ", ds)
        }
        fmt.Println("")
    }
}

func (p *Package) DotGraph(sb *stringbuffer.StringBuffer) {

    if p.dependencies.Len() == 0 {

        sb.Add(fmt.Sprintf("\t\"%s\";\n", p.Name))

    } else {

        for dep := range p.dependencies.Iter() {
            sb.Add(fmt.Sprintf("\t\"%s\" -> \"%s\";\n", p.Name, dep))
        }
    }
}


func (p *Package) UpToDate() bool {

    if p.Argv == nil {
        panic("Missing dag.Package.Argv")
    }

    var e os.Error
    var finfo *os.FileInfo
    var compiledModifiedTime int64
    var last, stop, i int
    var resultingFile string

    last = len(p.Argv) - 1
    resultingFile = p.Argv[last-p.Files.Len()]
    stop = last - p.Files.Len()

    finfo, e = os.Stat(resultingFile)

    if e != nil {
        return false
    } else {
        compiledModifiedTime = finfo.Mtime_ns
    }

    for i = last; i > stop; i-- {
        finfo, e = os.Stat(p.Argv[i])
        if e != nil {
            panic(fmt.Sprintf("Missing go file: %s\n", p.Argv[i]))
        } else {
            if finfo.Mtime_ns > compiledModifiedTime {
                return false
            }
        }
    }

    return true
}

func (p *Package) Ready(local, compiled *stringset.StringSet) bool {

    for dep := range p.dependencies.Iter() {
        if local.Contains(dep) && !compiled.Contains(dep) {
            return false
        }
    }

    return true
}

func (p *Package) Visit(node interface{}) (v ast.Visitor) {

    switch node.(type) {
    case *ast.BasicLit:
        bl, ok := node.(*ast.BasicLit)
        if ok {
            stripped := stripQuotes(string(bl.Value))
            p.dependencies.Add(stripped)
        }
    default: // nothing to do if not BasicLit
    }
    return p
}

func (t *TestCollector) Visit(node interface{}) (v ast.Visitor) {
    switch node.(type) {
    case *ast.FuncDecl:
        fdecl, ok := node.(*ast.FuncDecl)
        if ok {
            t.Names.Push(fdecl.Name.Name)
        }
    default: // nothing to do if not FuncDecl
    }
    return t
}

func stripQuotes(s string) string {
    stripped := s[1:(len(s) - 1)]
    return stripped
}

func addSeparatorPath(root string) string {
    if root[len(root)-1:] != "/" {
        root = root + "/"
    }
    return root
}

func getSyntaxTreeOrDie(file string, mode uint) *ast.File {
    absSynTree, err := parser.ParseFile(file, nil, mode)
    if err != nil {
        log.Exitf("%s\n", err)
    }
    return absSynTree
}
