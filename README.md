# coin
Pooled Bitcoin Mining  for charity

The idea of this project is to enrol a global community of Raspberry Pi (or similarly low power, hopefully largely idle) networked computers to participate in a distributed mining pool. Each such client will connect to a server which in turn connects to a single 'conductor'. The conductor is the external face of thsi mining operation and is the only machine that needs to operate a fill Bitcoin node. The servers monitor atached clients and distribute whatever block data they are given by the conductor. 
