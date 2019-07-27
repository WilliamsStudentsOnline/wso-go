pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        tool(name: 'Go 1.12.7', type: 'go')
        withEnv(["GOROOT=${root}", "PATH+GO=${root}/bin"]) {
          sh 'go version'
          sh '''#!/bin/bash -l
GOOS=linux go build -a -mod vendor -tags=jsoniter -o wso-go main.go'''
        }
        
      }
    }
    stage('Test') {
      steps {
        sh '''#!/bin/bash -l
go test -race .'''
      }
    }
  }
}
