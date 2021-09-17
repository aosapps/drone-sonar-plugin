# Harness CIE SonarCube Plugin with Quality Gateway

The plugin of Harness CIE to integrate with SonarQube (previously called Sonar), which is an open source code quality management platform and check the report results for status OK.

![Plugin Configuration](https://github.com/diegopereiraeng/harness-cie-sonarqube-scanner/blob/master/Sonar-CIE.png)

![Results](https://github.com/diegopereiraeng/harness-cie-sonarqube-scanner/blob/master/SonarResult.png)

![Console Results](https://github.com/diegopereiraeng/harness-cie-sonarqube-scanner/blob/master/SonarResult2.png)

Detail tutorials: [DOCS.md](DOCS.md).

### Build process
build go binary file: 
`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o harness-sonar`

build docker image
`docker build -t diegokoala/harness-cie-sonarqube-scanner .`


### Testing the docker image:
```commandline
docker run --rm \
  -e DRONE_REPO=test \
  -e PLUGIN_SOURCES=. \
  -e SONAR_HOST=http://localhost:9000 \
  -e SONAR_TOKEN=60878847cea1a31d817f0deee3daa7868c431433 \
  diegokoala/harness-cie-sonarqube-scanner
```

### Pipeline example
```yaml
steps
- name: code-analysis
  image: diegokoala/harness-cie-sonarqube-scanner
  settings:
      sonar_host:
        from_secret: sonar_host
      sonar_token:
        from_secret: sonar_token
```
