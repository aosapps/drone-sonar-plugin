FROM golang:1.17-alpine as build

WORKDIR /go/src/github.com/aosapps/drone-sonar-plugin

COPY *.go /go/src/github.com/aosapps/drone-sonar-plugin/

RUN	go mod init github.com/aosapps/drone-sonar-plugin \
	&& go mod tidy \
	&& GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o drone-sonar

FROM openjdk:11.0.14.1-jre

ARG SONAR_VERSION=4.6.2.2472
ARG SONAR_SCANNER_CLI=sonar-scanner-cli-${SONAR_VERSION}
ARG SONAR_SCANNER=sonar-scanner-${SONAR_VERSION}

RUN apt-get update \
    && apt-get install -y curl \
    && apt-get clean

COPY --from=build /go/src/github.com/aosapps/drone-sonar-plugin/drone-sonar /bin/

WORKDIR /bin

RUN curl -fsSO https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/${SONAR_SCANNER_CLI}.zip \
    && unzip ${SONAR_SCANNER_CLI}.zip \
    && rm ${SONAR_SCANNER_CLI}.zip

ENV PATH $PATH:/bin/${SONAR_SCANNER}/bin

ENTRYPOINT /bin/drone-sonar
