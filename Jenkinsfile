pipeline {
  agent {
    dockerfile {
      filename 'Dockerfile'
      args '-u root:sudo'
    }

  }
  stages {
    stage('Lint') {
      steps {
        sh '''wget -O - -q https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.17.1'''
        sh '''./bin/golangci-lint run --out-format checkstyle > golint-checkstyle.xml || true'''
      }
      post {
        always {
          recordIssues enabledForFailure: true, tools: [checkStyle(pattern: 'golint-checkstyle.xml')]
        }
      }
    }
    stage('Test') {
      steps {
        sh '''go get -u github.com/jstemmer/go-junit-report'''
        sh '''go get -u github.com/axw/gocov/gocov'''
        sh '''go get -u github.com/AlekSi/gocov-xml'''
        sh '''go test -v -coverprofile=c.out -race ./... 2>&1 | go-junit-report > report.xml'''
      }
      post {
        always {
          junit(testResults: 'report.xml', allowEmptyResults: true, healthScaleFactor: 1)
        }
        success {
          sh '''gocov convert c.out | gocov-xml > coverage.xml'''
          publishCoverage adapters: [coberturaAdapter('coverage.xml')], sourceFileResolver: sourceFiles('NEVER_STORE')
        }
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
