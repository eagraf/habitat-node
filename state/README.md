# Critical State Consensus
This module implements Replicated State Machines that agree on sequences of operations for important state about 
communities. Like all things in Habitat, the conensus algorithm can be chosen from a variety of implementations.

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

