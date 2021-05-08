# Critical State Consensus
This module implements Replicated State Machines that agree on sequences of operations for important state about communities. Like all things in Habitat, the conensus algorithm can be chosen from a variety of implementations.

## State Update Process
In normal execution, the following sequence takes place after a new operation is committed by the Replicated State Machine.

1. Next operation is agreed upon by consensus algorithm
2. The operation is written to disk in a Write-Ahead-Log
3. The new operation's reducer is applied to the current state, so that a new state is produced
    1. After `n` transitions, a snapshot of the current state is taken and written to disk
4. The state monitor receives the transition and updates all relevant TransitionSubscribers of the new operation

## Restart Process

1. The processes state is restored to the last snapshot
2. Any operations in the Write-Ahead-Log with a higher sequence number than the snapshot are applied to the state
3. The consensus module for each community is restarted, and asks for all updates that may have been missed
4. When updates are received, they are applied to the log
5. Continue operation as normal

## Write-Ahead-Log and Snapshoting

Each entry in a WAL file takes up one line, making it easy to append to the log file. Each log line has two components, the sequence number in human readable decimals, and then a base64 encoded entry that includes the transition data, sequence number, and timestmamp.

As transitions are committed by the consensus mechanism, periodic snapshots of the current state are taken. When a new snapshot is taken, the log file is renamed, with the first sequence number appended to the file name. A new log file is created, that will contain all subsequent logs until the next snapshot. In the event that the state machine has to recover from a crash, it can reinitialize state from the snapshot and then roll up the remaining transitions from the current log file.

