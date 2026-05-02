module github.com/woozymasta/lintkit

go 1.25.5

require (
	github.com/woozymasta/flags v0.3.2
	github.com/woozymasta/pathrules v0.1.2
	go.yaml.in/yaml/v3 v3.0.4
)

replace github.com/woozymasta/flags => ../../../go-flags

require (
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/term v0.42.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)
