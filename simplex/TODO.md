# Interesting bits wrt GUI

https://www.stat.washington.edu/research/reports/1991/tr201.pdf

# Glop for GUI

https://groups.google.com/forum/#!msg/or-tools-discuss/SZ-z5gE07fs/YFNOT1ZASDMJ

Glop was designed with operations research applications in mind. It does not implement the same primitives that Cassowary does, but in theory you could use Glop to implement the same primitives as Cassowary. I would be more work, though

Like Cassowary, Glop implements a dual simplex (accessible by setting the parameter use_dual_simplex=true). A dual simplex makes it possible to very efficiently add constraintwould be in s to an already solved problem, with minimal recomputation. As for retracting constraints, the bounds of the removed constraints need to be changed to -infinity/+infinity.

Glop implements an optimized version of the dual revised simplex. I'm not sure about Cassowary, but from their literature https://constraints.cs.washington.edu/solvers/cassowary-tochi.pdf Cassowary seems to implement the simpler and slower tableau version of the simplex. Moreover they handle unrestricted variables (with no upper- or lower-bounds) by using two simplex tableaux.

The main issue for using Glop in UI applications would most likely be the handling of constraint hierarchies. One suggestion would be to make the constraint optional by transforming a.x <= n into a.x + s = b with s>= 0, and to add a cost for s in the cost function. The cost would be all the higher as the priority of the constraint would be high.

# Handling unrestricted variables

https://youtu.be/sNjyHyTLc44?t=2399
