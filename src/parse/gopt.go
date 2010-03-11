// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package gopt

/*

Not too happy with the flag package provided
by the go-team, so this is another version.

Most notable difference:

  - Multiple 'option-strings' for a single option ('-r -rec -R')
  - Non-option arguments can come anywhere in argv
  - Option arguments can be in juxtaposition with flag
  - Only two types of options: string, bool


Usage:

 getopt := gopt.New();

 getopt.BoolOption("-h -help --help");
 getopt.BoolOption("-v -version --version");
 getopt.StringOption("-f -file --file --file=");
 getopt.StringOption("-l --list");
 getopt.StringOption("-I");

 args := getopt.Parse(os.Args[1:]);

 // getopt.IsSet("-h") == getopt.IsSet("-help") ..

 if getopt.IsSet("-help"){ println("-help"); }
 if getopt.IsSet("-v")   { println("-version"); }
 if getopt.IsSet("-file"){ println("--file ",getopt.Get("-f")); }
 if getopt.IsSet("-list"){ println("--list ",getopt.Get("-list"); }

 if getopt.IsSet("-I"){
     elms := getopt.GetMultiple("-I");
     for y := range elms { println("-I ",elms[y]);  }
 }

 for i := range args{
     println("remaining:",args[i]);
 }


*/


import(
    "strings";
    "container/vector";
    "fmt";
    "os";
)

type GetOpt struct{
    options *vector.Vector;
    cache map[string]Option;
}

func New() *GetOpt{
    g := new(GetOpt);
    g.options = new(vector.Vector);
    g.cache   = make(map[string]Option);
    return g;
}

func (g *GetOpt) isOption(o string) (Option) {
    _, ok := g.cache[o];
    if ok {
        element, _ := g.cache[o];
        return element;
    }
    return nil;
}

func stop(msg string, format ...interface{}){
    fmt.Fprintf(os.Stderr, msg, format);
    os.Exit(1);
}

func err(msg string, format ...interface{}){
    fmt.Fprintf(os.Stderr, msg, format);
}

func (g *GetOpt) getStringOption(o string) *StringOption{

    opt := g.isOption(o);

    if opt != nil {
        sopt, ok := opt.(*StringOption);
        if ok{
            return sopt;
        }else{
            stop("%s: is not a string option\n", o);
        }
    }else{
        stop("%s: is not an option at all\n", o);
    }

    return nil;
}

func (g *GetOpt) Get(o string) string{

    sopt := g.getStringOption(o);

    switch sopt.count {
        case 0 : stop("%s: is not set\n", o);
        case 1 : // fine do nothing
        default: err("[warning] option %s: has more arguments than 1\n", o);
    }
    return sopt.values[0];
}

func (g *GetOpt) GetMultiple(o string) []string{

    sopt := g.getStringOption(o);

    if sopt.count == 0{
        stop("%s: is not set\n", o);
    }

    return sopt.values[0:sopt.count];
}

func (g *GetOpt) Parse(argv []string) (args []string){

    var count int = 0;
    // args cannot be longer than argv, if no options
    // are given on the command line it is argv
    args = make([]string, len(argv));

    for i := 0; i < len(argv); i++ {

        opt := g.isOption(argv[i])

        if opt != nil {

            switch opt.(type) {
                case *BoolOption:
                    bopt, _ := opt.(*BoolOption);
                    bopt.setFlag();
                case *StringOption:
                    sopt, _ := opt.(*StringOption);
                    if i + 1 >= len(argv){
                        stop("missing argument for: %s\n",argv[i]);
                    }else{
                        sopt.addArgument(argv[i+1]);
                        i++;
                    }
            }

        }else{

            // arguments written next to options
            start , ok := g.juxtaOption(argv[i]);

            if ok {
                stropt := g.getStringOption(start);
                stropt.addArgument(argv[i][len(start):]);
            }else{
                args[count] = argv[i];
                count++;
            }
        }
    }

    return args[0:count];
}

func (g *GetOpt) juxtaOption(opt string)(string, bool){

    var tmpmax string = "";

    for o := range g.options.Iter() {

        sopt, ok := o.(*StringOption);

        if ok {
            s := sopt.startsWith(opt);
            if s != "" {
                if len(s) > len(tmpmax){
                    tmpmax = s;
                }
            }
        }
    }

    if tmpmax != ""{
        return tmpmax, true;
    }

    return "", false;
}

func (g *GetOpt) IsSet(o string) bool{
    _, ok := g.cache[o];
    if ok {
        element, _ := g.cache[o];
        return element.isSet();
    }else{
        stop("%s not an option\n", o);
    }
    return false;
}

func (g *GetOpt) BoolOption(optstr string){
    ops := strings.Split(optstr, " ", -1);
    boolopt := newBoolOption(ops);
    for i := range ops {
        g.cache[ops[i]] = boolopt;
    }
    g.options.Push(boolopt);
}

func (g *GetOpt) StringOption(optstr string){
    ops := strings.Split(optstr, " ", -1);
    stringopt := newStringOption(ops);
    for i := range ops {
        g.cache[ops[i]] = stringopt;
    }
    g.options.Push(stringopt);
}