
About:
------------------------------------------------------------
This program will hopefully make it easier to compile 
source code in the go programming language.
A dependency graph is constructed from imports, this
is sorted with a topological sort to figure out legal
compile order.


Build:
------------------------------------------------------------

This should be as easy as running the script ./build.sh


Install:
------------------------------------------------------------

Copy the file: gd  somewhere it can be found ($PATH)


Try it Out:
------------------------------------------------------------

You can try to compile the same source using the generated
executable: gd


$ ./gd src          # will compile source inside src
$ ./gd -p src       # will print dependency info gathered
$ ./gd -s src       # will print legal compile order
$ ./gd -o name src  # will produce executable 'name' of
                    # source-code inside src directory
$ ./gd src -test    # will run unit-tests



Philosophy (Babble?)
------------------------------------------------------------

Without a tool to figure out which order the source should
be compiled, Makefiles are usually the result. Makefiles
are static in nature, which make them a poor choice to handle
a dynamic problem like a changing source tree. They also make
flat structures quite common, since this usually simplifies
the Makefiles, but makes organisation far less intuitive than
a directory-tree package-structure.


Logo
------------------------------------------------------------

The logo was made with LaTeX and tikz, so I'll include it
here, if anyone is interested in creating their own :-)

=start LaTeX

\documentclass[12pt]{article}
\usepackage{tikz}
\usepackage{nopageno}


\begin{document}
\begin{tikzpicture}[remember picture,overlay]
  \node [scale=75,fill=black!100,opacity=.8, rounded corners]
   at (current page.center) {}; 
  \node [rotate=180,scale=63,text opacity=0.9,yellow]
   at (current page.center) {g};
\end{tikzpicture}
\end{document}


=end LaTeX




-bjarneh
