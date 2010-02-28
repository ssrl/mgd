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
    "parse/gopt";
	"strings";
)


func init(){

    // override IncludeFile to make walker pick up only .go files
    walker.IncludeFile = func(s string)bool{
        return strings.HasSuffix(s,".go");
    };

    // override IncludeDir to make walker ignore 'hidden' directories
    walker.IncludeDir = func(s string)bool{
        _, dirname := path.Split(s);
        return dirname[0] != '.';
    };

}

func gotRoot(){
    if os.Getenv("GOROOT") == "" {
        fmt.Fprintf(os.Stderr,"[ERROR] missing GOROOT\n");
        os.Exit(1);
    }
}


func main(){

    var files *vector.StringVector;

    var arch, output string;
    var dryrun bool;

    getopt := gopt.New();

    getopt.BoolOption("-h -help --help help");
    getopt.BoolOption("-v -version --version version");
    getopt.BoolOption("-s -sort --sort sort");
    getopt.BoolOption("-p -print --print");
    getopt.BoolOption("-d -dryrun --dryrun");
    getopt.StringOption("-a -arch --arch -arch= --arch=");
    getopt.StringOption("-o -output --output -output= --output=");

    args := getopt.Parse(os.Args[1:]);

    if getopt.IsSet("-help") { printHelp(); os.Exit(0); }
    if getopt.IsSet("-version") { printVersion(); os.Exit(0); }
    if getopt.IsSet("-dryrun"){ dryrun = true; }

    gotRoot();//?

    if getopt.IsSet("-arch"){ arch = getopt.Get("-a"); }
    if getopt.IsSet("-output"){ output = getopt.Get("-o"); }

    for i := 0; i < len(args) ; i++ {

        files = findFiles(args[i]);

        dgrph := dag.New();
        dgrph.Parse(args[i], files);

        if getopt.IsSet("-print") {
            dgrph.PrintInfo();
            os.Exit(0);
        }

        dgrph.GraphBuilder();
        sorted := dgrph.Topsort();

        if getopt.IsSet("-sort") {
            for pkg := range sorted.Iter() {
                rpkg, _ := pkg.(*dag.Package);
                fmt.Printf("%s\n", rpkg.Name);
            }
            os.Exit(0);
        }

        cmplr  := compiler.New(args[i], arch, dryrun);
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
  usage: gd [OPTIONS] src-directory

  options:

  -h --help        print this message and quit
  -v --version     print version and quit
  -p --print       print package info collected
  -s --sort        print legal compile order
  -o --output      link to produce program
  -a --arch        architecture (amd64,arm,386)
  -d --dryrun      print what gd would do (stdout)
    `;

    fmt.Println(helpMSG);
}

func printVersion(){
    fmt.Println("godag 0.1");
}
