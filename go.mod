module github.com/irisnet/iks

go 1.15

require (
	github.com/cosmos/cosmos-sdk v0.40.0
	github.com/cosmos/go-bip39 v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/irisnet/irishub v1.0.0-beta.0.20210113085814-d5705f0d31c9
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tendermint v0.34.1
	github.com/tendermint/tm-db v0.6.3
	gopkg.in/yaml.v2 v2.4.0

)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
