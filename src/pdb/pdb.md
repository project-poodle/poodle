
- Log commits


                                   +--- reader 1
                                   +--- reader 2
                 +---> table 1 <---+--- reader 3
                 |                 +--- ...
                 +---> table 2     +--- reader n
                 |
    pdb log ---> +---> table 3     ...
                 |
                 +---> ...         +--- reader 1
                 |                 +--- reader 2
                 +---> table n <---+--- reader 3
                                   +--- ...
                                   +--- reader n
