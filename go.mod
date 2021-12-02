module github.com/grafana/grafana-starter-datasource-backend

go 1.16

require (
	github.com/NovatecConsulting/grafana-api-go-sdk v0.0.0-20211202150040-01d8090638f4
	github.com/go-git/go-git/v5 v5.4.2
	github.com/grafana-tools/sdk v0.0.0-20211118073920-e7b85bb25aa9 // indirect
	github.com/grafana/grafana-plugin-sdk-go v0.112.0
	github.com/sergi/go-diff v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
)

//replace github.com/grafana-tools/sdk => github.com/NovatecConsulting/grafana-api-go-sdk v0.0.0-20211118073920-e7b85bb25aa9
