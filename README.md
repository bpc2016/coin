# coin
Pooled Bitcoin Mining for Charity

The idea of this project is to enrol a global community of Raspberry Pi (or similarly low power, hopefully largely idle) networked computers to participate in a distributed mining pool. 

Each such client will connect to a server which in turn connects to a single 'conductor'. The conductor is the external face of this mining operation and is the only machine that needs to operate a full Bitcoin node. 

The servers monitor attached clients and distribute whatever block data they are given by the conductor. 

All is written in Go and is an exercise on the one hand in a distributed synchronised system, and on the other in implementing the Bitcoin protocols required for mining.

Proceeds go to a charity because (a) that may be the only way to garner the required support and numbers and (b) it will likely involve far too many machines for serious sharing of proceeds :^)
