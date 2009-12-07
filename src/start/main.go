// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package main

import(
    "os";
	"fmt";
	"path";
	"utilz/walker";
    "cmplr/compiler";
    "cmplr/dag";
	"container/vector";
	"strings";
    "flag";
)

var helpflag, versionflag, tprint, dbprint bool;
var arch, output string;

func init(){

    flag.Usage = printHelp;
    flag.BoolVar(&helpflag, "h", false, "print help");
    flag.BoolVar(&helpflag, "help", false, "print help");
    flag.BoolVar(&versionflag, "v", false, "print version");
    flag.BoolVar(&versionflag, "version", false, "print version");
    flag.BoolVar(&tprint, "s", false, "print legal compile order");
    flag.BoolVar(&tprint, "sort", false, "print legal compile order");
    flag.BoolVar(&dbprint,"p", false, "print collected info");
    flag.BoolVar(&dbprint,"print", false, "print collected info");
    flag.StringVar(&arch,"a", "", "architecture");
    flag.StringVar(&arch,"arch", "", "architecture");
    flag.StringVar(&output,"o", "", "output");
    flag.StringVar(&output,"output", "", "output");

    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = func(s string)bool{
        return strings.HasSuffix(s,".go");
    };

    // override IncludeDir to make walker ignore 'hidden' directories
    walker.IncludeDir = func(s string)bool{
        _, dirname := path.Split(s);
        return dirname[0] != '.';
    };

    if os.Getenv("GOROOT") == "" {
        fmt.Fprintf(os.Stderr,"[ERROR] missing GOROOT\n");
        os.Exit(1);
    }
}


func main(){

    var files *vector.StringVector;

    flag.Parse();

    if helpflag    { printHelp(); os.Exit(0); }
    if versionflag { printVersion(); os.Exit(0); }

    for i := 0; i < flag.NArg() ; i++ {

        files = findFiles(flag.Arg(i));

        dgrph := dag.New();
        dgrph.Parse(flag.Arg(i), files);

        if dbprint {
            dgrph.PrintInfo();
            os.Exit(0);
        }

        dgrph.GraphBuilder();
        sorted := dgrph.Topsort();

        if tprint {
            for pkg := range sorted.Iter() {
                rpkg, _ := pkg.(*dag.Package);
                fmt.Printf("%s\n", rpkg.Name);
            }
            os.Exit(0);
        }

        cmplr  := compiler.New(flag.Arg(i), arch);
        cmplr.ForkCompile(sorted);

        if output != "" {
            cmplr.ForkLink(sorted, output);
        }
    }
}

func findFiles(pathname string) *vector.StringVector{

    var dir *os.Dir;
    var staterr  os.Error;
    var files *vector.StringVector;

    dir, staterr = os.Stat(pathname);

    if staterr != nil {
        fmt.Fprintf(os.Stderr,"[ERROR] %s\n", staterr);
        os.Exit(1);
    }else if ! dir.IsDirectory() {
        fmt.Fprintf(os.Stderr,"[ERROR] %s: is not a directory\n",
                    pathname);
        os.Exit(1);
    }else{
        files = walker.PathWalk(path.Clean(pathname));
    }

    return files;
}

func printHelp(){
    var helpMSG string =`
  usage: godag [OPTIONS] src-directory

  options:

  -h --help        print this message and quit
  -v --version     print version and quit
  -p --print       print package info collected
  -s --sort        print legal compile order
  -o --output      link to produce program
  -a --arch        architecture (arm64,arm,386)
    `;

    fmt.Println(helpMSG);
}

func printVersion(){
    fmt.Println("godag 0.1");
}
