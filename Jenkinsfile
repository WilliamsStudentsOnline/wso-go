pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
       sh 'go version'
       sh '''#!/bin/bash -l
GOOS=linux go build -a -mod vendor -tags=jsoniter -o wso-go main.go'''
      }
    }
    stage('Test') {
      steps {
        sh '''#!/bin/bash -l
go test -race .'''
      }
    }
  }
  options { buildDiscarder(logRotator(numToKeepStr: '2')) }
  post {
    cleanup {
      cleanWs()
    }
  }
}
