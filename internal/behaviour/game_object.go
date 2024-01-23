package behaviour

import "fmt"

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
	fmt.Println("Adding behaviour", behaviour)
	m.behaviours = append(m.behaviours, BehaviourWrapper{Behaviour: behaviour, started: false})
}

func (m *BehaviourManager) UpdateAll() {
	//	fmt.Print("Updating behaviours...", m.behaviours)
	for i := range m.behaviours {
		if !m.behaviours[i].started {
			m.behaviours[i].Behaviour.Start()
			m.behaviours[i].started = true
			fmt.Println("Started behaviour", m.behaviours[i].Behaviour)
		}
		m.behaviours[i].Behaviour.Update()

	}
}
