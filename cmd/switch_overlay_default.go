//go:build !windows

package cmd

type switchOverlay struct{}

func startSwitchOverlay(string) *switchOverlay   { return &switchOverlay{} }
func startAddPreparationOverlay() *switchOverlay { return &switchOverlay{} }
func startAddSuccessOverlay() *switchOverlay     { return &switchOverlay{} }
func (*switchOverlay) Step(string, float64)      {}
func (*switchOverlay) Close()                    {}
