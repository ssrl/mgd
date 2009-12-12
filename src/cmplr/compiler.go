// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package compiler

import(
    "os";
    "container/vector";
    "fmt";
    "utilz/handy";
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

    real := handy.Which(lookingFor);
    if real == "" {
        die("[ERROR] could not find compiler\n");
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

    real := handy.Which(lookingFor);
    if real == "" {
        die("[ERROR] could not find linker\n");
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
        default     : die("[ERROR] unknown architecture: %s\n",a);
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

        handy.StdExecve(argv);

    }
}

func (c *Compiler) ForkLink(pkgs *vector.Vector, output string){

    gotMain := new(vector.Vector);

    for p := range pkgs.Iter() {
        pk, _ := p.(*dag.Package);
        if pk.ShortName == "main" {
            gotMain.Push( pk );
        }
    }

    if gotMain.Len() == 0 {
        die("[ERROR] (linking) no main package found\n");
    }

    if gotMain.Len() > 1 {
        die("[ERROR] (linking) more than one main package found\n");
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

    handy.StdExecve(argv);

}

func die(strfmt string, v ...){
    fmt.Fprintf(os.Stderr, strfmt, v);
    os.Exit(1);
}
