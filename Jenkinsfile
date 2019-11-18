import groovy.json.JsonSlurper

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
    stage('Deploy for development') {
      when {
        branch 'master'
      }
      steps {
        sh '''make build-prod-linux'''
        sh '''make build-jobs-prod-linux'''
        script {
          def remote_dev = [:]
          remote_dev.name = "wsodev"
          remote_dev.host = "wso-dev.williams.edu"
          remote_dev.port = 22
          remote_dev.allowAnyHosts = true

          withCredentials([usernamePassword(credentialsId: 'wsodev_ssh_server', passwordVariable: 'SSH_PASS', usernameVariable: 'SSH_USER')]) {
            remote_dev.user = SSH_USER
            remote_dev.password = SSH_PASS

            // Main executable:
            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/wso-backend'
            sshPut remote: remote_dev, from: 'wso-backend_linux', into: '/home/wsodev/wso-go/wso-backend'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/wso-backend'

            // Jobs:
            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-catalog-update'
            sshPut remote: remote_dev, from: 'job-catalog-update_linux', into: '/home/wsodev/wso-go/job-catalog-update'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-catalog-update'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-update-all-factrak-survey-deficits'
            sshPut remote: remote_dev, from: 'job-update-all-factrak-survey-deficits_linux', into: '/home/wsodev/wso-go/job-update-all-factrak-survey-deficits'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-update-all-factrak-survey-deficits'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-update-all-users-from-ldap'
            sshPut remote: remote_dev, from: 'job-update-all-users-from-ldap_linux', into: '/home/wsodev/wso-go/job-update-all-users-from-ldap'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-update-all-users-from-ldap'

            sshRemove remote: remote_dev, path: '/home/wsodev/wso-go/job-dorms-update'
            sshPut remote: remote_dev, from: 'job-dorms-update_linux', into: '/home/wsodev/wso-go/job-dorms-update'
            sshCommand remote: remote_dev, command: 'chmod +x /home/wsodev/wso-go/job-dorms-update'

            // Restart WSO-Go
            sshCommand remote: remote_dev, command: '/bin/systemctl restart WSO-Go', sudo: true
          }
        }
        script {
          try {
            URL apiUrl = new URL("https://wso-dev.williams.edu/api/v2/health-check")
            def resp = new JsonSlurper().parseText(apiUrl.getText())
            return resp.ok
          } catch (Exception e) {
            return false
          }
        }
      }
      post {
        success {
          slackSend (color: '#00FF00', message: "WSO-Go Deployed on Development\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
        }
      }
    }
    stage('Deploy for production') {
          when {
            branch 'production'
          }
          steps {
            sh '''make build-prod-linux'''
            sh '''make build-jobs-prod-linux'''
            script {
              def remote_dev = [:]
              remote_dev.name = "wso"
              remote_dev.host = "wso.williams.edu"
              remote_dev.port = 22
              remote_dev.allowAnyHosts = true

              withCredentials([usernamePassword(credentialsId: 'wso_ssh_server', passwordVariable: 'SSH_PASS', usernameVariable: 'SSH_USER')]) {
                remote_dev.user = SSH_USER
                remote_dev.password = SSH_PASS

                // Main executable:
                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/wso-backend'
                sshPut remote: remote_dev, from: 'wso-backend_linux', into: '/home/wso/wso/wso-backend/wso-backend'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/wso-backend'

                // Jobs:
                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/job-catalog-update'
                sshPut remote: remote_dev, from: 'job-catalog-update_linux', into: '/home/wso/wso/wso-backend/jobs/catalog-update'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/catalog-update'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/job-update-all-factrak-survey-deficits'
                sshPut remote: remote_dev, from: 'job-update-all-factrak-survey-deficits_linux', into: '/home/wso/wso/wso-backend/jobs/update-all-factrak-survey-deficits'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/update-all-factrak-survey-deficits'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/job-update-all-users-from-ldap'
                sshPut remote: remote_dev, from: 'job-update-all-users-from-ldap_linux', into: '/home/wso/wso/wso-backend/jobs/update-all-users-from-ldap'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/update-all-users-from-ldap'

                sshRemove remote: remote_dev, path: '/home/wso/wso/wso-backend/jobs/job-dorms-update'
                sshPut remote: remote_dev, from: 'job-dorms-update_linux', into: '/home/wso/wso/wso-backend/jobs/dorms-update'
                sshCommand remote: remote_dev, command: 'chmod +x /home/wso/wso/wso-backend/jobs/dorms-update'

                // Restart WSO-Go
                sshCommand remote: remote_dev, command: '/bin/systemctl restart WSO-Go', sudo: true
              }
            }
            script {
              try {
                URL apiUrl = new URL("https://wso.williams.edu/api/v2/health-check")
                def resp = new JsonSlurper().parseText(apiUrl.getText())
                return resp.ok
              } catch (Exception e) {
                return false
              }
            }
          }
          post {
            success {
              slackSend (color: '#00FF00', message: "WSO-Go Deployed on Production\n Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})")
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
