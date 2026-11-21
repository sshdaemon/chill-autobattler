package main

import "runtime"

// workers is how many battles run at once (-cpu). The default leaves half the
// machine free so a long run doesn't make it sluggish.
var workers = max(1, runtime.NumCPU()/2)
