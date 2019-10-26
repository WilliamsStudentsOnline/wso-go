pipeline {
  agent {
    dockerfile {
      filename 'Dockerfile.builder'
      args '-u root:sudo'
    }

  }
  environment {
    CGO_ENABLED = 1
  }
  stages {
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
    stage('Deploy for wso-dev') {
      when {
        branch 'feature/continuous-deployment'
      }
      steps {
        sh '''make build-linux'''
        script {
          def remote_dev = [:]
          remote_dev.name = "wsodev"
          remote_dev.host = "wso-dev.williams.edu"
          remote_dev.port = 22
          remote_dev.allowAnyHosts = true

          withCredentials([usernamePassword(credentialsId: 'wsodev_ssh_server', passwordVariable: 'SSH_PASS', usernameVariable: 'SSH_USER')]) {
            remote_dev.user = SSH_USER
            remote_dev.password = SSH_PASS

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/wso-backend'
            sshPut remote: remote_dev, from: 'wso-backend_linux', into: '/home/wsodev/wso-go/wso-backend'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/wso-backend'
            sshCommand remote: remote_dev, command: 'sudo /bin/systemctl restart WSO-Go'
          }
        }
        script {
          try {
            new URL("https://wso-dev.williams.edu/api/v2/health-check").getText()
            return true
          } catch (Exception e) {
            return false
          }
        }
      }
      post {
        success {
          slackSend (color: '#00FF00', message: "SUCCESSFUL: Deployed on WSO-Dev.\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
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
