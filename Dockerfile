FROM golang:1.17-alpine as build

WORKDIR /go/src/github.com/aosapps/drone-sonar-plugin

COPY *.go /go/src/github.com/aosapps/drone-sonar-plugin/

RUN	go mod init github.com/aosapps/drone-sonar-plugin \
	&& go mod tidy \
	&& GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o drone-sonar

FROM sonarsource/sonar-scanner-cli:4.6

COPY --from=build /go/src/github.com/aosapps/drone-sonar-plugin/drone-sonar /bin/

WORKDIR /bin

RUN	curl -fsSO https://letsencrypt.org/certs/lets-encrypt-r3-cross-signed.pem \
	&& keytool -import -v -trustcacerts -cacerts -noprompt -storepass changeit \
		-file lets-encrypt-r3-cross-signed.pem \
		-alias "C=US, O=Let's Encrypt, CN=R3" \
	&& rm lets-encrypt-r3-cross-signed.pem

ENTRYPOINT /bin/drone-sonar
