// © Knug Industries 2009 all rights reserved 
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package dag

import(
    "container/vector";
    "utilz/stringset";
    "go/parser";
    "path";
    "go/ast";
    "os";
    "fmt";
)


type Dag struct{
    pkgs map[string] *Package; // package-name -> Package object
}

type Package struct{
    indegree int;
    Name, ShortName string;
    Files *vector.StringVector; // relative path of files
    dependencies *stringset.StringSet;
    children *vector.Vector; // packages that depend on this
}

func New() *Dag{
    d := new(Dag);
    d.pkgs = make(map[string]*Package);
    return d;
}

func newPackage() *Package{
    p := new(Package);
    p.indegree = 0;
    p.Files = new(vector.StringVector);
    p.dependencies = stringset.New();
    p.children = new(vector.Vector);
    return p;
}

func (d *Dag) Parse(root string, sv *vector.StringVector){

    if root[len(root)-1:len(root)] != "/" {
        root = root + "/";
    }

    for e := range sv.Iter() {

        tree := getSyntaxTreeOrDie(e);
        dir, _ := path.Split(e);
        unroot := dir[len(root):len(dir)];
        pkgname := path.Join(unroot, tree.Name.Obj.Name);

        _, ok := d.pkgs[pkgname];
        if ! ok {
            d.pkgs[pkgname] = newPackage();
            d.pkgs[pkgname].Name = pkgname;
            d.pkgs[pkgname].ShortName = tree.Name.Obj.Name;
        }

        ast.Walk( d.pkgs[pkgname], tree );
        d.pkgs[pkgname].Files.Push(e);
    }
}

func (d *Dag) addEdge(from, to string){
    fromNode := d.pkgs[from];
    toNode   := d.pkgs[to];
    fromNode.children.Push(toNode);
    toNode.indegree++;
}

func (d *Dag) GraphBuilder(){

    goRoot := path.Join(os.Getenv("GOROOT"), "src/pkg");

    for k,v := range d.pkgs {

        for dep := range v.dependencies.Iter() {

            if d.localDependency(dep) {
                d.addEdge(dep, k);
            }else if ! d.stdlibDependency(goRoot, dep) {
                fmt.Fprintf(os.Stderr,"[ERROR] Dependency: %s not found\n",dep);
                fmt.Fprintf(os.Stderr,"[ERROR] Did you use actual src-root?\n");
                os.Exit(1);
            }
        }
    }
}

func (d *Dag) Topsort() *vector.Vector{

    var node,child *Package;
    var cnt int = 0;

    zero := new(vector.Vector);
    done := new(vector.Vector);

    for _,v := range d.pkgs {
        if v.indegree == 0 {
            zero.Push(v);
        }
    }

    for zero.Len() > 0 {

        node,_ = zero.Pop().(*Package);

        for ch := range node.children.Iter() {
            child = ch.(*Package);
            child.indegree--;
            if child.indegree == 0 {
                zero.Push(child);
            }
        }
        cnt++;
        done.Push(node);
    }

    if cnt < len(d.pkgs) {
        fmt.Fprintf(os.Stderr,"[ERROR] loop in dependency graph\n");
        os.Exit(1);
    }

    return done;
}

func (d *Dag) localDependency(dep string) bool{
    _, ok := d.pkgs[dep];
    return ok;
}

func (d *Dag) stdlibDependency(root, dep string) bool{
    dir, staterr := os.Stat(path.Join(root, dep));
    if staterr != nil { return false; }
    return dir.IsDirectory();
}

func (d *Dag) PrintInfo(){

    fmt.Println("--------------------------------------");
    fmt.Println("Packages and Dependencies");
    fmt.Println("p = package, f = file, d = dependency ");
    fmt.Println("--------------------------------------\n");

    for k,v := range d.pkgs {
        fmt.Println("p ",k);
        for fs := range v.Files.Iter() {
            fmt.Println("f ",fs);
        }
        for ds := range v.dependencies.Iter() {
            fmt.Println("d ",ds);
        }
        fmt.Println("");
    }
}


func (p *Package) Visit(node interface{}) (v ast.Visitor){

    switch t := node.(type){
        case *ast.BasicLit:
            bl, ok := node.(*ast.BasicLit);
            if ok{
                stripped := stripQuotes(string(bl.Value));
                p.dependencies.Add(stripped);
            }
        default: // nothing to do if not BasicLit
    }
    return p;
}

func stripQuotes(s string) string{
    stripped := s[1:(len(s) -1)];
    return stripped;
}


func getSyntaxTreeOrDie(file string) (*ast.File){
    absSynTree, err := parser.ParseFile(file, nil, nil, parser.ImportsOnly);
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err);
        os.Exit(1);
    }
    return absSynTree;
}
