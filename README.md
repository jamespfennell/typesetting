# Typesetting repo

This work-in-progress repo hosts a collection of interelated Go packages
related to typesetting.

## Complete packages

- `github.com/jamespfennell/typesetting/pkg/knuthplass` - a full
    implementation of the Knuth-Plass line-breaking algorithm.
    
## Roadmap

The ideal end state of this repository is 
a new and modern implementation of Tex in Go,
provisionally called GoTex by analogy to CPython and Jython.
This ideal end state is very ambitious and is likely not to be attained.
However, the initial roadmap is:

1. Write code to read OTF or TTF font files and extract font metrics from them,
    using preexisting Go libraries. 
    This code should be written in such a way as to completely abstract away
    the font format.
    We want to avoid the pitfall of Tex, where the implementation details
    of a specific font file format are hard coded in the source.
    
1. Build a "driver" to glue together this font metrics data
    and the Knuth-Plass algorithm.
    It will need to include code for generating Knuth-Plass primitives 
    from text given different alignments (ragged right, justified, etc.).
    
1. Write code to export to PDF using preexisting Go libraries. 
    Again, will need to hide implementation details of PDF as much as possible.

1. Use these building blocks to develop a full Markdown to PDF tool.

1. Starting working on a Tex language parser...