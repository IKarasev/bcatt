package agregator

type NodesAttacks struct {
	history map[string]AttackInTick
}

type AttackInTick map[int]int

func NewNodesAttacks() *NodesAttacks {
	return &NodesAttacks{
		history: make(map[string]AttackInTick),
	}
}

func (na *NodesAttacks) AddAttack(nodeId string, tick int) {
	if _, ok := na.history[nodeId]; !ok {
		na.history[nodeId] = make(AttackInTick)
	}
	at := na.history[nodeId]
	if _, ok := at[tick]; !ok {
		at[tick] = 1
	} else {
		at[tick] += 1
	}
}

func (na *NodesAttacks) IsAttacked(nodeId string) bool {
	_, ok := na.history[nodeId]
	return ok
}

func (na *NodesAttacks) LastAtack(nodeId string) int {
	at, ok := na.history[nodeId]
	if !ok {
		return 0
	}
	last := 0
	for t := range at {
		if t > last {
			last = t
		}
	}
	return last
}
