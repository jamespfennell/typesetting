# GoTeX: a new and modern implementation of TeX

This repository hosts a side project whose aim to produce
    a new implementation of the TeX typesetting system, provisionally called GoTeX.
The goal is for GoTeX to be a full drop-in replacement for the standard (or "legacy") TeX implementation,
    as well as subsequent extensions of TeX such as pdfTeX and LuaTeX.
Note that LaTeX support is automatic: because LaTeX is a set of macros built on top of the TeX language,
    it works with any correct implementation of TeX, including GoTeX.

The project is currently in a very early stage.
A full implementation of the Knuth-Plass line-breaking algorithm
    has already been written (see package `knuthplass`).
Current development work is focussed on the pure language aspects of TeX - 
    implementing expansion primitives such as conditionals (package `tex/commands/conditional`) 
    and macros (package `tex/commands/macro`).
    
## Why create a new implementation of TeX?

There are a few compelling reasons to create a new implementation of TeX,
    but here I'll only discuss the main one.
In a nutshell: the TeX source code is extremely challenging to work with,
    and this directly holds back progress and innovation in the algorithmic typesetting space.

The legacy implementation of TeX was started around 4 decades ago 
    and from a software design perspective it's showing its age.
It's implemented in a custom dialect of Pascal called WEB which was created specifically by Donald Knuth
    for writing TeX.
This language is a problem
    (few people know it; compared to C++ or Java, it lacks important features for large-scale software development).
The TeX source itself contains what are now considered basic design anti-patterns,
    chief among them
    the prevalence of confusing mutable global state,
    and an almost total lack of encapsulation.
It's very hard to reason about the program and to modify it in any way.

...to be continued






