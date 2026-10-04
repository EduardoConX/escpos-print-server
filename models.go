package main

type Operation struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

type Body struct {
	Operations []Operation `json:"operations"`
	Printer    string      `json:"printer"`
}
