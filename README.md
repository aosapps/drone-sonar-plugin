**Note**: It is necessary to use this updated fork of the Drone SonarQube Plugin to avoid the following warning message on the SonarQube server:

```
You are using Node.js version 10, which reached end-of-life. Support for this version will be dropped in future release, please upgrade Node.js to more recent version.
```
This version has the following updates:
* golang:1.15-alpine
* openjdk:11.0.13-jre
* Node.js v14


Links:
*  [Github](https://github.com/namig7/drone-sonar-plugin-nodejs14)
*  [Dockerhub](https://hub.docker.com/repository/docker/namigg/drone-sonar-plugin-nodejs14)


# drone-sonar-plugin
The plugin of Drone CI to integrate with SonarQube (previously called Sonar), which is an open source code quality management platform.

Detail tutorials: [DOCS.md](DOCS.md).

### Build process
build go binary file: 
`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o drone-sonar`

build docker image
`docker build -t aosapps/drone-sonar-plugin .`


### Testing the docker image:
```commandline
docker run --rm \
  -e DRONE_REPO=test \
  -e PLUGIN_SOURCES=. \
  -e SONAR_HOST=http://localhost:9000 \
  -e SONAR_TOKEN=60878847cea1a31d817f0deee3daa7868c431433 \
  aosapps/drone-sonar-plugin
```

### Pipeline example
```yaml
steps
- name: code-analysis
  image: aosapps/drone-sonar-plugin
  settings:
      sonar_host:
        from_secret: sonar_host
      sonar_token:
        from_secret: sonar_token
```
