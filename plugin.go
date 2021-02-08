package main

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type (
	// Config struct
	Config struct {
		Key   string
		Name  string
		Host  string
		Token string

		Version         string
		Branch          string
		Sources         string
		Timeout         string
		Inclusions      string
		Exclusions      string
		Level           string
		ShowProfiling   string
		BranchAnalysis  bool
		UsingProperties bool
		TrustServerCert bool
	}
	// Plugin struct
	Plugin struct {
		Config Config
	}
)

// TrustServerCert : inject remote sonar server certificate in $JAVA_HOME/lib/security/cacerts
func (p Plugin) TrustServerCert() error {
	var f *os.File
	URIRegex := regexp.MustCompile(`^((?:ht|f)tp(?:s?)\:\/\/|~/|/)?([\w]+:\w+@)?([a-zA-Z]{1}([\w\-]+\.)+([\w]{2,5}))(:[\d]{1,5})?((/?\w+/)+|/?)(\w+\.[\w]{3,4})?((\?\w+=\w+)?(&\w+=\w+)*)?`)
	javaHome := os.Getenv("JAVA_HOME")
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	// get sonar host from provided URI
	host := URIRegex.FindStringSubmatch(p.Config.Host)[3]

	// connect to host and retrieve certificate chain
	fmt.Printf("\n/!\\ Trying to trust certificate for: %s\n\n", host)
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, 443), tlsConf)
	if err != nil {
		fmt.Printf("Error connecting to %s: %v\n", host, err)
		return err
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates

	// iter over chain to inject sonar cert into cacerts
	for _, cert := range certs {
		var namesToCheck []string
		hostFound := false
		namesToCheck = append(namesToCheck, cert.Subject.CommonName)
		for _, dnsname := range cert.DNSNames {
			namesToCheck = append(namesToCheck, dnsname)
		}
		for _, ipAddr := range cert.IPAddresses {
			namesToCheck = append(namesToCheck, fmt.Sprintf("%v", ipAddr))
		}
		for _, name := range namesToCheck {
			if name == host {
				hostFound = true
			}
		}
		if hostFound == false {
			continue
		}

		buf := bytes.NewBuffer([]byte{})
		err := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err != nil {
			fmt.Printf("Error parsing certificate: %v", err)
			return err
		}

		f, err = ioutil.TempFile("", "cert")
		if err != nil {
			fmt.Printf("Error opening temp file: %v", err)
			return err
		}

		defer os.Remove(f.Name())
		_, err = f.WriteString(fmt.Sprintf("%v", buf))
		if err != nil {
			fmt.Printf("Error writing temp file: %v", err)
			return err
		}
	}

	args := []string{
		"-importcert",
		"-alias",
		host,
		"-file",
		f.Name(),
		"-noprompt",
		"-keystore",
		fmt.Sprintf("%s/lib/security/cacerts", javaHome),
		"-storepass",
		"changeit",
	}
	cmd := exec.Command("keytool", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Exec : launch sonar analysis
func (p Plugin) Exec() error {
	args := []string{
		"-Dsonar.host.url=" + p.Config.Host,
		"-Dsonar.login=" + p.Config.Token,
	}

	if !p.Config.UsingProperties {
		argsParameter := []string{
			"-Dsonar.projectKey=" + strings.Replace(p.Config.Key, "/", ":", -1),
			"-Dsonar.projectName=" + p.Config.Name,
			"-Dsonar.projectVersion=" + p.Config.Version,
			"-Dsonar.sources=" + p.Config.Sources,
			"-Dsonar.ws.timeout=" + p.Config.Timeout,
			"-Dsonar.inclusions=" + p.Config.Inclusions,
			"-Dsonar.exclusions=" + p.Config.Exclusions,
			"-Dsonar.log.level=" + p.Config.Level,
			"-Dsonar.showProfiling=" + p.Config.ShowProfiling,
			"-Dsonar.scm.provider=git",
		}
		args = append(args, argsParameter...)
	}

	if p.Config.BranchAnalysis {
		args = append(args, "-Dsonar.branch.name="+p.Config.Branch)
	}

	cmd := exec.Command("sonar-scanner", args...)
	// fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("==> Code Analysis Result:\n")
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
