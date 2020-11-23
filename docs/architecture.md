# Architecture
loki has 3 main components that plugins need to implement in order to create and execute chaos scenarios.

1. **System**: Defines the ecosystem which knows how to load different resources in the system, how to validate when actual state and desired state are same etc.

2. **Destroyer**: Has the parsing logic for destroy section viz. exclusion and scenario.

3. **Killer**: Knows how to kill, delete etc. the resources of the system.

*Killer* and *Destroyer* together creates chaos scenarios and *System* validates whether software system has recovered from chaos scenario.

# Important terms

1. **Scenario**: is the collection of resources which will be killed/deleted etc. as a chaos test to check whether systems recover.

2. **Exclusion**: is the collection of resources which shouldn't be killed/deleted etc. as we expect system to not recover under such scenario.

3. **Ready Condition**: is the condition which determines when the systems have reached desired state and chaos tests can be executed.

# Design

<img src="https://github.com/narahari92/loki/raw/master/docs/architecture.png">

Plugins are developed by implementing interfaces `System`, `Destroyer` and `Killer` and then  registered  to loki. After that loki loads system, waits for ready condition to be satisfied and then identifies the state at the satisfaction of ready condition as desired state.

Loki then determines and executes chaos scenarios on system(s) and runs validate till system is recovered into desired state or times out.

Loki stops execution on first scenario failure.