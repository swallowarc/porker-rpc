package gateways

type (
	Factory interface {
		MemDBClient() MemDBClient
	}
)
