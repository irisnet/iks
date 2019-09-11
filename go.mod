module github.com/irisnet/iks

go 1.12

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/btcsuite/btcd v0.0.0-20190115013929-ed77733ec07d // indirect
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.0
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/irisnet/irishub v0.15.1
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/client_golang v0.9.2 // indirect
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.0-20190328153300-af7bedc223fb // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.0.3
	github.com/stretchr/testify v1.2.2
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tendermint/tendermint v0.31.0
	golang.org/x/sys v0.0.0-20190329044733-9eb1bfa1ce65 // indirect
	google.golang.org/genproto v0.0.0-20190327125643-d831d65fe17d // indirect
	google.golang.org/grpc v1.19.1 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace (
	github.com/tendermint/iavl => github.com/irisnet/iavl v0.12.2
	github.com/tendermint/tendermint => github.com/irisnet/tendermint v0.31.0
	golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
)
