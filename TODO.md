# Next steps

- Change the command map to be an expansion command map.

- Define an `ExpansionCommand` interface and redo the
    `func` based commands appropriately.
    Then make the map values `ExpansionCommand`.

- Make the macro type satisfy the interface -
    no need to cast it.
    
- Think about how to arrange the context object.
    Should probably have command sub-contexts,
    and logging ability.
    Can do this in a single definition
    by having `struct` within `struct`.
    
- Next, need to think adding new scopes during the
    expansion process through `{` and `}` symbols.
    It seems expansion can't be fully detached
    from things like parsing variable names and
    scoping?
    This will have repercussions for `\edef`.
