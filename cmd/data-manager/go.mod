module data-manager

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/PlatONEnetwork/PlatONE-Go v0.0.0-fe72c95c689da314dbca1c9a3707f1cc4874ffa6
	github.com/gin-gonic/gin v1.6.3
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.4.0
	go.mongodb.org/mongo-driver v1.4.1
)

replace github.com/PlatONEnetwork/PlatONE-Go => ../..
