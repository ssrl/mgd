// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package compiler

import(
    "os";
    "container/vector";
    "strings";
    "fmt";
    "cmplr/dag";
    "path";
)



type Compiler struct{
    root, arch, suffix, executable string;
}

func New(root, arch string) *Compiler{
    c      := new(Compiler);
    c.root  = root;
    c.arch, c.suffix = archNsuffix(arch);
    c.executable     = findCompiler(c.arch);
    return c;
}

func findCompiler(arch string) string{

    var lookingFor string;
    switch arch {
        case "arm"  : lookingFor = "5g";
        case "arm64": lookingFor = "6g";
        case "386"  : lookingFor = "8g";
    }

    real := which(lookingFor);
    if real == "" {
        fmt.Fprintf(os.Stderr,"[ERROR] could not find compiler\n");
        os.Exit(1);
    }
    return real;
}

func findLinker(arch string) string{

    var lookingFor string;
    switch arch {
        case "arm"  : lookingFor = "5l";
        case "arm64": lookingFor = "6l";
        case "386"  : lookingFor = "8l";
    }

    real := which(lookingFor);
    if real == "" {
        fmt.Fprintf(os.Stderr,"[ERROR] could not find linker\n");
        os.Exit(1);
    }
    return real;
}


func archNsuffix(arch string)(a, s string){

    if arch == "" {
        a = os.Getenv("GOARCH");
    }

    switch a {
        case "arm"  : s = ".5";
        case "arm64": s = ".6";
        case "386"  : s = ".8";
        default:
            fmt.Fprintf(os.Stderr,"[ERROR] unknown architecture: %s\n",a);
            os.Exit(1);
    }

    return a, s;
}

func (c *Compiler) String() string{
    s := "Compiler{ root=%s, arch=%s, suffix=%s, executable=%s }";
    return fmt.Sprintf(s, c.root, c.arch, c.suffix, c.executable);
}

func (c *Compiler) ForkCompile(pkgs *vector.Vector){

    for p := range pkgs.Iter() {
        pkg, _ := p.(*dag.Package);//safe cast, only Packages there

        argv := make([]string, 5 + pkg.Files.Len());
        i    := 0;
        argv[i] = c.executable; i++;
        argv[i] = "-I"; i++;
        argv[i] = c.root; i++;
        argv[i] = "-o"; i++;
        argv[i] = path.Join(c.root, pkg.Name) + c.suffix; i++;

        for f := range pkg.Files.Iter() {
            argv[i] = f;
            i++;
        }

        fmt.Println("compiling:",pkg.Name);

        stdExecute(argv);

    }
}

func stdExecute(argv []string){

    var fdesc []*os.File;

    fdesc = make([]*os.File, 3);
    fdesc[0] = os.Stdin;
    fdesc[1] = os.Stdout;
    fdesc[2] = os.Stderr;

    pid, err := os.ForkExec(argv[0], argv, os.Environ(), "", fdesc);

    if err != nil{
        fmt.Fprintf(os.Stderr, "[ERROR] %s\n",err);
        os.Exit(1);
    }

    wmsg, werr := os.Wait(pid, 0);

    if werr != nil || wmsg.WaitStatus != 0 {
        os.Exit(1);
    }

}

func (c *Compiler) Link(pkgs *vector.Vector, output string){

    gotMain := new(vector.Vector);

    for p := range pkgs.Iter() {
        pk, _ := p.(*dag.Package);
        if pk.ShortName == "main" {
            gotMain.Push( pk );
        }
    }

    if gotMain.Len() == 0 {
        fmt.Fprintf(os.Stderr,"[ERROR] (linking) no main package found\n");
        os.Exit(1);
    }

    if gotMain.Len() > 1 {
        fmt.Fprintf(os.Stderr,"[ERROR] (linking) more than one main package found\n");
        os.Exit(1);
    }

    pkg, _ := gotMain.Pop().(*dag.Package);

    linker := findLinker(c.arch);
    compiled := path.Join(c.root, pkg.Name) + c.suffix;

    argv := make([]string, 4);
    i    := 0;
    argv[i] = linker; i++;
    argv[i] = "-o"; i++;
    argv[i] = output; i++;
    argv[i] = compiled; i++;

    fmt.Println("linking  :",output);

    stdExecute(argv);

}

func which(cmd string) (string){

    var abspath string;
    var dir *os.Dir;
    var err os.Error;

    xpath := os.Getenv("PATH");
    dirs  := strings.Split(xpath, ":", 0);

    for i := range dirs {
        abspath = path.Join(dirs[i], cmd);
        dir, err = os.Stat(abspath);
        if err == nil{
            if dir.IsRegular(){
                if isExecutable(dir.Uid, dir.Permission()) {
                    return abspath;
                }
            }
        }
    }

    return "";
}

func isExecutable(uid uint32, perms int) bool {

    mode := 7;
    amode := (perms & mode);
    mode = mode << 6;
    umode := (perms & mode) >> 6;


    if amode == 7 || amode == 5 {
        return true;
    }

    if int(uid) == os.Getuid() {
        if umode == 7 || umode == 5 {
            return true;
        }
    }

    return false;
}
