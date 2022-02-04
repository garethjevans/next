TOKEN = $(shell yq e '."github.tools.sap".oauth_token' ~/.config/gh/hosts.yml)
GH_TOKEN = $(shell yq e '."github.com".oauth_token' ~/.config/gh/hosts.yml)

#DEBUG = "-debug"
DEBUG = ""
.PHONY: build
build:
	go build -o build/next main.go

run: build
	GITHUB_AUTH_TOKEN=$(TOKEN) ./build/next -host github.tools.sap -source-owner cki -source-repo dummy-repo $(DEBUG)
	GITHUB_AUTH_TOKEN=$(TOKEN) ./build/next -host github.tools.sap -source-owner cki -source-repo apache-tomcat $(DEBUG)
	GITHUB_AUTH_TOKEN=$(GH_TOKEN) ./build/next -source-owner paketo-buildpacks -source-repo spring-boot $(DEBUG)
