package main

//  patchStringValue specifies a patch operation for any object
type patchValue struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
