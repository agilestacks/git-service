#!/usr/bin/env groovy

@Library('agilestacks') _

import com.amazonaws.util.*
import com.amazonaws.auth.*
import com.amazonaws.services.ecr.*
import com.amazonaws.services.ecr.model.*
import com.amazonaws.regions.*

node('master') {
  stage('Checkout') {
    //checkout scm
    git credentialsId: 'github-user', url: 'https://github.com/agilestacks/git-service.git'
  }
}

podTemplate( inheritFrom: 'agilestacks',label: 'pod',
  containers: [
    containerTemplate(
      name: 'go',
      image: 'golang:1.9-alpine',
      ttyEnabled: true,
      command: 'cat')
  ],
  volumes: [emptyDirVolume(memory: false, mountPath: '/var/lib/docker')]
) {


  node('pod') {
    def imageTag1 = "${env.GIT_SERVICE_IMAGE}:${commitHash()}"
    def imageTag2 = "${env.GIT_SERVICE_IMAGE}:build-${env.BUILD_NUMBER}"

    dir('api') {
      stage('Compile') {
        container('go') {
          sh script: """
            GOPATH=\$(pwd)
            go get -u github.com/mitchellh/gox
            go get -u github.com/kardianos/govendor
          """, ctx: currentBuild
        }
        dir('src/gits') {
          container('go') {
            sh script: """
              GOPATH=\$(pwd)/../../
              \$GOPATH/bin/govendor sync
              go build -o \$GOBIN/linux/gits-musl -v gits
            """, ctx: currentBuild
          }
        }
      }

      stage('Build Container') {
        container('dind') {
          echo "Building image: ${imageTag1}"
          withEcr() {
            sh_ script: """
              docker build -t ${imageTag1} .
              docker tag ${imageTag1} ${imageTag2}
              docker push ${imageTag1}
              docker push ${imageTag2}
            """, ctx: currentBuild
          }
          echo "Pushed!"
        }
      }
    }
  }
}


