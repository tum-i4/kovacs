# Purpose

If a dispute should arise, this tool can, in most cases, determine the guilty party by examining the non-repudiation logs.

# Example

## Check if exchange ended successfully

Replace ${rep} with the file name of the desired non-repudiation protocol ```./verifier -checkSuccess ../listener/storage/${rep}.json```

## Solve a dispute

```./verifier -isDispute path/to/non-rep/{non-rep1}.json path/to/non-rep/{non-rep2}.json```
