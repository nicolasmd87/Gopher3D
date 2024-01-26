package behaviour

type PlayerBehaviour interface {
	Start()
	Update()
}

type BehaviourWrapper struct {
	Behaviour PlayerBehaviour
	started   bool
}

type BehaviourManager struct {
	behaviours []BehaviourWrapper
}

var GlobalBehaviourManager = NewBehaviourManager()

func NewBehaviourManager() *BehaviourManager {
	return &BehaviourManager{}
}

func (m *BehaviourManager) Add(behaviour PlayerBehaviour) {
	m.behaviours = append(m.behaviours, BehaviourWrapper{Behaviour: behaviour, started: false})
}

func (m *BehaviourManager) UpdateAll() {
	for i := range m.behaviours {
		if !m.behaviours[i].started {
			m.behaviours[i].Behaviour.Start()
			m.behaviours[i].started = true
		}
		m.behaviours[i].Behaviour.Update()

	}
}
