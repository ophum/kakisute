package controllertest

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Snapshot struct {
	Users  []*User `json:"users"`
	Cursor int     `json:"cursor"`
}

type Event struct {
	ID      int    `json:"id"`
	EventAt string `json:"event_at"`
	Action  string `json:"action"`
	Body    string `json:"body"`
}
