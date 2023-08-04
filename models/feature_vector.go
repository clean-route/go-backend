package models

// This is the struct which contains all the relevant features (ITEMP, IRH, IWD, IWS, IPM, FTEMP, FRH, FWD, FWS)
type FeatureVector struct {
	ITEMP     float64 `json:"ITEMP"`
	IRH       float64 `json:"IRH"`
	IWD       float64 `json:"IWD"`
	IWS       float64 `json:"IWS"`
	IPM       float64 `json:"IPM"`
	FTEMP     float64 `json:"FTEMP"`
	FRH       float64 `json:"FRH"`
	FWD       float64 `json:"FWD"`
	FWS       float64 `json:"FWS"`
	DelayCode uint8   `json:"delayCode"`
}
