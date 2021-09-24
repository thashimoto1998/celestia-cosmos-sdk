go 1.15

module github.com/cosmos/cosmos-sdk

require (
	github.com/99designs/keyring v1.1.6
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/adlio/schema v1.1.13 // indirect
	github.com/armon/go-metrics v0.3.4
	github.com/bgentry/speakeasy v0.1.0
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/celestiaorg/celestia-app v0.0.0-00010101000000-000000000000 // indirect
	github.com/celestiaorg/optimint v0.0.0-20210924100828-ef22978f2dd2 // indirect
	github.com/confio/ics23/go v0.6.3
	github.com/containerd/continuity v0.1.0 // indirect
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/iavl v0.15.3
	github.com/cosmos/ledger-cosmos-go v0.11.1
	github.com/enigmampc/btcutil v1.0.3-0.20200723161021-e2fb6adb2a25
	github.com/go-kit/kit v0.11.0 // indirect
	github.com/gogo/gateway v1.1.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/orderedcode v0.0.1 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/libp2p/go-libp2p v0.15.0 // indirect
	github.com/libp2p/go-libp2p-kad-dht v0.13.1 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.5.4 // indirect
	github.com/libp2p/go-libp2p-quic-transport v0.12.0 // indirect
	github.com/magiconair/properties v1.8.5
	github.com/mattn/go-isatty v0.0.13
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/otiai10/copy v1.2.0
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.30.0
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.23.0
	github.com/sasha-s/go-deadlock v0.2.1-0.20190427202633-1595213edefa // indirect
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.0
	github.com/tendermint/tm-db v0.6.4
	golang.org/x/crypto v0.0.0-20210813211128-0a44fdfbc16e
	golang.org/x/net v0.0.0-20210903162142-ad29c8ab022f // indirect
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
	google.golang.org/genproto v0.0.0-20210830153122-0bac4d21c8ea
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/celestiaorg/celestia-app => github.com/celestiaorg/lazyledger-app v0.0.0-20210909134530-18e69b513b3f
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
)
